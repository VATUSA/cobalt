package cli

type CommandFunc func(args []string) error

var Commands = map[string]CommandFunc{
	"run":        RunBackgroundJob,
	"run_inline": RunBackgroundJobInline,
}
