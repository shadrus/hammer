package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/dollarshaveclub/line"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()
var confFile string
var reportFile string
var debugFlag bool

func init() {
	flag.StringVar(&confFile, "conf", "", "yaml or json config file")
	flag.StringVar(&reportFile, "out", "", "report csv file")
	flag.BoolVar(&debugFlag, "d", false, "enables debug logs")
	flag.Parse()
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
