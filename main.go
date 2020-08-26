package main

import (
	"fmt"
	"log"
	"time"
)

func main() {

	fmt.Printf("\nScript started at: %s\n\n", time.Now().String())

	workerHandler := WorkerHandler{
		StartWithPort:      3001,
		NumGenerateNumbers: 100,
		// this interval should be adjusted according to `numDataSamples` & `numWorkers`
		// obviously launching more workers results in delayed process launch, hence increased `PauseInterval` value
		PauseInterval: 5 * time.Millisecond,
		MaxRetry:      10,
	}

	result := workerHandler.Handle(150, 16)

	for k, v := range result.DataSamples {
		println(fmt.Sprintf("%d. %d", k, v))
	}

	log.Printf("Total numbers: %d", len(result.DataSamples))
	log.Printf("Total time spent: %s", result.TotalTimeSpent.String())

	fmt.Printf("Script finished at: %s", time.Now().String())
}
