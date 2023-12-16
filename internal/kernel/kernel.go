package kernel

import (
	"context"
	"sync"
	"time"

	. "SimulatorMAS/internal/logger"
	"SimulatorMAS/internal/types"
)

var Kernel *kernel

type kernel struct {
	mu     sync.RWMutex
	appCtx context.Context
	wg     *sync.WaitGroup

	NumberCorpse struct {
		Count int
		mu    sync.RWMutex
	}

	greensTable map[uint32]types.IGreens

	rodentTable map[uint32]types.IRodent
}

func NewKernel(ctx context.Context, wg *sync.WaitGroup) *kernel {
	return &kernel{
		appCtx: ctx,
		wg:     wg,
		NumberCorpse: struct {
			Count int
			mu    sync.RWMutex
		}{Count: 0},
		greensTable: make(map[uint32]types.IGreens),
		rodentTable: make(map[uint32]types.IRodent),
	}
}

func (k *kernel) CollectingMetrics() {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-ticker.C:
				KernelInfo("Kernel: Greens count %d, Rodent count %d, Corpse count %d",
					len(k.greensTable), len(k.rodentTable), k.GetNumberCorpse())
			}
		}
	}()
}

func (k *kernel) NumberCorpseInc() {
	k.NumberCorpse.mu.Lock()
	defer k.NumberCorpse.mu.Unlock()
	k.NumberCorpse.Count += 1
}

func (k *kernel) NumberCorpseAdd(count int) {
	k.NumberCorpse.mu.Lock()
	defer k.NumberCorpse.mu.Unlock()
	k.NumberCorpse.Count += count
}

func (k *kernel) NumberCorpseDec() {
	k.NumberCorpse.mu.Lock()
	defer k.NumberCorpse.mu.Unlock()
	k.NumberCorpse.Count -= 1
}

func (k *kernel) GetNumberCorpse() int {
	k.NumberCorpse.mu.Lock()
	defer k.NumberCorpse.mu.Unlock()
	return k.NumberCorpse.Count
}
