// ----------------------------------------------------------------------------
// The Proprietary or MIT-Red License
// Copyright (c) 2012-2022 Leopotam <leopotam@yandex.ru>
// ----------------------------------------------------------------------------

package ecs_test

import (
	"ecs"
	"testing"
)

func TestEntityWithComponent(t *testing.T) {
	w := ecs.NewWorld()
	p := ecs.GetPool[C1](w)
	e := w.NewEntity()
	if e != 0 && w.GetEntityGen(e) != 1 {
		t.Fatalf("invalid entity id/gen")
	}
	p.Add(e)
	w.Destroy()
}

func TestSamePools(t *testing.T) {
	w := ecs.NewWorld()
	p1 := ecs.GetPool[C1](w)
	p2 := ecs.GetPool[C1](w)
	if p1 != p2 {
		t.Fatalf("pools are not equal.")
	}
	w.Destroy()
}

func TestReuseEntityID(t *testing.T) {
	w := ecs.NewWorld()
	e := w.NewEntity()
	id1 := e
	gen1 := w.GetEntityGen(e)
	w.DelEntity(e)
	e = w.NewEntity()
	id2 := e
	gen2 := w.GetEntityGen(e)
	if id1 != id2 && gen2 != gen1+1 {
		t.Fatalf("invalid entity id/gen")
	}
	w.Destroy()
}

func TestComponentAutoReset(t *testing.T) {
	w := ecs.NewWorld()
	e := w.NewEntity()
	p := ecs.GetPool[C2](w)
	p.Add(e)
	c2 := p.Get(e)
	if c2.ID != -1 {
		t.Fatalf("invalid component reset on new component.")
	}
	c2.ID = 1
	w.DelEntity(e)

	e = w.NewEntity()
	c2 = p.Add(e)
	if c2.ID != -1 {
		t.Fatalf("invalid component reset on reused component.")
	}
	w.Destroy()
}
