package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/dollarshaveclub/line"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()
var confFile string
var reportFile string
var debugFlag bool
var port int

func init() {
	flag.StringVar(&confFile, "conf", "", "yaml or json config file")
	flag.StringVar(&reportFile, "out", "", "report csv file")
	flag.IntVar(&port, "p", 0, "port to get statistic over HTTP")
	flag.BoolVar(&debugFlag, "d", false, "enables debug logs")
	flag.Parse()
}

func startRest(port int, task *Task) {
	http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(task)
	})

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func main() {
	if debugFlag == false {
		log.Level = logrus.FatalLevel
		log.Out = ioutil.Discard
	} else {
		log.Out = os.Stdout
		log.Formatter = new(logrus.TextFormatter)
		log.Level = logrus.DebugLevel
	}
	if confFile == "" {
		// TODO remove after all params from config would be in flags
		fmt.Println("Config file is required")
	}
	data, err := ioutil.ReadFile(confFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	task, err := NewTask(data)
	if err != nil {
		fmt.Println(err)
		return
	}
	line.Green("Working. Please wait...\n")
	if port > 0 {
		go startRest(9000, task)
	}
	results := task.Start()
	for i := 1; i <= task.Steps; i++ {
		stepStat := task.StepStats(i)
		if stepStat.Timeouts > 0 {
			line.Prefix("  ").Red().Println(stepStat)
		} else {
			line.Prefix("  ").White().Println(stepStat)
		}

	}
	stats := task.Stats()
	line.Green().Printf("Performed %d requests\n", stats.ResponsesGot)
	line.Green().Printf("Average request time per Task %f ms\n", stats.AvarageRequest.Seconds()*1000)
	line.Green().Printf("Longest request time per Task %f ms\n", stats.LongestRequest.Seconds()*1000)
	if stats.Timeouts > 0 {
		line.Red().Printf("Timeout were for %d requests", stats.Timeouts)
	} else {
		line.Green().Printf("No timeouts")
	}
	if reportFile != "" {
		err = CreateCSVReport(reportFile, results)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

}
