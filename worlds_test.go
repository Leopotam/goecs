// ----------------------------------------------------------------------------
// The Proprietary or MIT-Red License
// Copyright (c) 2012-2022 Leopotam <leopotam@yandex.ru>
// ----------------------------------------------------------------------------

package goecs_test

import (
	"math"
	"testing"

	"github.com/leopotam/goecs"
)

type C1 struct{}
type C2 struct{ ID int }
type C3 struct{}
type C4 struct{}
type C5 struct{}

func (c2 *C2) Reset() {
	c2.ID = -1
}

func TestWorldCreate(t *testing.T) {
	w := goecs.NewWorld()
	w.Destroy()
}

func TestWorldResize(t *testing.T) {
	w := goecs.NewWorldWithConfig(goecs.WorldConfig{WorldEntitiesSize: 2})
	_ = goecs.GetPool[C1](w)
	entities := make([]int, 0, 4)
	for i := 0; i < 3; i++ {
		entities = append(entities, w.NewEntity())
	}
	for _, entity := range entities {
		w.DelEntity(entity)
	}
	w.Destroy()
}

func TestWorldEmptyEntity(t *testing.T) {
	w := goecs.NewWorld()
	defer func(world *goecs.World) {
		if r := recover(); r == nil {
			t.Errorf("code should panic.")
		}
		world.Destroy()
	}(w)
	w.NewEntity()
	w.Destroy()
	t.Errorf("code should panic.")
}

func TestWorldDelInvalidEntity(t *testing.T) {
	w := goecs.NewWorldWithConfig(goecs.WorldConfig{WorldEntitiesSize: 2})
	defer func(world *goecs.World) {
		if r := recover(); r == nil {
			t.Errorf("code should panic.")
		}
		world.Destroy()
	}(w)
	w.DelEntity(3)
	t.Errorf("code should panic.")
}

func TestWorldDelEntityTwice(t *testing.T) {
	w := goecs.NewWorld()
	e := w.NewEntity()
	w.DelEntity(e)
	w.DelEntity(e)
	w.Destroy()
}

func TestWorldGenEntityOverflow(t *testing.T) {
	w := goecs.NewWorld()
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
