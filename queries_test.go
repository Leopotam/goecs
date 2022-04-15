// ----------------------------------------------------------------------------
// The Proprietary or MIT-Red License
// Copyright (c) 2012-2022 Leopotam <leopotam@yandex.ru>
// ----------------------------------------------------------------------------

package ecs_test

import (
	"ecs"
	"testing"
)

func TestQueryWithOneInc(t *testing.T) {
	w := ecs.NewWorld()
	e1 := w.NewEntity()
	e2 := w.NewEntity()
	p := ecs.GetPool[C1](w)
	p.Add(e1)
	p.Add(e2)
	q := ecs.NewQuery[ecs.Inc1[C1]](w)
	i := 0
	for it := q.Iter(); it.Next(); {
		i++
	}
	if i != 2 {
		t.Errorf("invalid entities count in query.")
	}
	w.Destroy()
}

func TestQueryWithTwoInc(t *testing.T) {
	w := ecs.NewWorld()
	e1 := w.NewEntity()
	e2 := w.NewEntity()
	p1 := ecs.GetPool[C1](w)
	p2 := ecs.GetPool[C2](w)
	p1.Add(e1)
	p2.Add(e2)
	q := ecs.NewQuery[ecs.Inc2[C1, C2]](w)
	i := 0
	for it := q.Iter(); it.Next(); {
		i++
	}
	if i != 0 {
		t.Errorf("invalid entities count in query.")
	}

	p1.Add(e2)
	i = 0
	for it := q.Iter(); it.Next(); {
		i++
	}
	if i != 1 {
		t.Errorf("invalid entities count in query.")
	}

	p2.Add(e1)
	i = 0
	for it := q.Iter(); it.Next(); {
		i++
	}
	if i != 2 {
		t.Errorf("invalid entities count in query.")
	}

	w.Destroy()
}

func TestQueryWithOneIncOneExc(t *testing.T) {
	w := ecs.NewWorld()
	e := w.NewEntity()
	p1 := ecs.GetPool[C1](w)
	p2 := ecs.GetPool[C2](w)
	p1.Add(e)
	p2.Add(e)
	q1 := ecs.NewQuery[ecs.Inc1[C1]](w)
	i := 0
	for it := q1.Iter(); it.Next(); {
		i++
	}
	if i != 1 {
		t.Errorf("invalid entities count in query.")
	}

	q2 := ecs.NewQueryWithExc[ecs.Inc1[C1], ecs.Exc1[C2]](w)
	i = 0
	for it := q2.Iter(); it.Next(); {
		i++
	}
	if i != 0 {
		t.Errorf("invalid entities count in query.")
	}
	w.Destroy()
}

func TestQueryWithLongIncAndLongExc(t *testing.T) {
	w := ecs.NewWorld()
	ecs.NewQuery[ecs.Inc1[C1]](w)
	ecs.NewQuery[ecs.Inc2[C1, C2]](w)
	ecs.NewQuery[ecs.Inc3[C1, C2, C3]](w)
	ecs.NewQuery[ecs.Inc4[C1, C2, C3, C4]](w)
	ecs.NewQueryWithExc[ecs.Inc1[C1], ecs.Exc1[C2]](w)
	ecs.NewQueryWithExc[ecs.Inc1[C1], ecs.Exc2[C2, C3]](w)
	w.Destroy()
}

func TestQueryIter(t *testing.T) {
	w := ecs.NewWorld()
	e1 := w.NewEntity()
	e2 := w.NewEntity()
	e3 := w.NewEntity()
	c1Pool := ecs.GetPool[C1](w)
	c2Pool := ecs.GetPool[C2](w)
	c3Pool := ecs.GetPool[C3](w)
	c1Pool.Add(e1)
	c1Pool.Add(e2)
	c1Pool.Add(e3)
	c2Pool.Add(e1)
	c2Pool.Add(e2)
	c2Pool.Add(e3)
	w.DelEntity(e2)
	c3Pool.Add(w.NewEntity())
	c3Pool.Add(w.NewEntity())
	c3Pool.Add(w.NewEntity())
	q1 := ecs.NewQuery[ecs.Inc1[C1]](w)
	i := 0
	for it := q1.Iter(); it.Next(); {
		_ = it.GetEntity()
		i++
	}
	if i != 2 {
		t.Errorf("invalid entities count: %d", i)
	}

	q2 := ecs.NewQueryWithExc[ecs.Inc1[C1], ecs.Exc1[C2]](w)
	i = 0
	for it := q2.Iter(); it.Next(); {
		_ = it.GetEntity()
		i++
	}
	if i != 0 {
		t.Errorf("invalid entities count: %d", i)
	}

	q3 := ecs.NewQueryWithExc[ecs.Inc2[C1, C2], ecs.Exc1[C4]](w)
	i = 0
	for it := q3.Iter(); it.Next(); {
		_ = it.GetEntity()
		i++
	}
	if i != 2 {
		t.Errorf("invalid entities count: %d", i)
	}

	q4 := ecs.NewQueryWithExc[ecs.Inc2[C1, C3], ecs.Exc1[C4]](w)
	i = 0
	for it := q4.Iter(); it.Next(); {
		_ = it.GetEntity()
		i++
	}
	if i != 0 {
		t.Errorf("invalid entities count: %d", i)
	}

	w.Destroy()
}
