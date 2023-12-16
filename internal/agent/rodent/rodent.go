package rodent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	. "SimulatorMAS/internal/broker"
	. "SimulatorMAS/internal/config_parser"
	. "SimulatorMAS/internal/logger"
	"SimulatorMAS/internal/types"
)

type rodent struct {
	mu                  sync.RWMutex
	kernel              types.IKernel
	pid                 uint32
	maxHungryToDie      int
	currentHungryToDie  int
	maxLifeTime         time.Duration
	intervalAddLifeTime time.Duration
	intervalLunchTime   time.Duration
	NumberOfCorpses     int
	appCtx              context.Context
	rodentCtx           context.Context
	rodentCtxCancel     context.CancelFunc
	killCh              chan struct{}
	ageCh               chan struct{}
	hungryCh            chan struct{}
	isAlive             bool
	appWG               *sync.WaitGroup
}

func NewRodent(ctx context.Context, uuidValue uuid.UUID, wg *sync.WaitGroup, kernel types.IKernel) *rodent {
	rodentCtx, cancel := context.WithCancel(context.Background())
	return &rodent{
		pid:                 uuidValue.ID(),
		appCtx:              ctx,
		rodentCtx:           rodentCtx,
		rodentCtxCancel:     cancel,
		kernel:              kernel,
		maxLifeTime:         Cfg.Rodent.MaxLifeTime,
		intervalAddLifeTime: Cfg.Rodent.IntervalAddLifeTime,
		intervalLunchTime:   Cfg.Rodent.IntervalLunchTime,
		NumberOfCorpses:     Cfg.Rodent.NumberOfCorpses,
		killCh:              make(chan struct{}, 1),
		ageCh:               make(chan struct{}, 1),
		hungryCh:            make(chan struct{}, 1),
		currentHungryToDie:  0,
		maxHungryToDie:      Cfg.Rodent.MaxHungryToDie,
		isAlive:             true,
		appWG:               wg,
	}
}

func (r *rodent) Kill() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.isAlive = false

	r.killCh <- struct{}{}
}

func (r *rodent) IsAlive() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	res := r.isAlive

	return res
}

func (r *rodent) GetPID() uint32 {
	return r.pid
}

func (r *rodent) Run(wg *sync.WaitGroup) {
	RodentInfo("Rodent № %d: I am born", r.pid)
	go r.growth()
	go func() {
		select {
		case <-r.appCtx.Done():
			r.mu.Lock()
			r.rodentCtxCancel()
			r.isAlive = false
			r.ageCh <- struct{}{}
			r.killCh <- struct{}{}
			r.mu.Unlock()

			RodentInfo("Rodent № %d: I am dead. Some unknown reasons", r.pid)
			wg.Done()
			return
		case <-r.ageCh:
			r.mu.Lock()
			r.rodentCtxCancel()
			r.mu.Unlock()

			RodentInfo("Rodent № %d: I became old and died", r.pid)
			wg.Done()
			return
		case <-r.hungryCh:
			r.mu.Lock()
			r.rodentCtxCancel()
			r.mu.Unlock()

			RodentInfo("Rodent № %d: I starved to death", r.pid)
			wg.Done()
			return
		case <-r.killCh:
			r.mu.Lock()
			r.rodentCtxCancel()
			r.mu.Unlock()

			RodentInfo("Rodent № %d: I am dead", r.pid)
			wg.Done()
			return
		}
	}()
	return
}

func (r *rodent) growth() {
	tickerAddLifeTime := time.NewTicker(r.intervalAddLifeTime)
	tickerLunchTime := time.NewTicker(r.intervalLunchTime)
	currentLifeTime := time.Duration(0)
	for {
		select {
		case <-tickerAddLifeTime.C:
			currentLifeTime += r.intervalAddLifeTime

			if currentLifeTime >= r.maxLifeTime {
				r.kernel.DeleteRodentByPID(r.pid)

				r.mu.Lock()
				r.isAlive = false
				r.ageCh <- struct{}{}
				r.mu.Unlock()

				r.kernel.NumberCorpseAdd(r.NumberOfCorpses)
				err := Broker.PutMessageToQueueNTimes(
					"corpses",
					fmt.Sprintf("corpse type: Rodent, pid: %d", r.pid),
					r.NumberOfCorpses,
				)
				if err != nil {
					RodentError("Rodent № %d: I can't push message", r.pid)
				}

				return
			}
		case <-tickerLunchTime.C:
			info, isExist := r.kernel.IsReproductionPossible()
			if isExist {
				RodentInfo("Rodent № %d: I get greens. Greens info: '%s', Starting reproduction... ", r.pid, info)
				err := Broker.PutMessageToQueue("rodent_req_to_born", fmt.Sprintf("from Rodent №: %d", r.pid))
				if err != nil {
					RodentError("Rodent № %d: I can't push message", r.pid)
				}
				r.currentHungryToDie = 0
			} else {
				r.currentHungryToDie += 1
			}
			if r.currentHungryToDie >= r.maxHungryToDie {
				r.kernel.DeleteRodentByPID(r.pid)

				r.mu.Lock()
				r.isAlive = false
				r.hungryCh <- struct{}{}
				r.mu.Unlock()

				r.kernel.NumberCorpseAdd(r.NumberOfCorpses)
				err := Broker.PutMessageToQueueNTimes(
					"corpses",
					fmt.Sprintf("corpse type: Rodent, pid: %d", r.pid),
					r.NumberOfCorpses,
				)
				if err != nil {
					RodentError("Rodent № %d: I can't push message", r.pid)
				}
			}
		case <-r.rodentCtx.Done():
			return
		}
	}
}
