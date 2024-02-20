package interactive

import (
	"sync"

	"github.com/connorkuljis/block-cli/blocker"
	"github.com/connorkuljis/block-cli/tasks"
)

type Remote struct {
	Task    tasks.Task
	Blocker blocker.HostsBlocker

	Wg     *sync.WaitGroup
	Pause  chan bool
	Cancel chan bool
	Finish chan bool
}
