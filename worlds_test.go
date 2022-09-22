// ----------------------------------------------------------------------------
// The Proprietary or MIT-Red License
// Copyright (c) 2012-2022 Leopotam <leopotam@yandex.ru>
// ----------------------------------------------------------------------------

package ecs_test

import (
	"math"
	"reflect"
	"testing"

	"leopotam.com/go/ecs"
)

type C1 struct{}
type C2 struct{ ID int }
type C3 struct{ ID int }
type C4 struct{}
type C5 struct{}
type C6 struct{}

func (c2 *C2) Reset() {
	c2.ID = -1
}

func (c2 *C2) Copy(src *C2) {
	c2.ID = src.ID * 2
}

func TestWorldCreate(t *testing.T) {
	w := ecs.NewWorld()
	w.Destroy()
}

func TestWorldResize(t *testing.T) {
	w := ecs.NewWorldWithConfig(ecs.WorldConfig{WorldEntitiesSize: 2})
	p := ecs.GetPool[C1](w)
	_ = ecs.GetFilter[ecs.Inc1[C1]](w)
	for i := 0; i < 3; i++ {
		p.Add(w.NewEntity())
	}
	size := len(w.GetRawEntities())
	if size != w.GetRawEntityItemSize()*3 {
		t.Errorf("invalid raw entities size: %d.", size)
	}
	w.Destroy()
}

func TestWorldRawEntitySize(t *testing.T) {
	compSize := 2
	w := ecs.NewWorldWithConfig(ecs.WorldConfig{EntityComponentsSize: compSize})
	size := w.GetRawEntityItemSize()
	if size != (ecs.RawEntityOffsetComponents + compSize) {
		t.Errorf("invalid raw entity item size: %d.", size)
	}
	w.Destroy()
}

func TestWorldEntityComponentsResize(t *testing.T) {
	w := ecs.NewWorldWithConfig(ecs.WorldConfig{EntityComponentsSize: 2})
	p1 := ecs.GetPool[C1](w)
	p2 := ecs.GetPool[C2](w)
	p3 := ecs.GetPool[C3](w)
	e := w.NewEntity()
	p1.Add(e)
	p2.Add(e)
	p3.Add(e)
	w.Destroy()
}

func TestRawEntityData(t *testing.T) {
	w := ecs.NewWorld()
	p1 := ecs.GetPool[C1](w)
	p2 := ecs.GetPool[C2](w)
	f := ecs.GetFilterWithExc[ecs.Inc1[C1], ecs.Exc1[C2]](w)
	if count := f.GetEntitiesCount(); count != 0 {
		t.Errorf("invalid fcount: %d", count)
	}
	e := w.NewEntity()
	offset := w.GetRawEntityOffset(e)
	entites := w.GetRawEntities()
	if gen := entites[offset+ecs.RawEntityOffsetGen]; gen != 1 {
		t.Errorf("invalid gen: %d", gen)
	}
	if count := entites[offset+ecs.RawEntityOffsetComponentsCount]; count != 0 {
		t.Errorf("invalid ccount: %d", count)
	}
	p1.Add(e)
	if count := f.GetEntitiesCount(); count != 1 {
		t.Errorf("invalid fcount: %d", count)
	}
	if count := entites[offset+ecs.RawEntityOffsetComponentsCount]; count != 1 {
		t.Errorf("invalid ccount: %d", count)
	}
	if c1 := entites[offset+ecs.RawEntityOffsetComponents]; c1 != p1.GetID() {
		t.Errorf("invalid c1: %d", c1)
	}
	p2.Add(e)
	if count := f.GetEntitiesCount(); count != 0 {
		t.Errorf("invalid fcount: %d", count)
	}
	if count := entites[offset+ecs.RawEntityOffsetComponentsCount]; count != 2 {
		t.Errorf("invalid ccount: %d", count)
	}
	if c1 := entites[offset+ecs.RawEntityOffsetComponents]; c1 != p1.GetID() {
		t.Errorf("invalid c1: %d", c1)
	}
	if c2 := entites[offset+ecs.RawEntityOffsetComponents+1]; c2 != p2.GetID() {
		t.Errorf("invalid c1: %d", c2)
	}
	w.Destroy()
}

func TestPoolsOverflow(t *testing.T) {
	w := ecs.NewWorld()
	defer func(world *ecs.World) {
		if r := recover(); r == nil {
			t.Errorf("code should panic")
		}
		world.Destroy()
	}(w)
	var pools []ecs.IPool
	for i := 0; i < math.MaxInt16; i++ {
		pools = append(pools, nil)
	}
	*ecs.DebugGetPoolsPtr(w) = pools
	_ = ecs.GetPool[C1](w)
	t.Errorf("code should panic")
}

func TestWorldEmptyEntity(t *testing.T) {
	w := ecs.NewWorld()
	defer func(world *ecs.World) {
		if r := recover(); r == nil {
			t.Errorf("code should panic")
		}
		world.Destroy()
	}(w)
	w.NewEntity()
	w.Destroy()
	t.Errorf("code should panic")
}

func TestWorldDelInvalidEntity(t *testing.T) {
	w := ecs.NewWorldWithConfig(ecs.WorldConfig{WorldEntitiesSize: 2})
	defer func(world *ecs.World) {
		if r := recover(); r == nil {
			t.Errorf("code should panic")
		}
		world.Destroy()
	}(w)
	w.DelEntity(3)
	t.Errorf("code should panic")
}

func TestWorldDelEntityTwice(t *testing.T) {
	w := ecs.NewWorld()
	e := w.NewEntity()
	w.DelEntity(e)
	w.DelEntity(e)
	w.Destroy()
}

func TestWorldGenEntityOverflow(t *testing.T) {
	w := ecs.NewWorld()
	for i := 0; i < math.MaxInt16; i++ {
		e := w.NewEntity()
		w.DelEntity(e)
	}
	e := w.NewEntity()
	gen := w.GetEntityGen(e)
	if gen != 1 {
		t.Errorf("invalid entity gen on overflow: %d.", gen)
	}
	w.DelEntity(e)
	w.Destroy()
}

type worldEventListener struct{}

func (l *worldEventListener) OnEntityCreated(entity int)        {}
func (l *worldEventListener) OnEntityChanged(entity int)        {}
func (l *worldEventListener) OnEntityDestroyed(entity int)      {}
func (l *worldEventListener) OnWorldResized(newSize int)        {}
func (l *worldEventListener) OnWorldDestroyed(world *ecs.World) {}

func TestWorldDebugListeners(t *testing.T) {
	w := ecs.NewWorldWithConfig(ecs.WorldConfig{WorldEntitiesSize: 2})
	listener := worldEventListener{}
	w.AddEventListener(&listener)
	p := ecs.GetPool[C1](w)
	_ = ecs.GetFilter[ecs.Inc1[C1]](w)
	for i := 0; i < 3; i++ {
		p.Add(w.NewEntity())
	}
	w.Destroy()
	w.RemoveEventListener(&listener)
}

func TestWorldGetComponentsInfo(t *testing.T) {
	w := ecs.NewWorld()
	p := ecs.GetPool[C1](w)
	f := ecs.GetFilter[ecs.Inc1[C1]](w)
	for i := 0; i < 3; i++ {
		p.Add(w.NewEntity())
	}
	var typesList []reflect.Type
	var valuesList []any
	for it := f.Iter(); it.Next(); {
		typesList = w.GetComponentTypes(it.GetEntity(), typesList)
		if len(typesList) != 1 {
			t.Errorf("invalid component types list")
		}
		typesList = typesList[:0]
		valuesList = w.GetComponentValues(it.GetEntity(), valuesList)
		if len(valuesList) != 1 {
			t.Errorf("invalid component values list")
		}
		valuesList = valuesList[:0]
	}
	w.Destroy()
}

func TestCopyEntity(t *testing.T) {
	w := ecs.NewWorld()
	p := ecs.GetPool[C2](w)
	srcE := w.NewEntity()
	dstE := w.NewEntity()
	srcC := p.Add(srcE)
	srcC.ID = 123
	w.CopyEntity(srcE, dstE)
	dstC := p.Get(dstE)
	if srcC.ID*2 != dstC.ID {
		t.Errorf("invalid component data")
	}
	w.Destroy()
}
