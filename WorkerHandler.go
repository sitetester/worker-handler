package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"time"
)

type WorkerHandler struct {
	StartWithPort      int
	NumGenerateNumbers int

	// this should be a reasonable value, not too short, not v long
	PauseInterval time.Duration
	MaxRetry      int
}

type WorkerIdWithPort struct {
	WorkerId int
	Port     int
}

type Result struct {
	DataSamples    []int
	TotalTimeSpent time.Duration
}

func (handler WorkerHandler) Handle(numDataSamples int, numWorkers int) Result {

	start := time.Now()
	dataSamplesChan := make(chan int)
	var dataSamples []int

	var workerIdWithPort WorkerIdWithPort
	port := handler.StartWithPort

	for i := 1; i <= numWorkers; i++ {
		workerIdWithPort = WorkerIdWithPort{
			Port: port, WorkerId: i,
		}
		port += 1

		go launchWorker(workerIdWithPort)
		handler.pause()

		go handler.collectDataSamples(workerIdWithPort, dataSamplesChan, 1)
	}

forLoop:
	for {
		select {

		case dataSample := <-dataSamplesChan:
			dataSamples = append(dataSamples, dataSample)

			if len(dataSamples) == numDataSamples {
				break forLoop
			}
		}
	}

	return Result{
		DataSamples:    dataSamples,
		TotalTimeSpent: time.Since(start),
	}
}

// Launches worker binary with configured port & workerId
// Example: ./bin/worker.mac -workerId 1 -port 3001
func launchWorker(workerIdWithPort WorkerIdWithPort) {

	port := fmt.Sprintf("%s%d", "-port=", workerIdWithPort.Port)
	workerId := fmt.Sprintf("%s%d", "-workerId=", workerIdWithPort.WorkerId)

	log.Println(fmt.Sprintf("Launching worker with params %s %s", port, workerId))

	err := exec.Command(getOsDependentWorker(), port, workerId).Start()
	if err != nil {
		log.Fatal(err)
	}
}

func getOsDependentWorker() string {

	goos := runtime.GOOS

	if goos == "darwin" {
		return "./bin/worker.mac"

	} else if goos == "linux" {
		return "./bin/worker.linux"

	} else if goos == "windows" {
		return "./bin/worker.windows"
	}

	log.Fatal("Unsupported OS!")
	return ""
}

// Parses random number from chunked HTTP response
func (handler WorkerHandler) collectDataSamples(workerIdWithPort WorkerIdWithPort, dataSamplesChan chan int, retryCount int) {

	url := fmt.Sprintf("http://localhost:%d/rnd?n=%d", workerIdWithPort.Port, handler.NumGenerateNumbers)
	resp, err := http.Get(url)
	if err != nil {
		// it could be that process has still not launched yet (in case of short pause interval)
		handler.pause()

		if retryCount > handler.MaxRetry {
			log.Fatalf("Worker not responding for URL: %s, Error: %s", url, err.Error())
		}

		// retry HTTP after configured pause interval
		handler.collectDataSamples(workerIdWithPort, dataSamplesChan, retryCount+1)
	}

	re := regexp.MustCompile(`rnd=(\d+)(\r)?\n`)
	buff := make([]byte, 20)

	for {
		n, err := resp.Body.Read(buff)
		if err != nil {
			log.Printf("Worker crashed for URL: %s, %s\n", url, err.Error())

			go launchWorker(workerIdWithPort)
			handler.pause()
			handler.collectDataSamples(workerIdWithPort, dataSamplesChan, 1)
		}

		if n > 0 {
			line := string(buff[:n])
			matches := re.FindStringSubmatch(line)
			if len(matches) > 0 {
				dataSample, err := strconv.Atoi(matches[1])
				if err != nil {
					log.Fatalf("Invalid input: %s, %s", err.Error(), line)
				}

				dataSamplesChan <- dataSample
			} else {
				log.Printf("Unexpected line: %s\n", line)
			}
		}
	}
}

// Wait for process to launch, otherwise it throws `tcp: connection refused` for `GET` request
func (handler WorkerHandler) pause() {
	time.Sleep(handler.PauseInterval)
}
