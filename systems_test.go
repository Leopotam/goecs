// ----------------------------------------------------------------------------
// The Proprietary or MIT-Red License
// Copyright (c) 2012-2022 Leopotam <leopotam@yandex.ru>
// ----------------------------------------------------------------------------

package ecs_test

import (
	"ecs"
	"testing"
)

type InitSystem1 struct {
	Counter *int
}
type RunSystem1 struct {
	Counter *int
}
type DestroySystem1 struct {
	Counter *int
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

func TestSystemRegistration(t *testing.T) {
	w := ecs.NewWorld()
	s := ecs.NewSystems(w)
	counter := 0
	s.
		Add(&InitSystem1{Counter: &counter}).
		Add(&RunSystem1{Counter: &counter}).
		Add(&DestroySystem1{Counter: &counter}).
		Init()
	s.Run()
	s.Destroy()
	if counter != 3 {
		t.Fatalf("invalid system calls.")
	}
	w.Destroy()
}
