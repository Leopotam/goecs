// ----------------------------------------------------------------------------
// The Proprietary or MIT-Red License
// Copyright (c) 2012-2022 Leopotam <leopotam@yandex.ru>
// ----------------------------------------------------------------------------

package goecs_test

import (
	"testing"

	"github.com/leopotam/goecs"
)

func TestQueryWithOneInc(t *testing.T) {
	w := goecs.NewWorld()
	e1 := w.NewEntity()
	e2 := w.NewEntity()
	p := goecs.GetPool[C1](w)
	p.Add(e1)
	p.Add(e2)
	q := goecs.NewQuery[goecs.Inc1[C1]](w)
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
	w := goecs.NewWorld()
	e1 := w.NewEntity()
	e2 := w.NewEntity()
	p1 := goecs.GetPool[C1](w)
	p2 := goecs.GetPool[C2](w)
	p1.Add(e1)
	p2.Add(e2)
	q := goecs.NewQuery[goecs.Inc2[C1, C2]](w)
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
	w := goecs.NewWorld()
	e := w.NewEntity()
	p1 := goecs.GetPool[C1](w)
	p2 := goecs.GetPool[C2](w)
	p1.Add(e)
	p2.Add(e)
	q1 := goecs.NewQuery[goecs.Inc1[C1]](w)
	i := 0
	for it := q1.Iter(); it.Next(); {
		i++
	}
	if i != 1 {
		t.Errorf("invalid entities count in query.")
	}

	q2 := goecs.NewQueryWithExc[goecs.Inc1[C1], goecs.Exc1[C2]](w)
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
	w := goecs.NewWorld()
	goecs.NewQuery[goecs.Inc1[C1]](w)
	goecs.NewQuery[goecs.Inc2[C1, C2]](w)
	goecs.NewQuery[goecs.Inc3[C1, C2, C3]](w)
	goecs.NewQuery[goecs.Inc4[C1, C2, C3, C4]](w)
	goecs.NewQueryWithExc[goecs.Inc1[C1], goecs.Exc1[C2]](w)
	goecs.NewQueryWithExc[goecs.Inc1[C1], goecs.Exc2[C2, C3]](w)
	w.Destroy()
}

func TestQueryIter(t *testing.T) {
	w := goecs.NewWorld()
	e1 := w.NewEntity()
	e2 := w.NewEntity()
	e3 := w.NewEntity()
	c1Pool := goecs.GetPool[C1](w)
	c2Pool := goecs.GetPool[C2](w)
	c3Pool := goecs.GetPool[C3](w)
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
	q1 := goecs.NewQuery[goecs.Inc1[C1]](w)
	i := 0
	for it := q1.Iter(); it.Next(); {
		_ = it.GetEntity()
		i++
	}
	if i != 2 {
		t.Errorf("invalid entities count: %d", i)
	}

	q2 := goecs.NewQueryWithExc[goecs.Inc1[C1], goecs.Exc1[C2]](w)
	i = 0
	for it := q2.Iter(); it.Next(); {
		_ = it.GetEntity()
		i++
	}
	if i != 0 {
		t.Errorf("invalid entities count: %d", i)
	}

	q3 := goecs.NewQueryWithExc[goecs.Inc2[C1, C2], goecs.Exc1[C4]](w)
	i = 0
	for it := q3.Iter(); it.Next(); {
		_ = it.GetEntity()
		i++
	}
	if i != 2 {
		t.Errorf("invalid entities count: %d", i)
	}

	q4 := goecs.NewQueryWithExc[goecs.Inc2[C1, C3], goecs.Exc1[C4]](w)
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

func TestInvalidIterNext(t *testing.T) {
	w := goecs.NewWorld()
	defer func(world *goecs.World) {
		if r := recover(); r == nil {
			t.Errorf("code should panic.")
		}
		world.Destroy()
	}(w)
	// p := goecs.GetPool[C2](w)
	// p.Get(0)
	q := goecs.NewQuery[goecs.Inc1[C1]](w)
	it := q.Iter()
	it.Next()
	it.Next()
	t.Errorf("code should panic.")
}

func TestInvalidIterWithExcNext(t *testing.T) {
	w := goecs.NewWorld()
	defer func(world *goecs.World) {
		if r := recover(); r == nil {
			t.Errorf("code should panic.")
		}
		world.Destroy()
	}(w)
	q := goecs.NewQueryWithExc[goecs.Inc1[C1], goecs.Exc1[C2]](w)
	it := q.Iter()
	it.Next()
	it.Next()
	t.Errorf("code should panic.")
}

func BenchmarkQueryWithOneEmptyInc(b *testing.B) {
	w := goecs.NewWorld()
	p := goecs.GetPool[C1](w)
	for i := 0; i < 100000; i++ {
		p.Add(w.NewEntity())
	}
	q := goecs.NewQuery[goecs.Inc1[C1]](w)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for it := q.Iter(); it.Next(); {
			_ = p.Get(it.GetEntity())
		}
	}
	b.StopTimer()
	w.Destroy()
}

func BenchmarkQueryWithOneNonEmptyInc(b *testing.B) {
	w := goecs.NewWorld()
	p := goecs.GetPool[C2](w)
	for i := 0; i < 100000; i++ {
		p.Add(w.NewEntity())
	}
	q := goecs.NewQuery[goecs.Inc1[C2]](w)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for it := q.Iter(); it.Next(); {
			_ = p.Get(it.GetEntity())
		}
	}
	b.StopTimer()
	w.Destroy()
}

func BenchmarkQueryWithTwoInc(b *testing.B) {
	w := goecs.NewWorld()
	p1 := goecs.GetPool[C1](w)
	p2 := goecs.GetPool[C2](w)
	for i := 0; i < 100000; i++ {
		e := w.NewEntity()
		p1.Add(e)
		p2.Add(e)
	}
	q := goecs.NewQuery[goecs.Inc2[C1, C2]](w)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for it := q.Iter(); it.Next(); {
			_ = p1.Get(it.GetEntity())
		}
	}
	b.StopTimer()
	w.Destroy()
}

func BenchmarkQueryWithTreeInc(b *testing.B) {
	w := goecs.NewWorld()
	p1 := goecs.GetPool[C1](w)
	p2 := goecs.GetPool[C2](w)
	p3 := goecs.GetPool[C3](w)
	for i := 0; i < 100000; i++ {
		e := w.NewEntity()
		p1.Add(e)
		p2.Add(e)
		p3.Add(e)
	}
	q := goecs.NewQuery[goecs.Inc3[C1, C2, C3]](w)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for it := q.Iter(); it.Next(); {
			_ = p1.Get(it.GetEntity())
		}
	}
	b.StopTimer()
	w.Destroy()
}

func BenchmarkQueryWithFourInc(b *testing.B) {
	w := goecs.NewWorld()
	p1 := goecs.GetPool[C1](w)
	p2 := goecs.GetPool[C2](w)
	p3 := goecs.GetPool[C3](w)
	p4 := goecs.GetPool[C4](w)
	for i := 0; i < 100000; i++ {
		e := w.NewEntity()
		p1.Add(e)
		p2.Add(e)
		p3.Add(e)
		p4.Add(e)
	}
	q := goecs.NewQuery[goecs.Inc4[C1, C2, C3, C4]](w)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for it := q.Iter(); it.Next(); {
			_ = p1.Get(it.GetEntity())
		}
	}
	b.StopTimer()
	w.Destroy()
}

func BenchmarkQueryWithTwoIncAndOneExc(b *testing.B) {
	w := goecs.NewWorld()
	p1 := goecs.GetPool[C1](w)
	p2 := goecs.GetPool[C2](w)
	for i := 0; i < 100000; i++ {
		e := w.NewEntity()
		p1.Add(e)
		p2.Add(e)
	}
	q := goecs.NewQueryWithExc[goecs.Inc2[C1, C2], goecs.Exc1[C3]](w)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for it := q.Iter(); it.Next(); {
			_ = p1.Get(it.GetEntity())
		}
	}
	b.StopTimer()
	w.Destroy()
}
