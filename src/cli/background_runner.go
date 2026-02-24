package cli

import (
	"errors"
	"fmt"
	"log"
	"vatusa-cobalt/background"
)

func RunBackgroundJob(args []string) error {
	if len(args) < 1 {
		return errors.New("must provide a job name")
	}
	jobName := args[0]
	jobArgs := []string{}
	if len(args) > 1 {
		jobArgs = args[1:]
	}

	job := background.NewJob(jobName, jobArgs...)
	err := job.Run()
	if err != nil {
		return err
	}
	return nil
}

func RunBackgroundJobInline(args []string) error {
	if len(args) < 1 {
		return errors.New("must provide a job name")
	}
	jobName := args[0]
	jobArgs := []string{}
	if len(args) > 1 {
		jobArgs = args[1:]
	}

	jobFunc, ok := background.JobDefinitions[jobName]
	if !ok {
		return errors.New(fmt.Sprintf("job '%s' not found", jobName))
	}

	err := jobFunc(jobArgs)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}
