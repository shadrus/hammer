package main

import (
	"flag"
	"io/ioutil"
	"os"

	"github.com/dollarshaveclub/line"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()
var confFile string
var reportFile string

func init() {
	flag.StringVar(&confFile, "conf", "", "yaml config file")
	flag.StringVar(&reportFile, "out", "", "report csv file")
	flag.Parse()
}

func main() {
	log.Out = os.Stdout
	log.Formatter = new(logrus.TextFormatter)
	log.Level = logrus.ErrorLevel
	data, err := ioutil.ReadFile(confFile)
	if err != nil {
		log.Error(err)
		return
	}
	task, err := NewTask(data)
	if err != nil {
		log.Error(err)
		return
	}
	line.Blue("Working. Please wait...\n")
	results := task.Start()
	for i := 1; i <= task.Durability; i++ {
		stepStat := task.StepStats(i)
		line.Prefix("  ").White().Println(stepStat)
	}
	stats := task.Stats()
	line.Blue().Printf("Performed %d requests\n", stats.ResponsesGot)
	line.Blue().Printf("Average request time per Task %f ms\n", stats.AvarageRequest.Seconds()*1000)
	line.Blue().Printf("Longest request time per Task %f ms\n", stats.LongestRequest.Seconds()*1000)
	line.Red().Printf("Timeout were for %d requests", stats.Timeouts)
	if reportFile != "" {
		err = CreateCSVReport(reportFile, results)
		if err != nil {
			log.Error(err)
			return
		}
	}

}
