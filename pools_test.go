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
		t.Errorf("invalid entity id/gen")
	}
	p.Add(e)
	w.Destroy()
}

func TestSamePools(t *testing.T) {
	w := ecs.NewWorld()
	p1 := ecs.GetPool[C1](w)
	p2 := ecs.GetPool[C1](w)
	if p1 != p2 {
		t.Errorf("pools are not equal.")
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
		t.Errorf("invalid entity id/gen")
	}
	w.DelEntity(e)
	w.Destroy()
}

func TestComponentAutoReset(t *testing.T) {
	w := ecs.NewWorld()
	e := w.NewEntity()
	p := ecs.GetPool[C2](w)
	p.Add(e)
	c2 := p.Get(e)
	if c2.ID != -1 {
		t.Errorf("invalid component reset on new component.")
	}
	c2.ID = 1
	w.DelEntity(e)

	e = w.NewEntity()
	c2 = p.Add(e)
	if c2.ID != -1 {
		t.Errorf("invalid component reset on reused component.")
	}
	w.Destroy()
}

func TestGetAdditionalInfo(t *testing.T) {
	w := ecs.NewWorld()
	p := ecs.GetPool[C1](w)
	if w != p.GetWorld() {
		t.Errorf("invalid world in pool.")
	}
}

func TestInvalidAdd(t *testing.T) {
	w := ecs.NewWorld()
	defer func(world *ecs.World) {
		if r := recover(); r == nil {
			t.Errorf("code should panic.")
		}
		world.Destroy()
	}(w)
	p := ecs.GetPool[C1](w)
	e := w.NewEntity()
	p.Add(e)
	p.Add(e)
	t.Errorf("code should panic.")
}

func TestInvalidGet1(t *testing.T) {
	w := ecs.NewWorld()
	defer func(world *ecs.World) {
		if r := recover(); r == nil {
			t.Errorf("code should panic.")
		}
		world.Destroy()
	}(w)
	p := ecs.GetPool[C2](w)
	p.Get(0)
	t.Errorf("code should panic.")
}

func TestInvalidGet2(t *testing.T) {
	w := ecs.NewWorld()
	e := w.NewEntity()
	defer func(world *ecs.World, entity int) {
		if r := recover(); r == nil {
			t.Errorf("code should panic.")
		}
		world.DelEntity(entity)
		world.Destroy()
	}(w, e)
	p := ecs.GetPool[C2](w)
	p.Get(e)
	t.Errorf("code should panic.")
}

func TestInvalidDel1(t *testing.T) {
	w := ecs.NewWorld()
	defer func(world *ecs.World) {
		if r := recover(); r == nil {
			t.Errorf("code should panic.")
		}
		world.Destroy()
	}(w)
	p := ecs.GetPool[C2](w)
	p.Del(0)
	t.Errorf("code should panic.")
}

func TestInvalidDel2(t *testing.T) {
	w := ecs.NewWorld()
	p := ecs.GetPool[C1](w)
	e := w.NewEntity()
	p.Del(e)
	w.DelEntity(e)
	w.Destroy()
}
