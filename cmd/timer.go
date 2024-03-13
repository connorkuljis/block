package cmd

// import (
// 	"fmt"
// 	"log"
// 	"sync"
// 	"time"

// 	"github.com/connorkuljis/block-cli/blocker"
// 	"github.com/connorkuljis/block-cli/interactive"
// 	"github.com/connorkuljis/block-cli/tasks"
// 	"github.com/spf13/cobra"
// )

// var timerCmd = &cobra.Command{
// 	Use: "timer",
// 	Run: func(cmd *cobra.Command, args []string) {
// 		fmt.Println("timer")
// 		// no args
// 		createdAt := time.Now()

// 		currentTask := tasks.InsertTask(tasks.NewTask("test", -1, true, false))

// 		timer(currentTask)

// 		finishedAt := time.Now()
// 		actualDuration := finishedAt.Sub(createdAt)

// 		tasks.UpdateCompletionPercent(currentTask, -1)

// 		// persist calculations
// 		if err := tasks.UpdateFinishTimeAndDuration(currentTask, finishedAt, actualDuration); err != nil {
// 			log.Fatal(err)
// 		}
// 	},
// }

// func timer(currentTask tasks.Task) {
// 	blocker := blocker.NewHostsBlocker()
// 	err := blocker.Start()
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	// initialise remote
// 	r := interactive.Remote{
// 		Task:    currentTask,
// 		Blocker: blocker,
// 		Wg:      &sync.WaitGroup{},
// 		Pause:   make(chan bool, 1),
// 		Cancel:  make(chan bool, 1),
// 		Finish:  make(chan bool, 1),
// 	}

// 	r.Wg.Add(2)
// 	go interactive.PollInput(r)
// 	go incrementer(r)
// 	r.Wg.Wait()

// 	err = blocker.Stop()
// 	if err != nil {
// 		log.Println(err)
// 	}
// }

// func incrementer(r interactive.Remote) {
// 	ticker := time.NewTicker(time.Second * 1)

// 	i := 0
// 	paused := false
// 	for {
// 		select {
// 		case <-r.Finish:
// 			r.Wg.Done()
// 			return
// 		case <-r.Cancel:
// 			r.Wg.Done()
// 			return
// 		case <-r.Pause:
// 			paused = !paused
// 		case <-ticker.C:
// 			if !paused {
// 				fmt.Printf("%d seconds\n", i)
// 				i++
// 			}
// 		}
// 	}
// }
