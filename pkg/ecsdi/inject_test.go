// ----------------------------------------------------------------------------
// The Proprietary or MIT-Red License
// Copyright (c) 2012-2022 Leopotam <leopotam@yandex.ru>
// ----------------------------------------------------------------------------

package ecsdi_test

import (
	"testing"

	"leopotam.com/go/ecs"
	"leopotam.com/go/ecs/pkg/ecsdi"
)

type customData struct {
	ID int
}

type c1 struct{}
type c2 struct{}

var _ c1
var _ c2

type worldSystem1 struct {
	World       ecsdi.World
	EventsWorld ecsdi.World `ecsdi:"events"`
}

type poolSystem1 struct {
	C1Pool ecsdi.Pool[c1]
}
type poolSystem2 struct {
	C1Pool       ecsdi.Pool[c1]
	EventsC1Pool ecsdi.Pool[c1] `ecsdi:"events"`
}

type filterSystem1 struct {
	C1Filter ecsdi.Filter[ecs.Inc1[c1]]
}
type filterSystem2 struct {
	C1WithoutC2Filter ecsdi.FilterWithExc[ecs.Inc1[c1], ecs.Exc1[c2]]
}
type filterSystem3 struct {
	C1Filter       ecsdi.Filter[ecs.Inc1[c1]]
	EventsC1Filter ecsdi.Filter[ecs.Inc1[c1]] `ecsdi:"events"`
}
type filterSystem4 struct {
	EventsC1WithoutC2Filter ecsdi.FilterWithExc[ecs.Inc1[c1], ecs.Exc1[c2]] `ecsdi:"events"`
}

type customSystem1 struct {
	Data ecsdi.Custom[customData]
}

func (ws *worldSystem1) Init(s *ecs.Systems) {}

func (ps *poolSystem1) Init(s *ecs.Systems) {}
func (ps *poolSystem2) Init(s *ecs.Systems) {}

func (qs *filterSystem1) Init(s *ecs.Systems) {}
func (qs *filterSystem2) Init(s *ecs.Systems) {}
func (qs *filterSystem3) Init(s *ecs.Systems) {}
func (qs *filterSystem4) Init(s *ecs.Systems) {}

func (cs *customSystem1) Init(s *ecs.Systems) {}

func TestInjectDefaultWorld(t *testing.T) {
	w := ecs.NewWorld()
	s := ecs.NewSystems(w)
	sys := worldSystem1{}
	s.Add(&sys)
	ecsdi.Inject(s).Init()
	if w != sys.World.Value {
		t.Errorf("invalid world inject.")
	}
	s.Destroy()
	w.Destroy()
}

func TestInjectCustomWorld(t *testing.T) {
	w1 := ecs.NewWorld()
	w2 := ecs.NewWorld()
	s := ecs.NewSystems(w1)
	sys := worldSystem1{}
	s.AddWorld(w2, "events").Add(&sys)
	ecsdi.Inject(s).Init()
	if w2 != sys.EventsWorld.Value {
		t.Errorf("invalid custom world inject.")
	}
	s.Destroy()
	w1.Destroy()
	w2.Destroy()
}

func TestInjectPool(t *testing.T) {
	w := ecs.NewWorld()
	s := ecs.NewSystems(w)
	sys := poolSystem1{}
	c1Pool := ecs.GetPool[c1](w)
	s.Add(&sys)
	ecsdi.Inject(s).Init()
	if c1Pool != sys.C1Pool.Value {
		t.Errorf("invalid pool inject.")
	}
	s.Destroy()
	w.Destroy()
}

func TestInjectPoolNewEntity(t *testing.T) {
	w := ecs.NewWorld()
	s := ecs.NewSystems(w)
	sys := poolSystem1{}
	s.Add(&sys)
	ecsdi.Inject(s).Init()
	c1, e := sys.C1Pool.NewEntity()
	if c1 == nil {
		t.Errorf("invalid pool inject.")
	}
	w.DelEntity(e)
	s.Destroy()
	w.Destroy()
}

func TestInjectPoolFromCustomWorld(t *testing.T) {
	w1 := ecs.NewWorld()
	w2 := ecs.NewWorld()
	s := ecs.NewSystems(w1)
	sys := poolSystem2{}
	c1Pool := ecs.GetPool[c1](w2)
	s.AddWorld(w2, "events").Add(&sys)
	ecsdi.Inject(s).Init()
	if c1Pool != sys.EventsC1Pool.Value {
		t.Errorf("invalid pool from custom world inject.")
	}
	s.Destroy()
	w1.Destroy()
}

func TestInvalidPoolFromUndefinedWorld(t *testing.T) {
	w := ecs.NewWorld()
	s := ecs.NewSystems(w)
	sys := poolSystem2{}
	defer func(world *ecs.World, systems *ecs.Systems) {
		if r := recover(); r == nil {
			t.Errorf("code should panic.")
		}
		systems.Destroy()
		world.Destroy()
	}(w, s)
	s.Add(&sys)
	ecsdi.Inject(s).Init()
	t.Errorf("code should panic.")
}

func TestInjectFilter(t *testing.T) {
	w := ecs.NewWorld()
	s := ecs.NewSystems(w)
	sys := filterSystem1{}
	s.Add(&sys)
	ecsdi.Inject(s).Init()
	if sys.C1Filter.Value == nil {
		t.Errorf("invalid filter inject.")
	}
	s.Destroy()
	w.Destroy()
}

func TestInjectFilterWithExc(t *testing.T) {
	w := ecs.NewWorld()
	s := ecs.NewSystems(w)
	sys := filterSystem2{}
	s.Add(&sys)
	ecsdi.Inject(s).Init()
	if sys.C1WithoutC2Filter.Value == nil {
		t.Errorf("invalid filter inject.")
	}
	s.Destroy()
	w.Destroy()
}

func TestInjectFilterFromCustomWorld(t *testing.T) {
	w1 := ecs.NewWorld()
	w2 := ecs.NewWorld()
	s := ecs.NewSystems(w1)
	sys := filterSystem3{}
	s.AddWorld(w2, "events").Add(&sys)
	ecsdi.Inject(s).Init()
	if sys.EventsC1Filter.Value == nil {
		t.Errorf("invalid filter inject.")
	}
	s.Destroy()
	w1.Destroy()
	w2.Destroy()
}

func TestInjectFilterPools(t *testing.T) {
	w := ecs.NewWorld()
	p := ecs.GetPool[c1](w)
	s := ecs.NewSystems(w)
	sys := filterSystem1{}
	s.Add(&sys)
	ecsdi.Inject(s).Init()
	if sys.C1Filter.Pools == nil || sys.C1Filter.Pools.Inc1 != p {
		t.Errorf("invalid filter pools inject.")
	}
	s.Destroy()
	w.Destroy()
}

func TestInjectFilterWithExcPools(t *testing.T) {
	w := ecs.NewWorld()
	p := ecs.GetPool[c1](w)
	s := ecs.NewSystems(w)
	sys := filterSystem2{}
	s.Add(&sys)
	ecsdi.Inject(s).Init()
	if sys.C1WithoutC2Filter.Pools == nil || sys.C1WithoutC2Filter.Pools.Inc1 != p {
		t.Errorf("invalid filter pools inject.")
	}
	s.Destroy()
	w.Destroy()
}

func TestInvalidFilterFromUndefinedWorld(t *testing.T) {
	w := ecs.NewWorld()
	s := ecs.NewSystems(w)
	sys := filterSystem3{}
	defer func(world *ecs.World, systems *ecs.Systems) {
		if r := recover(); r == nil {
			t.Errorf("code should panic.")
		}
		systems.Destroy()
		world.Destroy()
	}(w, s)
	s.Add(&sys)
	ecsdi.Inject(s).Init()
	t.Errorf("code should panic.")
}

func TestInvalidFilterWithExcFromUndefinedWorld(t *testing.T) {
	w := ecs.NewWorld()
	s := ecs.NewSystems(w)
	sys := filterSystem4{}
	defer func(world *ecs.World, systems *ecs.Systems) {
		if r := recover(); r == nil {
			t.Errorf("code should panic.")
		}
		systems.Destroy()
		world.Destroy()
	}(w, s)
	s.Add(&sys)
	ecsdi.Inject(s).Init()
	t.Errorf("code should panic.")
}

func TestInjectCustomData(t *testing.T) {
	w := ecs.NewWorld()
	s := ecs.NewSystems(w)
	ts1 := customSystem1{}
	cd := &customData{ID: 1}
	s.Add(&ts1)
	ecsdi.Inject(s, cd).Init()
	if cd != ts1.Data.Value || cd.ID != ts1.Data.Value.ID {
		t.Errorf("invalid custom data inject.")
	}
	s.Destroy()
	w.Destroy()
}
