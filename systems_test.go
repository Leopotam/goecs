// ----------------------------------------------------------------------------
// The Proprietary or MIT-Red License
// Copyright (c) 2012-2022 Leopotam <leopotam@yandex.ru>
// ----------------------------------------------------------------------------

package ecs_test

import (
	"ecs"
	"testing"
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

func (s *PreInitSystem1) PreInit(systems *ecs.Systems) {
	*s.Counter++
}
func (s *InitSystem1) Init(systems *ecs.Systems) {
	*s.Counter++
}
func (s *RunSystem1) Run(systems *ecs.Systems) {
	*s.Counter++
}
func (s *DestroySystem1) Destroy(systems *ecs.Systems) {
	*s.Counter++
}
func (s *PostDestroySystem1) PostDestroy(systems *ecs.Systems) {
	*s.Counter++
}

type invalidSystem struct{}

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
		t.Errorf("invalid system calls.")
	}
	w.Destroy()
}

func TestSystemsInvalidType(t *testing.T) {
	w := ecs.NewWorld()
	systems := ecs.NewSystems(w)
	defer func(world *ecs.World, systems *ecs.Systems) {
		if r := recover(); r == nil {
			t.Errorf("code should panic.")
		}
		if systems != nil {
			systems.Destroy()
		}
		if world != nil {
			world.Destroy()
		}
	}(w, systems)
	systems.Add(&invalidSystem{})
	t.Errorf("code should panic.")
}
