package main

import (
	"log"
	"sync"

	"github.com/jmoiron/sqlx"
)

var (
	db          *sqlx.DB
	currentTask *Task
	cfg         Config
	flags       Flags
)

type Flags struct {
	DisableBlocker bool
	ScreenRecorder bool
	Verbose        bool
}

type Remote struct {
	Pause  chan bool
	Cancel chan bool
	Finish chan bool
	wg     *sync.WaitGroup
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&flags.DisableBlocker, "no-block", "d", false, "Do not block hosts file.")
	rootCmd.PersistentFlags().BoolVarP(&flags.ScreenRecorder, "screen-recorder", "x", false, "Enable screen recorder.")
	rootCmd.PersistentFlags().BoolVarP(&flags.Verbose, "verbose", "v", false, "Logs additional details.")
}

func main() {
	var err error

	cfg, err = InitConfig()
	if err != nil {
		log.Fatal(err)
	}

	db, err = InitDB()
	if err != nil {
		log.Fatal(err)
	}

	rootCmd.AddCommand(startCmd, historyCmd, deleteTaskCmd, resetDNSCmd)

	err = rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}

	db.Close()
}

func start() {
	r := Remote{
		Pause:  make(chan bool, 1),
		Cancel: make(chan bool, 1),
		Finish: make(chan bool, 1),
		wg:     &sync.WaitGroup{},
	}

	r.wg.Add(2)
	go RenderProgressBar(r)
	go PollInput(r)

	if flags.ScreenRecorder {
		r.wg.Add(1)
		go FfmpegCaptureScreen(r)
	}

	r.wg.Wait()
}
