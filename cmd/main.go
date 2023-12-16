package main

import (
	"context"
	"sync"
	"time"

	. "SimulatorMAS/internal/config_parser"
	. "SimulatorMAS/internal/kernel"
	. "SimulatorMAS/internal/population"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := sync.WaitGroup{}
	//wg.Add(1)

	Kernel = NewKernel(ctx, &wg)
	Kernel.WaitingNewGreens()
	Kernel.WaitingNewRodent()
	Kernel.CollectingMetrics()

	CreateAndRunPopulation(ctx, &wg, GreensAgent, Cfg.Greens.Count)

	time.Sleep(time.Second * 2)

	CreateAndRunPopulation(ctx, &wg, RodentAgent, Cfg.Rodent.Count)

	wg.Wait()
	return
}
