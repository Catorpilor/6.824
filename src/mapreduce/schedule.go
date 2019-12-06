package mapreduce

import (
	"fmt"
)

//
// schedule() starts and waits for all tasks in the given phase (mapPhase
// or reducePhase). the mapFiles argument holds the names of the files that
// are the inputs to the map phase, one per map task. nReduce is the
// number of reduce tasks. the registerChan argument yields a stream
// of registered workers; each item is the worker's RPC address,
// suitable for passing to call(). registerChan will yield all
// existing registered workers (if any) and new ones as they register.
//
func schedule(jobName string, mapFiles []string, nReduce int, phase jobPhase, registerChan chan string) {
	var ntasks int
	var n_other int // number of inputs (for reduce) or outputs (for map)
	switch phase {
	case mapPhase:
		ntasks = len(mapFiles)
		n_other = nReduce
	case reducePhase:
		ntasks = nReduce
		n_other = len(mapFiles)
	}

	fmt.Printf("Schedule: %v %v tasks (%d I/Os)\n", ntasks, phase, n_other)
	// ctasks behaves like a task queue
	ctasks := make(chan int)
	go func() {
		for i := 0; i < ntasks; i++ {
			ctasks <- i
		}
	}()
	successTasks := 0
	success := make(chan int)
	for {
		select {
		case idx := <-ctasks:
			go func() {
				w := <-registerChan
				args := DoTaskArgs{JobName: jobName, Phase: phase, TaskNumber: idx, NumOtherPhase: n_other}
				if phase == mapPhase {
					args.File = mapFiles[idx]
				}
				if call(w, "Worker.DoTask", args, nil) {
					go func() { registerChan <- w }()
					success <- 1
				} else {
					ctasks <- idx
				}

			}()
		case <-success:
			successTasks++
		default:
			if successTasks == ntasks {
				goto END
			}
		}
	}

	// All ntasks tasks have to be scheduled on workers. Once all tasks
	// have completed successfully, schedule() should return.
	//
	// Your code here (Part III, Part IV).
	//
END:
	fmt.Printf("Schedule: %v done\n", phase)
}
