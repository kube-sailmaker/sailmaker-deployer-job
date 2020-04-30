package main

import (
	"github.com/kube-sailmaker/sailmaker-deployer-job/job"
	"github.com/kube-sailmaker/sailmaker-deployer-job/opts"
	"os"
)

func main() {
	flags := opts.FlagSets()
	if len(os.Args) < 2 {
		opts.ExitWithUsage("Command Missing", flags["deploy"])
	}
	switch os.Args[1] {
	case "deploy":
		job.Process(opts.ProcessDeployment(flags["deploy"]))
	default:
		opts.ExitWithUsage("Unknown command", flags["deploy"])
	}
}
