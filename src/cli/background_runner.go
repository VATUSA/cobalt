package cli

import (
	"errors"
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
