[![Build Status](https://travis-ci.org/jrallison/go-workers.png)](https://travis-ci.org/jrallison/go-workers)
[![GoDoc](https://godoc.org/github.com/jrallison/go-workers?status.png)](https://godoc.org/github.com/jrallison/go-workers)

[Sidekiq](http://sidekiq.org/) compatible
background workers in [golang](http://golang.org/).

* reliable queueing for all queues using [brpoplpush](http://redis.io/commands/brpoplpush)
* handles retries
* support custom middleware
* customize concurrency per queue
* responds to Unix signals to safely wait for jobs to finish before exiting.
* provides stats on what jobs are currently running
* well tested

Example usage:

```go
package main

import (
	"github.com/topfreegames/go-workers"
	workers "go-workers"
)

func myJob(message *workers.Msg) {
	// do something with your message
	// message.Jid()
	// message.Args() is a wrapper around go-simplejson (http://godoc.org/github.com/bitly/go-simplejson)
}

type myMiddleware struct{}

func (r *myMiddleware) Call(queue string, message *workers.Msg, next func() bool) (acknowledge bool) {
	// do something before each message is processed
	acknowledge = next()
	// do something after each message is processed
	return
}

func main() {
	workers.Configure(workers.Options{
		// location of redis instance
		Address: "localhost:6379",
		// instance of the database
		Database: "0",
		// number of connections to keep open with redis
		PoolSize: "30",
		// unique process id for this instance of workers (for proper recovery of inprogress jobs on crash)
		ProcessID: "1",
	})

	workers.Middleware.Append(&myMiddleware{})

	// pull messages from "myqueue" with concurrency of 10
	workers.Process("myqueue", myJob, 10)

	// pull messages from "myqueue2" with concurrency of 20
	workers.Process("myqueue2", myJob, 20)

	// Add a job to a queue
	workers.Enqueue("myqueue3", "Add", []int{1, 2})

	// Add a job to a queue with retry
	workers.EnqueueWithOptions("myqueue3", "Add", []int{1, 2}, workers.EnqueueOptions{Retry: true})

	// Add a job to a queue in a different redis instance
	workers.EnqueueWithOptions("myqueue4", "Add", []int{1, 2},
		workers.EnqueueOptions{
			Retry: true,
			ConnectionOptions: workers.Options{
				Address:  "localhost:6378",
				Database: "my-database",
				PoolSize: 10,
				Password: "pass",
			},
		},
	)

	// stats will be available at http://localhost:8080/stats
	go workers.StatsServer(8080)

	// Blocks until process is told to exit via unix signal
	workers.Run()
}
```

Initial development sponsored by [Customer.io](http://customer.io)
