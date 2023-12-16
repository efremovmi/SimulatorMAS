package kernel

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"

	"SimulatorMAS/internal/agent/greens"
	. "SimulatorMAS/internal/broker"
)

func (k *kernel) CreateAndRunGreens(ctx context.Context, wg *sync.WaitGroup) {
	k.mu.Lock()
	defer k.mu.Unlock()

	g := greens.NewGreens(ctx, uuid.New(), wg, k)
	k.greensTable[g.GetPID()] = g

	wg.Add(1)
	g.Run(wg)

	return
}

func (k *kernel) WaitingNewGreens() {
	greensReqToBornChan := Broker.GetGreensReqToBornChan()
	go func() {
		for _ = range greensReqToBornChan {
			k.CreateAndRunGreens(k.appCtx, k.wg)
			k.NumberCorpseDec()
		}
	}()
}

func (k *kernel) IsReproductionPossible() (string, bool) {
	k.mu.Lock()
	defer k.mu.Unlock()

	if len(k.rodentTable) >= 2 {
		return k.eatGreensIfExist()
	}

	return "", false
}

func (k *kernel) CheckGreensByPID(pid uint32) bool {
	k.mu.Lock()
	defer k.mu.Unlock()

	if _, ok := k.greensTable[pid]; ok {
		return true
	}

	return false
}

func (k *kernel) DeleteGreensByPID(pid uint32) {
	k.mu.Lock()
	defer k.mu.Unlock()

	_, ok := k.greensTable[pid]
	if !ok {
		return
	}

	delete(k.greensTable, pid)
}

func (k *kernel) deleteGreensByPID(pid uint32) {
	_, ok := k.greensTable[pid]
	if !ok {
		return
	}

	delete(k.greensTable, pid)
}

func (k *kernel) eatGreensIfExist() (string, bool) {
	// restriction on the minimum amount of greenery
	if len(k.greensTable) < 3 {
		return "", false
	}

	for _, v := range k.greensTable {
		if v.IsAlive() {
			v.Kill()
			k.deleteGreensByPID(v.GetPID())
			return fmt.Sprintf("%d", v.GetPID()), true
		}
	}

	return "", false
}
