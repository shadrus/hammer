package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

//RequestStatistic result of a BenchRequest
type RequestStatistic struct {
	Duration   time.Duration
	StatusCode int
	Step       int
}

func (rs *RequestStatistic) String() string {
	return fmt.Sprintf("Request duration: %f. Status code %d", rs.Duration.Seconds()*1000, rs.StatusCode)
}

// Statistic ...
type Statistic struct {
	ResponsesGot   int
	Timeouts       int
	LongestRequest time.Duration
	AvarageRequest time.Duration
}

//StepStatistic stat for one step
type StepStatistic struct {
	Statistic
	RPS  int
	Step int
}

func (ss StepStatistic) String() string {
	return fmt.Sprintf("Step #%d. RPS: %d. Average: %f. Longest: %f, Timeouts: %d", ss.Step, ss.RPS, ss.AvarageRequest.Seconds()*1000, ss.LongestRequest.Seconds()*1000, ss.Timeouts)
}

// BenchRequest request with measures
type BenchRequest struct {
	*http.Request
	RequestTimeout int
}

// MakeRequest performs request
func (r *BenchRequest) MakeRequest(step int, wg *sync.WaitGroup, resChan chan<- RequestStatistic) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.RequestTimeout)*time.Second)
	defer cancel()
	defer wg.Done()
	go func() {
		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				resChan <- RequestStatistic{time.Duration(r.RequestTimeout*1000) * time.Millisecond, 408, step}
			}
		}
	}()

	req := r.WithContext(ctx)
	startTime := time.Now()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error(err)
		return
	}
	defer resp.Body.Close()
	reqDuration := time.Since(startTime)
	resChan <- RequestStatistic{reqDuration, resp.StatusCode, step}
}

// NewBenchRequest factory for new BenchRequest
func NewBenchRequest(task *Task) (*BenchRequest, error) {
	request, err := http.NewRequest(task.Method, task.URL, strings.NewReader(task.Body))
	if err != nil {
		return nil, err
	}
	q := request.URL.Query()
	for key, value := range task.Params {
		q.Add(key, value)
	}
	request.URL.RawQuery = q.Encode()
	h := request.Header
	for key, value := range task.Headers {
		h.Add(key, value)
	}
	for key, value := range task.BasicAuth {
		if key != "" && value != "" {
			request.SetBasicAuth(key, value)
		}
	}
	req := &BenchRequest{request, task.Timeout}
	return req, nil
}

// Task settings
type Task struct {
	URL       string
	Method    string
	Params    map[string]string
	Headers   map[string]string
	Timeout   int
	Steps     int
	MaxRPS    int     `yaml:"maxRPS"`
	GrowsCoef float64 `yaml:"growsCoef"`
	Results   []*RequestStatistic
	Body      string
	BasicAuth map[string]string
}

func (t *Task) findGrowthCoef() float64 {
	return float64(t.MaxRPS) / (float64(t.Steps) * t.GrowsCoef)
}

func (t *Task) getCurrentRPS(step int) int {
	rps := int(t.findGrowthCoef() * float64(step))
	if rps > t.MaxRPS {
		return t.MaxRPS
	}
	return rps
}

// Start performing task
func (t *Task) Start() []*RequestStatistic {
	t.Results = make([]*RequestStatistic, 0)
	var wg sync.WaitGroup
	var resChan = make(chan RequestStatistic)
	defer close(resChan)
	go func(c <-chan RequestStatistic) {
		for {
			stat := <-c
			// Hack why there is empty responses?
			if stat.Duration > 0 {
				log.WithFields(logrus.Fields{
					"request time":    stat.Duration,
					"response status": stat.StatusCode,
					"step":            stat.Step,
				}).Debug("Got response")
				t.Results = append(t.Results, &stat)
			}

		}

	}(resChan)
	limiter := time.Tick(1 * time.Second)
	for step := 1; step <= t.Steps; step++ {
		<-limiter
		log.Infof("Performing step %d", step)
		for i := 0; i < t.getCurrentRPS(step); i++ {
			request, err := NewBenchRequest(t)
			if err != nil {
				log.Error(err)
				return nil
			}
			wg.Add(1)
			go request.MakeRequest(step, &wg, resChan)
		}
		log.Debug("waiting to complete step")
		wg.Wait()
	}
	// BUG in debug panic: send on closed channel
	log.Debug("Task time elapsed")
	time.Sleep(1 * time.Second)
	return t.Results
}

// StepStats statistics for the step
func (t *Task) StepStats(step int) StepStatistic {
	var longest time.Duration
	var avarage time.Duration
	responsesCount := 0
	timeouts := 0
	for idx, value := range t.Results {
		if value.Step == step {
			responsesCount++
			if value.StatusCode == 408 {
				timeouts++
			}
			if idx == 0 || t.Results[idx-1].Step < step {
				longest = value.Duration
				avarage = value.Duration
				continue
			}
			if t.Results[idx-1].Step == step {
				if value.Duration > longest {
					longest = value.Duration
				}
				avarage = (avarage + value.Duration) / 2
			}
		}
		if value.Step > step {
			break
		}
	}
	return StepStatistic{Statistic{ResponsesGot: responsesCount, Timeouts: timeouts, LongestRequest: longest, AvarageRequest: avarage}, t.getCurrentRPS(step), step}
}

// Stats prints Task statisctic
func (t *Task) Stats() Statistic {
	var stats Statistic
	for i := 1; i <= t.Steps; i++ {
		stepStat := t.StepStats(i)
		if stepStat.LongestRequest > stats.LongestRequest {
			stats.LongestRequest = stepStat.LongestRequest
		}
		if i > 1 {
			stats.AvarageRequest = (stepStat.AvarageRequest + stats.AvarageRequest) / 2
		}
		stats.Timeouts = stats.Timeouts + stepStat.Timeouts
		stats.ResponsesGot = stats.ResponsesGot + stepStat.ResponsesGot
	}
	return stats
}

func selectUnmarshaller(data []byte) func(in []byte, out interface{}) (err error) {
	if json.Valid(data) {
		return json.Unmarshal
	}
	return yaml.Unmarshal
}

// NewTask return Task instance based in config data
func NewTask(data []byte) (*Task, error) {
	task := Task{}
	unmarshallFunc := selectUnmarshaller(data)
	err := unmarshallFunc(data, &task)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

//CreateCSVReport create csv file base on requests stats
func CreateCSVReport(filename string, stats []*RequestStatistic) error {
	file, err := os.Create(filename)
	checkError("Cannot create file", err)
	defer file.Close()
	w := csv.NewWriter(file)
	header := []string{"step", "duration", "status"}
	if err := w.Write(header); err != nil {
		return err
	}
	for _, record := range stats {
		line := []string{strconv.Itoa(record.Step), strconv.Itoa(int(record.Duration.Seconds() * 1000)), strconv.Itoa(record.StatusCode)}
		if err := w.Write(line); err != nil {
			return err
		}
	}

	// Write any buffered data to the underlying writer (standard output).
	w.Flush()

	if err := w.Error(); err != nil {
		return err
	}
	return nil
}

func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}
