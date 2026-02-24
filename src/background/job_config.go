package background

type JobFunc func(args []string) error

var JobDefinitions = map[string]JobFunc{
	"test": TestJob,
}
