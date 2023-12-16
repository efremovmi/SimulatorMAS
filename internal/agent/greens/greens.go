package greens

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

type greens struct {
	mu                  sync.RWMutex
	kernel              types.IKernel
	pid                 uint32
	maxLifeTime         time.Duration
	intervalAddLifeTime time.Duration
	intervalLunchTime   time.Duration
	appCtx              context.Context
	greensCtx           context.Context
	greensCtxCancel     context.CancelFunc
	killCh              chan struct{}
	ageCh               chan struct{}
	isAlive             bool
	appWG               *sync.WaitGroup
}

func NewGreens(ctx context.Context, uuidValue uuid.UUID, wg *sync.WaitGroup, kernel types.IKernel) *greens {
	greensCtx, cancel := context.WithCancel(context.Background())
	return &greens{
		pid:                 uuidValue.ID(),
		kernel:              kernel,
		maxLifeTime:         Cfg.Greens.MaxLifeTime,
		intervalAddLifeTime: Cfg.Greens.IntervalAddLifeTime,
		intervalLunchTime:   Cfg.Greens.IntervalLunchTime,
		appCtx:              ctx,
		greensCtx:           greensCtx,
		greensCtxCancel:     cancel,
		killCh:              make(chan struct{}, 1),
		ageCh:               make(chan struct{}, 1),
		isAlive:             true,
		appWG:               wg,
	}
}

func (g *greens) Kill() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.isAlive = false

	g.killCh <- struct{}{}
}

func (g *greens) IsAlive() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	res := g.isAlive

	return res
}

func (g *greens) GetPID() uint32 {
	return g.pid
}

func (g *greens) Run(wg *sync.WaitGroup) {
	GreensInfo("Greens № %d: I am born", g.pid)
	go g.growth()
	go func() {
		select {
		case <-g.appCtx.Done():
			g.mu.Lock()
			g.greensCtxCancel()
			g.isAlive = false
			g.ageCh <- struct{}{}
			g.killCh <- struct{}{}
			g.mu.Unlock()

			GreensInfo("Greens № %d: I am dead. Some unknown reasons", g.pid)
			wg.Done()
			return
		case <-g.ageCh:
			g.mu.Lock()
			g.greensCtxCancel()
			g.mu.Unlock()

			GreensInfo("Greens № %d: I became old and died", g.pid)
			wg.Done()
			return
		case <-g.killCh:
			g.mu.Lock()
			g.greensCtxCancel()
			g.mu.Unlock()

			GreensInfo("Greens № %d: I am dead", g.pid)
			wg.Done()
			return
		}
	}()
	return
}

func (g *greens) growth() {
	tickerAddLifeTime := time.NewTicker(g.intervalAddLifeTime)
	tickerLunchTime := time.NewTicker(g.intervalLunchTime)
	currentLifeTime := time.Duration(0)
	for {
		select {
		case <-tickerAddLifeTime.C:
			currentLifeTime += g.intervalAddLifeTime

			if currentLifeTime >= g.maxLifeTime {
				g.kernel.DeleteGreensByPID(g.pid)

				g.mu.Lock()
				g.isAlive = false
				g.ageCh <- struct{}{}
				g.mu.Unlock()

				g.kernel.NumberCorpseInc()
				err := Broker.PutMessageToQueue("corpses",
					fmt.Sprintf("corpse type: Greens, pid: %d", g.pid))
				if err != nil {
					GreensError("Greens № %d: I can't push message", g.pid)
				}

				return
			}
		case <-tickerLunchTime.C:
			msg, err := Broker.GetMessageFromQueue("corpses")
			if err == nil {
				GreensInfo("Greens № %d: I get corpse. Corpse info: '%s', Starting replication... ", g.pid, msg)
				err := Broker.PutMessageToQueue("greens_req_to_born", fmt.Sprintf("from greens №: %d", g.pid))
				if err != nil {
					GreensError("Greens № %d: I can't push message", g.pid)
				}

			}
		case <-g.greensCtx.Done():
			return
		}
	}
}
