package population

import (
	"context"
	"sync"
	"time"

	. "SimulatorMAS/internal/kernel"
)

type AgentType int

const (
	GreensAgent AgentType = iota
	RodentAgent
)

type Subject interface {
	CreateAndRun(ctx context.Context, wg *sync.WaitGroup)
}

func CreateAndRunPopulation(ctx context.Context, wg *sync.WaitGroup, agentType AgentType, countAgents int) {
	for i := 0; i < countAgents; i++ {
		time.Sleep(100 * time.Millisecond)
		switch agentType {
		case GreensAgent:
			Kernel.CreateAndRunGreens(ctx, wg)
		case RodentAgent:
			Kernel.CreateAndRunRodent(ctx, wg)
		}
	}
}
