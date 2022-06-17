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

type mtSystem struct {
	entities  int
	chunkSize int
	payload   int
	World     ecsdi.World
	Filter    ecsdi.Filter[ecs.Inc1[c1]]
	C1Pool    ecsdi.Pool[c1]
}

func newSystem(entities, chunk, payload int) any {
	return &mtSystem{
		entities:  entities,
		chunkSize: chunk,
		payload:   payload,
	}
}

func (t *mtSystem) Init(systems *ecs.Systems) {
	for i := 0; i < t.entities; i++ {
		t.C1Pool.Value.Add(t.World.Value.NewEntity())
	}
}

func (t *mtSystem) Run(systems *ecs.Systems) {
	ecsmt.Run(t, t.Filter.Value, t.chunkSize)
}

func (t *mtSystem) Process(entities []int, from, before int) {
	for i := from; i < before; i++ {
		c1 := t.Filter.Pools.Inc1.Get(entities[i])
		for i := 0; i < 10000; i++ {
			c1.counter = (c1.counter + 1) % 10000
		}
	}
}

func TestMT(t *testing.T) {
	w := ecs.NewWorld()
	s := ecs.NewSystems(w)
	s.Add(newSystem(10, 1, 1))
	ecsdi.Inject(s)
	s.Init()
	s.Run()
	s.Destroy()
	w.Destroy()
}

func TestMTEmptyTask(t *testing.T) {
	w := ecs.NewWorld()
	s := ecs.NewSystems(w)
	s.Add(newSystem(0, 5, 1))
	ecsdi.Inject(s)
	s.Init()
	s.Run()
	s.Destroy()
	w.Destroy()
}

func TestMTHugeData(t *testing.T) {
	w := ecs.NewWorld()
	s := ecs.NewSystems(w)
	s.Add(newSystem(100, 5, 1))
	ecsdi.Inject(s)
	s.Init()
	s.Run()
	s.Destroy()
	w.Destroy()
}

func TestMTSmallData(t *testing.T) {
	w := ecs.NewWorld()
	s := ecs.NewSystems(w)
	s.Add(newSystem(1, 50, 1))
	ecsdi.Inject(s)
	s.Init()
	s.Run()
	s.Destroy()
	w.Destroy()
}

func BenchmarkMT(b *testing.B) {
	w := ecs.NewWorld()
	s := ecs.NewSystems(w)
	s.Add(newSystem(10000, 1000, 1000))
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
