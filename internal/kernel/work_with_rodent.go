package kernel

import (
	"context"
	"sync"

	"github.com/google/uuid"

	"SimulatorMAS/internal/agent/rodent"
	. "SimulatorMAS/internal/broker"
)

func (k *kernel) CreateAndRunRodent(ctx context.Context, wg *sync.WaitGroup) {
	k.mu.Lock()
	defer k.mu.Unlock()

	nr := rodent.NewRodent(ctx, uuid.New(), wg, k)
	k.rodentTable[nr.GetPID()] = nr

	wg.Add(1)
	nr.Run(wg)

	return
}

func (k *kernel) WaitingNewRodent() {
	rodentReqToBornChan := Broker.GetRodentReqToBornChan()
	go func() {
		for _ = range rodentReqToBornChan {
			k.CreateAndRunRodent(k.appCtx, k.wg)
		}
	}()
}

func (k *kernel) CheckRodentByPID(pid uint32) bool {
	k.mu.Lock()
	defer k.mu.Unlock()

	if _, ok := k.rodentTable[pid]; ok {
		return true
	}

	return false
}

func (k *kernel) DeleteRodentByPID(pid uint32) {
	k.mu.Lock()
	defer k.mu.Unlock()

	_, ok := k.rodentTable[pid]
	if !ok {
		return
	}

	delete(k.rodentTable, pid)
}

func (k *kernel) deleteRodentByPID(pid uint32) {
	_, ok := k.rodentTable[pid]
	if !ok {
		return
	}

	delete(k.rodentTable, pid)
}
