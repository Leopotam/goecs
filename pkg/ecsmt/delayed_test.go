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

type delayedAddSystem struct {
	entities      int
	c1DelayedPool *ecsmt.DelayedPool[c1]
	delayedBuffer ecsmt.IDelayedBuffer
	World         ecsdi.World
	Filter        ecsdi.Filter[ecs.Inc1[c1]]
}

func (s *delayedAddSystem) Init(systems *ecs.Systems) {
	// add one entity with required component,
	// otherwise workers will not be started.
	s.Filter.Pools.Inc1.Add(s.World.Value.NewEntity())

	s.c1DelayedPool = ecsmt.NewDelayedPool[c1]()
	s.delayedBuffer = ecsmt.NewDelayedBuffer(s.World.Value, s.c1DelayedPool)
}

func (s *delayedAddSystem) Run(systems *ecs.Systems) {
	ecsmt.RunTask(s, s.Filter.Value, 10)
	s.delayedBuffer.Process()
}

func (s *delayedAddSystem) Process(entities []int, from, before int) {
	for i := 0; i < s.entities; i++ {
		e := s.delayedBuffer.NewEntity()
		s.c1DelayedPool.Add(e, c1{counter: i + 1})
	}
}

type delayedDelEntitySystem struct {
	chunkSize     int
	delayedBuffer ecsmt.IDelayedBuffer
	World         ecsdi.World
	Filter        ecsdi.Filter[ecs.Inc1[c1]]
}

func (s *delayedDelEntitySystem) Init(systems *ecs.Systems) {
	s.delayedBuffer = ecsmt.NewDelayedBuffer(s.World.Value)
}

func (s *delayedDelEntitySystem) Run(systems *ecs.Systems) {
	ecsmt.RunTask(s, s.Filter.Value, s.chunkSize)
	s.delayedBuffer.Process()
}

func (s *delayedDelEntitySystem) Process(entities []int, from, before int) {
	for i := from; i < before; i++ {
		s.delayedBuffer.DelEntity(entities[i])
	}
}

type delayedDelComponentSystem struct {
	chunkSize     int
	c1DelayedPool *ecsmt.DelayedPool[c1]
	delayedBuffer ecsmt.IDelayedBuffer
	World         ecsdi.World
	Filter        ecsdi.Filter[ecs.Inc1[c1]]
}

func (s *delayedDelComponentSystem) Init(systems *ecs.Systems) {
	s.c1DelayedPool = ecsmt.NewDelayedPool[c1]()
	s.delayedBuffer = ecsmt.NewDelayedBuffer(s.World.Value, s.c1DelayedPool)
}

func (s *delayedDelComponentSystem) Run(systems *ecs.Systems) {
	ecsmt.RunTask(s, s.Filter.Value, s.chunkSize)
	s.delayedBuffer.Process()
}

func (s *delayedDelComponentSystem) Process(entities []int, from, before int) {
	for i := from; i < before; i++ {
		s.c1DelayedPool.Del(entities[i])
	}
}

type delayedHasComponentSystem struct {
	chunkSize     int
	c1DelayedPool *ecsmt.DelayedPool[c1]
	delayedBuffer ecsmt.IDelayedBuffer
	World         ecsdi.World
	Filter        ecsdi.Filter[ecs.Inc1[c1]]
}

func (s *delayedHasComponentSystem) Init(systems *ecs.Systems) {
	s.c1DelayedPool = ecsmt.NewDelayedPool[c1]()
	s.delayedBuffer = ecsmt.NewDelayedBuffer(s.World.Value, s.c1DelayedPool)
}

func (s *delayedHasComponentSystem) Run(systems *ecs.Systems) {
	ecsmt.RunTask(s, s.Filter.Value, s.chunkSize)
}

func (s *delayedHasComponentSystem) Process(entities []int, from, before int) {
	for i := from; i < before; i++ {
		if s.c1DelayedPool.Has(entities[i]) {
			s.c1DelayedPool.Get(entities[i])
		}
	}
}

func TestDelayedDefault(t *testing.T) {
	w := ecs.NewWorld()
	s := ecs.NewSystems(w)
	addSystem := delayedAddSystem{entities: 10}
	s.Add(&addSystem)
	ecsdi.Inject(s)
	s.Init()
	s.Run()
	if addSystem.Filter.Value.GetEntitiesCount() != 11 {
		t.Errorf("invalid entities count after add: %v", addSystem.Filter.Value.GetEntitiesCount())
	}
	s.Destroy()
	w.Destroy()
}

func TestDelayedDelEntity(t *testing.T) {
	w := ecs.NewWorld()
	s := ecs.NewSystems(w)
	addSystem := delayedAddSystem{entities: 100}
	s.Add(&addSystem)
	s.Add(&delayedDelEntitySystem{chunkSize: 10})
	ecsdi.Inject(s)
	s.Init()
	s.Run()
	if addSystem.Filter.Value.GetEntitiesCount() != 0 {
		t.Errorf("invalid entities count after delentity: %v", addSystem.Filter.Value.GetEntitiesCount())
	}
	s.Destroy()
	w.Destroy()
}

func TestDelayedComponent(t *testing.T) {
	w := ecs.NewWorld()
	s := ecs.NewSystems(w)
	addSystem := delayedAddSystem{entities: 100}
	s.Add(&addSystem)
	s.Add(&delayedDelComponentSystem{chunkSize: 10})
	ecsdi.Inject(s)
	s.Init()
	s.Run()
	if addSystem.Filter.Value.GetEntitiesCount() != 0 {
		t.Errorf("invalid entities count after delcomponent: %v", addSystem.Filter.Value.GetEntitiesCount())
	}
	s.Destroy()
	w.Destroy()
}

func TestDelayedHasComponent(t *testing.T) {
	w := ecs.NewWorld()
	s := ecs.NewSystems(w)
	addSystem := delayedAddSystem{entities: 100}
	s.Add(&addSystem)
	s.Add(&delayedHasComponentSystem{chunkSize: 10})
	ecsdi.Inject(s)
	s.Init()
	s.Run()
	if addSystem.Filter.Value.GetEntitiesCount() != 101 {
		t.Errorf("invalid entities count after hascomponent: %v", addSystem.Filter.Value.GetEntitiesCount())
	}
	s.Destroy()
	w.Destroy()
}

func TestDelayedInvalidPoolDoubleLinked(t *testing.T) {
	w := ecs.NewWorld()
	defer func(world *ecs.World) {
		if r := recover(); r == nil {
			t.Errorf("code should panic.")
		}
		world.Destroy()
	}(w)
	p := ecsmt.NewDelayedPool[c1]()
	_ = ecsmt.NewDelayedBuffer(w, p)
	_ = ecsmt.NewDelayedBuffer(w, p)
	t.Errorf("code should panic.")
}

func TestDelayedInvalidPoolAdd(t *testing.T) {
	w := ecs.NewWorld()
	defer func(world *ecs.World) {
		if r := recover(); r == nil {
			t.Errorf("code should panic.")
		}
		world.Destroy()
	}(w)
	var p ecsmt.DelayedPool[c1]
	p.Add(0, c1{})
	t.Errorf("code should panic.")
}

func TestDelayedInvalidPoolGet(t *testing.T) {
	w := ecs.NewWorld()
	defer func(world *ecs.World) {
		if r := recover(); r == nil {
			t.Errorf("code should panic.")
		}
		world.Destroy()
	}(w)
	var p ecsmt.DelayedPool[c1]
	p.Get(0)
	t.Errorf("code should panic.")
}

func TestDelayedInvalidPoolHas(t *testing.T) {
	w := ecs.NewWorld()
	defer func(world *ecs.World) {
		if r := recover(); r == nil {
			t.Errorf("code should panic.")
		}
		world.Destroy()
	}(w)
	var p ecsmt.DelayedPool[c1]
	p.Has(0)
	t.Errorf("code should panic.")
}

func TestDelayedInvalidPoolDel(t *testing.T) {
	w := ecs.NewWorld()
	defer func(world *ecs.World) {
		if r := recover(); r == nil {
			t.Errorf("code should panic.")
		}
		world.Destroy()
	}(w)
	var p ecsmt.DelayedPool[c1]
	p.Del(0)
	t.Errorf("code should panic.")
}

func TestDelayedInvalidPoolDelDelayed(t *testing.T) {
	w := ecs.NewWorld()
	defer func(world *ecs.World) {
		if r := recover(); r == nil {
			t.Errorf("code should panic.")
		}
		world.Destroy()
	}(w)
	p := ecsmt.NewDelayedPool[c1]()
	_ = ecsmt.NewDelayedBuffer(w, p)
	p.Del(-1)
	t.Errorf("code should panic.")
}

func TestDelayedInvalidPoolHasDelayed(t *testing.T) {
	w := ecs.NewWorld()
	defer func(world *ecs.World) {
		if r := recover(); r == nil {
			t.Errorf("code should panic.")
		}
		world.Destroy()
	}(w)
	p := ecsmt.NewDelayedPool[c1]()
	_ = ecsmt.NewDelayedBuffer(w, p)
	p.Has(-1)
	t.Errorf("code should panic.")
}

func TestDelayedInvalidPoolGetDelayed(t *testing.T) {
	w := ecs.NewWorld()
	defer func(world *ecs.World) {
		if r := recover(); r == nil {
			t.Errorf("code should panic.")
		}
		world.Destroy()
	}(w)
	p := ecsmt.NewDelayedPool[c1]()
	_ = ecsmt.NewDelayedBuffer(w, p)
	p.Get(-1)
	t.Errorf("code should panic.")
}

func TestDelayedInvalidBufferDelEntityDelayed(t *testing.T) {
	w := ecs.NewWorld()
	defer func(world *ecs.World) {
		if r := recover(); r == nil {
			t.Errorf("code should panic.")
		}
		world.Destroy()
	}(w)
	p := ecsmt.NewDelayedPool[c1]()
	b := ecsmt.NewDelayedBuffer(w, p)
	b.DelEntity(-1)
	t.Errorf("code should panic.")
}

func TestDelayedInvalidBufferDoubleDelEntity(t *testing.T) {
	w := ecs.NewWorld()
	defer func(world *ecs.World) {
		if r := recover(); r == nil {
			t.Errorf("code should panic.")
		}
		world.Destroy()
	}(w)
	f := ecs.GetFilter[ecs.Inc1[c1]](w)
	p := ecsmt.NewDelayedPool[c1]()
	b := ecsmt.NewDelayedBuffer(w, p)
	p.Add(b.NewEntity(), c1{})
	b.Process()
	if f.GetEntitiesCount() != 1 {
		return
	}
	e := f.GetRawEntities()[0]
	b.DelEntity(e)
	b.DelEntity(e)
	b.Process()
	t.Errorf("code should panic.")
}
