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
		t.Fatalf("invalid entities count in query.")
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
		t.Fatalf("invalid entities count in query.")
	}

	p1.Add(e2)
	i = 0
	for it := q.Iter(); it.Next(); {
		i++
	}
	if i != 1 {
		t.Fatalf("invalid entities count in query.")
	}

	p2.Add(e1)
	i = 0
	for it := q.Iter(); it.Next(); {
		i++
	}
	if i != 2 {
		t.Fatalf("invalid entities count in query.")
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
		t.Fatalf("invalid entities count in query.")
	}

	q2 := ecs.NewQueryWithExc[ecs.Inc1[C1], ecs.Exc1[C2]](w)
	i = 0
	for it := q2.Iter(); it.Next(); {
		i++
	}
	if i != 0 {
		t.Fatalf("invalid entities count in query.")
	}
	w.Destroy()
}
