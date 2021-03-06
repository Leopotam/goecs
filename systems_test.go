// ----------------------------------------------------------------------------
// The Proprietary or MIT-Red License
// Copyright (c) 2012-2022 Leopotam <leopotam@yandex.ru>
// ----------------------------------------------------------------------------

package ecs_test

import (
	"testing"

	"leopotam.com/go/ecs"
)

type PreInitSystem1 struct {
	Counter *int
}
type InitSystem1 struct {
	Counter *int
}
type RunSystem1 struct {
	Counter *int
}
type DestroySystem1 struct {
	Counter *int
}
type PostDestroySystem1 struct {
	Counter *int
}

func (s *PreInitSystem1) PreInit(systems ecs.ISystems) {
	*s.Counter++
}
func (s *InitSystem1) Init(systems ecs.ISystems) {
	*s.Counter++
}
func (s *RunSystem1) Run(systems ecs.ISystems) {
	*s.Counter++
}
func (s *DestroySystem1) Destroy(systems ecs.ISystems) {
	*s.Counter++
}
func (s *PostDestroySystem1) PostDestroy(systems ecs.ISystems) {
	*s.Counter++
}

type InvalidSystem1 struct{}
type PreInitInvalidSystem1 struct{}
type InitInvalidSystem1 struct{}
type InitInvalidSystem2 struct{}
type RunInvalidSystem1 struct{}
type DestroyInvalidSystem1 struct{}
type PostDestroyInvalidSystem1 struct{}

func (s *PreInitInvalidSystem1) PreInit(systems ecs.ISystems) { systems.GetWorld("").NewEntity() }
func (s *InitInvalidSystem1) Init(systems ecs.ISystems)       { systems.GetWorld("").NewEntity() }
func (s *RunInvalidSystem1) Run(systems ecs.ISystems)         { systems.GetWorld("").NewEntity() }
func (s *DestroyInvalidSystem1) Destroy(systems ecs.ISystems) { systems.GetWorld("").NewEntity() }
func (s *PostDestroyInvalidSystem1) PostDestroy(systems ecs.ISystems) {
	systems.GetWorld("").NewEntity()
}
func (s *InitInvalidSystem2) Init(systems ecs.ISystems) {
	systems.GetWorld("events").NewEntity()
}

func TestSystemRegistration(t *testing.T) {
	w := ecs.NewWorld()
	s := ecs.NewSystems(w)
	counter := 0
	s.
		Add(&PreInitSystem1{Counter: &counter}).
		Add(&InitSystem1{Counter: &counter}).
		Add(&RunSystem1{Counter: &counter}).
		Add(&DestroySystem1{Counter: &counter}).
		Add(&PostDestroySystem1{Counter: &counter}).
		Init()
	s.Run()
	s.Destroy()
	if counter != 5 {
		t.Errorf("invalid system calls")
	}
	w.Destroy()
}

func TestSystemGetWorlds(t *testing.T) {
	w1 := ecs.NewWorld()
	w2 := ecs.NewWorld()
	s := ecs.NewSystems(w1)
	s.
		AddWorld(w2, "events").
		Init()
	if s.GetWorld("") != w1 {
		t.Errorf("invalid default world")
	}
	if s.GetWorld("events") != w2 {
		t.Errorf("invalid named world")
	}
	if s.GetWorld("events1") != nil {
		t.Errorf("invalid named world")
	}
	namedWorlds := s.GetNamedWorlds()
	if len(namedWorlds) != 1 {
		t.Errorf("invalid named worlds amount: %v", len(namedWorlds))
	}
	s.Destroy()
	w1.Destroy()
	w2.Destroy()
}

func TestSystemGetAllSystems(t *testing.T) {
	w := ecs.NewWorld()
	counter := 0
	sList := []any{
		&PreInitSystem1{Counter: &counter},
		&InitSystem1{Counter: &counter},
		&RunSystem1{Counter: &counter},
		&DestroySystem1{Counter: &counter},
		&PostDestroySystem1{Counter: &counter},
	}
	s := ecs.NewSystems(w)
	for _, ss := range sList {
		s.Add(ss)
	}
	s.Init()
	allSystems := s.GetAllSystems()
	if len(allSystems) != len(sList) {
		t.Errorf("invalid systems amount")
	}
	s.Destroy()
	w.Destroy()
}

func TestSystemGetRunSystems(t *testing.T) {
	w := ecs.NewWorld()
	counter := 0
	sList := []any{
		&PreInitSystem1{Counter: &counter},
		&InitSystem1{Counter: &counter},
		&RunSystem1{Counter: &counter},
		&DestroySystem1{Counter: &counter},
		&PostDestroySystem1{Counter: &counter},
	}
	s := ecs.NewSystems(w)
	for _, ss := range sList {
		s.Add(ss)
	}
	s.Init()
	runsCount := 0
	for _, v := range s.GetAllSystems() {
		if _, ok := v.(ecs.IRunSystem); ok {
			runsCount++
		}
	}
	if runsCount != 1 {
		t.Errorf("invalid run systems amount: %v", runsCount)
	}
	s.Destroy()
	w.Destroy()
}

func TestSystemsInvalidType(t *testing.T) {
	w := ecs.NewWorld()
	systems := ecs.NewSystems(w)
	defer func(world *ecs.World, systems ecs.ISystems) {
		if r := recover(); r == nil {
			t.Errorf("code should panic")
		}
		systems.Destroy()
		world.Destroy()
	}(w, systems)
	systems.Add(&InvalidSystem1{})
	t.Errorf("code should panic")
}

func TestSystemsLeakedPreInit(t *testing.T) {
	w := ecs.NewWorld()
	systems := ecs.NewSystems(w)
	defer func(world *ecs.World, systems ecs.ISystems) {
		if r := recover(); r == nil {
			t.Errorf("code should panic")
		}
		systems.Destroy()
		world.Destroy()
	}(w, systems)
	systems.Add(&PreInitInvalidSystem1{})
	systems.Init()
	t.Errorf("code should panic")
}

func TestSystemsLeakedInit(t *testing.T) {
	w := ecs.NewWorld()
	systems := ecs.NewSystems(w)
	defer func(world *ecs.World, systems ecs.ISystems) {
		if r := recover(); r == nil {
			t.Errorf("code should panic")
		}
		systems.Destroy()
		world.Destroy()
	}(w, systems)
	systems.Add(&InitInvalidSystem1{})
	systems.Init()
	t.Errorf("code should panic")
}

func TestSystemsLeakedAtNamedWorldInit(t *testing.T) {
	w := ecs.NewWorld()
	w1 := ecs.NewWorld()
	systems := ecs.NewSystems(w)
	defer func(world *ecs.World, systems ecs.ISystems) {
		if r := recover(); r == nil {
			t.Errorf("code should panic")
		}
		systems.Destroy()
		world.Destroy()
	}(w, systems)
	systems.
		AddWorld(w1, "events").
		Add(&InitInvalidSystem2{}).
		Init()
	t.Errorf("code should panic")
}

func TestSystemsLeakedRun(t *testing.T) {
	w := ecs.NewWorld()
	systems := ecs.NewSystems(w)
	defer func(world *ecs.World, systems ecs.ISystems) {
		if r := recover(); r == nil {
			t.Errorf("code should panic")
		}
		systems.Destroy()
		world.Destroy()
	}(w, systems)
	systems.Add(&RunInvalidSystem1{})
	systems.Init()
	systems.Run()
	t.Errorf("code should panic")
}

func TestSystemsLeakedDestroy(t *testing.T) {
	w := ecs.NewWorld()
	systems := ecs.NewSystems(w)
	defer func(world *ecs.World, systems ecs.ISystems) {
		if r := recover(); r == nil {
			t.Errorf("code should panic")
		}
		world.Destroy()
	}(w, systems)
	systems.Add(&DestroyInvalidSystem1{})
	systems.Init()
	systems.Destroy()
	t.Errorf("code should panic")
}

func TestSystemsLeakedPostDestroy(t *testing.T) {
	w := ecs.NewWorld()
	systems := ecs.NewSystems(w)
	defer func(world *ecs.World, systems ecs.ISystems) {
		if r := recover(); r == nil {
			t.Errorf("code should panic")
		}
		world.Destroy()
	}(w, systems)
	systems.Add(&PostDestroyInvalidSystem1{})
	systems.Init()
	systems.Destroy()
	t.Errorf("code should panic")
}

func TestSystemsAddWorldTwice(t *testing.T) {
	w := ecs.NewWorld()
	systems := ecs.NewSystems(w)
	defer func(world *ecs.World, systems ecs.ISystems) {
		if r := recover(); r == nil {
			t.Errorf("code should panic")
		}
		world.Destroy()
	}(w, systems)
	systems.AddWorld(w, "events")
	systems.AddWorld(w, "events")
	t.Errorf("code should panic")
}
