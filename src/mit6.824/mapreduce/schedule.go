package mapreduce

import (
	"fmt"
	"sync"
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

	// All ntasks tasks have to be scheduled on workers. Once all tasks
	// have completed successfully, schedule() should return.
	//
	// Your code here (Part III, Part IV).
	//
	var wg sync.WaitGroup
	wg.Add(ntasks)

	taskChan := make(chan int, 3)
	go func() {
		for i := 0; i < ntasks; i++ {
			taskChan <- i
		}
	}()

	go func() {
		for {
			availableWorker := <-registerChan
			task := <-taskChan

			doTaskArgs := DoTaskArgs{
				JobName:       jobName,
				File:          mapFiles[task],
				Phase:         phase,
				TaskNumber:    task,
				NumOtherPhase: n_other,
			}

			go func() {
				if call(availableWorker, "Worker.DoTask", doTaskArgs, new(struct{})) {
					// 任务成功
					wg.Done()
					// 任务结束后，availableWorker 闲置，availableWorker 进入 registerChan 等待下一次分配任务
					registerChan <- availableWorker
				} else {
					// 任务失败, 重新提交回任务, 等待下一次调度分配
					taskChan <- doTaskArgs.TaskNumber
				}
			}()
		}
	}()

	wg.Wait()
	fmt.Printf("Schedule: %v done\n", phase)
}
