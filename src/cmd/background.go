package main

import (
	"fmt"
	"log"
	"os"
	"vatusa-cobalt/background"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Job name not specified")
		return
	}
	jobName := os.Args[1]
	var jobArgs []string
	if len(os.Args) > 2 {
		jobArgs = os.Args[2:]
	}
	log.Printf("Starting background job %s\n", jobName)
	log.Printf("Args: %v\n", jobArgs)

	jobFunc, ok := background.JobDefinitions[jobName]
	if !ok {
		log.Fatalf("Job definition %s not found\n", jobName)
	}

	err := jobFunc(jobArgs)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Finished background job %s\n", jobName)
}
