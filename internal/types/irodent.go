package types

import "sync"

type IRodent interface {
	Run(wg *sync.WaitGroup)
	Kill()
	GetPID() uint32
	IsAlive() bool
}
