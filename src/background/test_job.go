package background

import "log"

func TestJob(args []string) error {
	log.Printf("This is a test job. It doesn't actually do anything except print this message.\n")
	return nil
}
