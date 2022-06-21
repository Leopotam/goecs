// ----------------------------------------------------------------------------
// The Proprietary or MIT-Red License
// Copyright (c) 2012-2022 Leopotam <leopotam@yandex.ru>
// ----------------------------------------------------------------------------

package ecsmt_test

import (
	"testing"

	"leopotam.com/go/ecs"
	"leopotam.com/go/ecs/pkg/ecsdi"
	"leopotam.com/go/ecs/pkg/ecsmt"
)

type c1 struct {
	counter int
}

type wrkSystem struct {
	entities  int
	chunkSize int
	payload   int
	World     ecsdi.World
	Filter    ecsdi.Filter[ecs.Inc1[c1]]
	C1Pool    ecsdi.Pool[c1]
}

func newWrkSystem(entities, chunk, payload int) any {
	return &wrkSystem{
		entities:  entities,
		chunkSize: chunk,
		payload:   payload,
	}
}

func (s *wrkSystem) Init(systems *ecs.Systems) {
	for i := 0; i < s.entities; i++ {
		s.C1Pool.Value.Add(s.World.Value.NewEntity())
	}
}

func (s *wrkSystem) Run(systems *ecs.Systems) {
	ecsmt.Run(s, s.Filter.Value, s.chunkSize)
}

func (s *wrkSystem) Process(entities []int, from, before int) {
	for i := from; i < before; i++ {
		c1 := s.Filter.Pools.Inc1.Get(entities[i])
		for i := 0; i < s.payload; i++ {
			c1.counter = (c1.counter + 1) % 10000
		}
	}
}

func TestWrkDefault(t *testing.T) {
	w := ecs.NewWorld()
	s := ecs.NewSystems(w)
	s.Add(newWrkSystem(10, 1, 1))
	ecsdi.Inject(s)
	s.Init()
	s.Run()
	s.Destroy()
	w.Destroy()
}

func TestWrkEmptyTask(t *testing.T) {
	w := ecs.NewWorld()
	s := ecs.NewSystems(w)
	s.Add(newWrkSystem(0, 5, 1))
	ecsdi.Inject(s)
	s.Init()
	s.Run()
	s.Destroy()
	w.Destroy()
}

func TestWrkHugeData(t *testing.T) {
	w := ecs.NewWorld()
	s := ecs.NewSystems(w)
	s.Add(newWrkSystem(100, 5, 1))
	ecsdi.Inject(s)
	s.Init()
	s.Run()
	s.Destroy()
	w.Destroy()
}

func TestWrkSmallData(t *testing.T) {
	w := ecs.NewWorld()
	s := ecs.NewSystems(w)
	s.Add(newWrkSystem(1, 50, 1))
	ecsdi.Inject(s)
	s.Init()
	s.Run()
	s.Destroy()
	w.Destroy()
}

func BenchmarkWorkers(b *testing.B) {
	w := ecs.NewWorld()
	s := ecs.NewSystems(w)
	s.Add(newWrkSystem(10000, 1000, 1000))
	ecsdi.Inject(s)
	s.Init()
	s.Run()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Run()
	}
	b.StopTimer()
	s.Destroy()
	w.Destroy()
}
