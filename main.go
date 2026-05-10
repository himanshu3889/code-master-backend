package main

import (
	"github.com/himanshu3889/code-master-backend/cmd/server"
	"github.com/himanshu3889/code-master-backend/codeRunner"
)

func main() {
	// Start the code runner
	codeRunner.LoadEnv()
	codeRunner.StartCodeExecutionJobQueue(2)

	defer codeRunner.StopCodeExecutionJobQueue()
	defer codeRunner.ShutdownCodeRunners()

	server.RunServer()
}
