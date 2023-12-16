package types

import (
	"sync"
)

type IGreens interface {
	Run(wg *sync.WaitGroup)
	Kill()
	GetPID() uint32
	IsAlive() bool
}
