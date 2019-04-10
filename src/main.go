package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
)

func init() {
	logger = &logrus.Logger{
		Out:          os.Stderr,
		Formatter:    new(logrus.TextFormatter),
		Hooks:        make(logrus.LevelHooks),
		Level:        logrus.InfoLevel,
		ExitFunc:     os.Exit,
		ReportCaller: false,
	}
}

type Job struct {
	Args     []string
	N, Total int
}

var (
	quiet       bool
	workerCount int
	csvFile     string
	logger		*logrus.Logger
)

func producer(jobs [][]string) <-chan Job {
	ch := make(chan Job)
	go func() {
		for n, args := range jobs {
			ch <- Job{args, n, len(jobs)}
		}
		close(ch)
	}()
	return ch
}

func readJobStdout(id int, cmd string, job Job, reader *bufio.Reader) {
	for {
		line, err2 := reader.ReadString('\n')
		if err2 != nil {
			if err2 != io.EOF {
				logger.WithError(err2).Warnf("[%d] %d/%d, %s %s read result err ",
					id, job.N, job.Total, cmd, strings.Join(job.Args, " "))
			}
			break
		}
		logger.Infof("[%d] %d/%d, %s %s: %s",
			id, job.N, job.Total, cmd, strings.Join(job.Args, " "), line)
	}
}

func worker(id int, wg *sync.WaitGroup, cmd string, queue <-chan Job) {
	for job := range queue {
		if !quiet {
			logger.Infof("[%d] %d/%d: %s %s", id, job.N, job.Total, cmd, strings.Join(job.Args, " "))
		}
		wholeCmd := append([]string{cmd}, job.Args...)
		command := exec.Command("/bin/sh", wholeCmd...)

		stdout, e := command.StdoutPipe()
		if e != nil {
			logger.Warnf("[%d] %d/%d, %s %s can not obtain stdout pipe with: %s ",
				id, job.N, job.Total, cmd, strings.Join(job.Args, " "),e)
			return
		}
		stderr, e := command.StderrPipe()
		if e != nil {
			logger.Warnf("[%d] %d/%d, %s %s can not obtain stderr pipe with: %s ",
				id, job.N, job.Total, cmd, strings.Join(job.Args, " "),e)
			return
		}

		if e = command.Start(); e!=nil {
			logger.Warnf("[%d] %d/%d, %s %s execute err with: %s ",
				id, job.N, job.Total, cmd, strings.Join(job.Args, " "),e)
			return
		}

		reader := bufio.NewReader(stdout)
		go readJobStdout(id, cmd, job, reader)

		err_reader := bufio.NewReader(stderr)
		go readJobStdout(id, cmd, job, err_reader)

		if err := command.Wait(); err != nil {
			logger.Warnf("[%d] %d/%d, %s %s wait err with: %s ",
				id, job.N, job.Total, cmd, strings.Join(job.Args, " "),e)
			return
		}
	}
	wg.Done()
}

func main() {
	flag.BoolVar(&quiet, "q", false, "Quiet")
	flag.IntVar(&workerCount, "n", 8, "Number of `workers`")
	flag.StringVar(&csvFile, "csv", "", "CSV file containing arguments for each job")
	flag.Parse()

	if workerCount < 1 || (csvFile == "" && flag.NArg() < 2) || (csvFile != "" && flag.NArg() < 1) {
		fmt.Println("Usage:")
		fmt.Println(os.Args[0], "[-q] [-n <workers>] <cmd> <job1> [<job2> ...]")
		fmt.Println(os.Args[0], "[-q] [-n <workers>] -csv <csv-file> <cmd>")
		return
	}

	var jobs [][]string
	cmd := flag.Arg(0)
	if csvFile == "" {
		for _, arg := range flag.Args()[1:] {
			jobs = append(jobs, strings.Split(arg, ","))
		}
	} else {
		var f io.Reader
		f, err := os.Open(csvFile)
		if err != nil {
			logger.WithError(err).WithField("fileName", csvFile).Errorln("open csv file failed.")
			return
		}
		jobs, err = csv.NewReader(f).ReadAll()
		if err != nil && err != io.EOF {
			logger.Errorf("Failed to parse", csvFile, err)
			return
		}
	}

	queue := producer(jobs)

	var wg sync.WaitGroup
	wg.Add(workerCount)
	for id := 1; id <= workerCount; id++ {
		go worker(id, &wg, cmd, queue)
	}
	wg.Wait()
}
