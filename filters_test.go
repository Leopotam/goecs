// ----------------------------------------------------------------------------
// The Proprietary or MIT-Red License
// Copyright (c) 2012-2022 Leopotam <leopotam@yandex.ru>
// ----------------------------------------------------------------------------

package ecs_test

import (
	"testing"

	"leopotam.com/go/ecs"
)

func TestFilterApi(t *testing.T) {
	w := ecs.NewWorldWithConfig(ecs.WorldConfig{WorldEntitiesSize: 1234})
	f1 := ecs.GetFilter[ecs.Inc1[C1]](w)
	ecs.GetPool[C1](w).Add(w.NewEntity())
	if f1.GetWorld() != w {
		t.Errorf("invalid world")
	}
	if len(f1.GetSparseIndices()) != 1234 {
		t.Errorf("invalid sparse index")
	}
	if len(f1.GetRawEntities()) != 1 {
		t.Errorf("invalid dense index")
	}
	w.Destroy()
}

func TestFilterWithOneInc(t *testing.T) {
	w := ecs.NewWorld()
	e1 := w.NewEntity()
	e2 := w.NewEntity()
	p := ecs.GetPool[C1](w)
	p.Add(e1)
	p.Add(e2)
	f := ecs.GetFilter[ecs.Inc1[C1]](w)
	i := 0
	for it := f.Iter(); it.Next(); {
		i++
	}
	if i != 2 {
		t.Errorf("invalid entities count in filter")
	}
	w.Destroy()
}

func TestFilterWithOneIncMultipleTimes(t *testing.T) {
	w := ecs.NewWorld()
	fi1 := ecs.GetFilter[ecs.Inc1[C1]](w)
	fi2 := ecs.GetFilter[ecs.Inc1[C1]](w)
	fi3 := ecs.GetFilter[ecs.Inc1[C1]](w)
	if fi1 != fi2 || fi1 != fi3 {
		t.Errorf("filters not equals")
	}
	fe1 := ecs.GetFilterWithExc[ecs.Inc1[C1], ecs.Exc1[C2]](w)
	fe2 := ecs.GetFilterWithExc[ecs.Inc1[C1], ecs.Exc1[C2]](w)
	fe3 := ecs.GetFilterWithExc[ecs.Inc1[C1], ecs.Exc1[C2]](w)
	if fe1 != fe2 || fe1 != fe3 {
		t.Errorf("filters not equals")
	}
	w.Destroy()
}

func TestFilterConstraintFillPools(t *testing.T) {
	w := ecs.NewWorld()
	var i1 ecs.Inc1[C1]
	var i2 ecs.Inc2[C1, C2]
	var i3 ecs.Inc3[C1, C2, C3]
	var i4 ecs.Inc4[C1, C2, C3, C4]
	var i5 ecs.Inc5[C1, C2, C3, C4, C5]
	var i6 ecs.Inc6[C1, C2, C3, C4, C5, C6]
	i11 := i1.FillPools(w).(*ecs.Inc1[C1])
	i21 := i2.FillPools(w).(*ecs.Inc2[C1, C2])
	i31 := i3.FillPools(w).(*ecs.Inc3[C1, C2, C3])
	i41 := i4.FillPools(w).(*ecs.Inc4[C1, C2, C3, C4])
	i51 := i5.FillPools(w).(*ecs.Inc5[C1, C2, C3, C4, C5])
	i61 := i6.FillPools(w).(*ecs.Inc6[C1, C2, C3, C4, C5, C6])

	if i61.Inc1 == nil ||
		i61.Inc2 == nil ||
		i61.Inc3 == nil ||
		i61.Inc4 == nil ||
		i61.Inc5 == nil ||
		i61.Inc6 == nil ||
		i61.Inc1 != i11.Inc1 ||
		i61.Inc2 != i21.Inc2 ||
		i61.Inc3 != i31.Inc3 ||
		i61.Inc4 != i41.Inc4 ||
		i61.Inc5 != i51.Inc5 ||
		i61.Inc6 != ecs.GetPool[C6](w) {
		t.Errorf("pools not filled")
	}
	w.Destroy()
}

func TestFilterWithOneIncNested(t *testing.T) {
	w := ecs.NewWorld()
	e1 := w.NewEntity()
	e2 := w.NewEntity()
	p := ecs.GetPool[C1](w)
	p.Add(e1)
	p.Add(e2)
	f := ecs.GetFilter[ecs.Inc1[C1]](w)
	i := 0
	for it := f.Iter(); it.Next(); {
		for it2 := f.Iter(); it2.Next(); {
			p.Add(w.NewEntity())
			i++
		}
	}
	for it := f.Iter(); it.Next(); {
		w.DelEntity(it.GetEntity())
	}
	if i != 4 {
		t.Errorf("invalid entities count in filter")
	}
	w.Destroy()
}

func TestFilterWithTwoInc(t *testing.T) {
	w := ecs.NewWorld()
	e1 := w.NewEntity()
	e2 := w.NewEntity()
	p1 := ecs.GetPool[C1](w)
	p2 := ecs.GetPool[C2](w)
	p1.Add(e1)
	p2.Add(e2)
	f := ecs.GetFilter[ecs.Inc2[C1, C2]](w)
	i := 0
	for it := f.Iter(); it.Next(); {
		i++
	}
	if i != 0 {
		t.Errorf("invalid entities count in filter")
	}

	p1.Add(e2)
	i = 0
	for it := f.Iter(); it.Next(); {
		i++
	}
	if i != 1 {
		t.Errorf("invalid entities count in filter")
	}

	p2.Add(e1)
	i = 0
	for it := f.Iter(); it.Next(); {
		i++
	}
	if i != 2 {
		t.Errorf("invalid entities count in filter")
	}

	w.Destroy()
}

func TestFilterWithOneIncOneExc(t *testing.T) {
	w := ecs.NewWorld()
	p1 := ecs.GetPool[C1](w)
	p2 := ecs.GetPool[C2](w)
	f1 := ecs.GetFilter[ecs.Inc1[C1]](w)
	f2 := ecs.GetFilterWithExc[ecs.Inc1[C1], ecs.Exc1[C2]](w)
	e := w.NewEntity()
	p1.Add(e)
	p2.Add(e)
	if f1.GetEntitiesCount() != 1 {
		t.Errorf("invalid entities count in filter")
	}
	i := 0
	for it := f1.Iter(); it.Next(); {
		i++
	}
	if i != 1 {
		t.Errorf("invalid iterations count in filter")
	}

	if f2.GetEntitiesCount() != 0 {
		t.Errorf("invalid entities count in filter")
	}
	i = 0
	for it := f2.Iter(); it.Next(); {
		i++
	}
	if i != 0 {
		t.Errorf("invalid iterations count in filter")
	}
	w.Destroy()
}

func TestFilterWithLongIncAndLongExc(t *testing.T) {
	w := ecs.NewWorld()
	ecs.GetFilter[ecs.Inc1[C1]](w)
	ecs.GetFilter[ecs.Inc2[C1, C2]](w)
	ecs.GetFilter[ecs.Inc3[C1, C2, C3]](w)
	ecs.GetFilter[ecs.Inc4[C1, C2, C3, C4]](w)
	ecs.GetFilter[ecs.Inc5[C1, C2, C3, C4, C5]](w)
	ecs.GetFilter[ecs.Inc6[C1, C2, C3, C4, C5, C6]](w)
	ecs.GetFilterWithExc[ecs.Inc1[C1], ecs.Exc1[C2]](w)
	ecs.GetFilterWithExc[ecs.Inc1[C1], ecs.Exc2[C2, C3]](w)
	ecs.GetFilterWithExc[ecs.Inc1[C1], ecs.Exc3[C2, C3, C4]](w)
	w.Destroy()
}

func TestFilterIter(t *testing.T) {
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
	f1 := ecs.GetFilter[ecs.Inc1[C1]](w)
	i := 0
	for it := f1.Iter(); it.Next(); {
		_ = it.GetEntity()
		i++
	}
	if i != 2 {
		t.Errorf("invalid entities count: %d", i)
	}

	f2 := ecs.GetFilterWithExc[ecs.Inc1[C1], ecs.Exc1[C2]](w)
	i = 0
	for it := f2.Iter(); it.Next(); {
		_ = it.GetEntity()
		i++
	}
	if i != 0 {
		t.Errorf("invalid entities count: %d", i)
	}

	f3 := ecs.GetFilterWithExc[ecs.Inc2[C1, C2], ecs.Exc1[C4]](w)
	i = 0
	for it := f3.Iter(); it.Next(); {
		_ = it.GetEntity()
		i++
	}
	if i != 2 {
		t.Errorf("invalid entities count: %d", i)
	}

	f4 := ecs.GetFilterWithExc[ecs.Inc2[C1, C3], ecs.Exc1[C4]](w)
	i = 0
	for it := f4.Iter(); it.Next(); {
		_ = it.GetEntity()
		i++
	}
	if i != 0 {
		t.Errorf("invalid entities count: %d", i)
	}

	w.Destroy()
}

func TestFilterDelExcComponent(t *testing.T) {
	w := ecs.NewWorld()
	_ = ecs.GetFilterWithExc[ecs.Inc1[C1], ecs.Exc1[C2]](w)
	p1 := ecs.GetPool[C1](w)
	p2 := ecs.GetPool[C2](w)
	e := w.NewEntity()
	p1.Add(e)
	p2.Add(e)
	p2.Del(e)
	w.Destroy()
}

func TestFilterDelComplexExcComponent(t *testing.T) {
	w := ecs.NewWorld()
	_ = ecs.GetFilterWithExc[ecs.Inc1[C1], ecs.Exc2[C2, C3]](w)
	p1 := ecs.GetPool[C1](w)
	p2 := ecs.GetPool[C2](w)
	p3 := ecs.GetPool[C3](w)
	e := w.NewEntity()
	p1.Add(e)
	p2.Add(e)
	p3.Add(e)
	p2.Del(e)
	w.Destroy()
}

func TestFilterInvalidIterNext(t *testing.T) {
	w := ecs.NewWorld()
	defer func(world *ecs.World) {
		if r := recover(); r == nil {
			t.Errorf("code should panic")
		}
		world.Destroy()
	}(w)
	f := ecs.GetFilter[ecs.Inc1[C1]](w)
	it := f.Iter()
	it.Next()
	it.Next()
	t.Errorf("code should panic")
}

func TestFilterInvalidIterWithExcNext(t *testing.T) {
	w := ecs.NewWorld()
	defer func(world *ecs.World) {
		if r := recover(); r == nil {
			t.Errorf("code should panic")
		}
		world.Destroy()
	}(w)
	f := ecs.GetFilterWithExc[ecs.Inc1[C1], ecs.Exc1[C2]](w)
	it := f.Iter()
	it.Next()
	it.Next()
	t.Errorf("code should panic")
}

func TestFilterInvalidDelAddRelatedComponent(t *testing.T) {
	w := ecs.NewWorld()
	defer func(world *ecs.World) {
		if r := recover(); r == nil {
			t.Errorf("code should panic")
		}
		world.Destroy()
	}(w)
	f := ecs.GetFilterWithExc[ecs.Inc1[C1], ecs.Exc1[C2]](w)
	p1 := ecs.GetPool[C1](w)
	// p2 := ecs.GetPool[C2](w)
	p3 := ecs.GetPool[C3](w)
	e := w.NewEntity()
	p1.Add(e)
	p3.Add(e)
	for it := f.Iter(); it.Next(); {
		p1.Del(it.GetEntity())
		p1.Add(it.GetEntity())
	}
	t.Errorf("code should panic")
}

func TestFilterInvalidAddDelRelatedComponent(t *testing.T) {
	w := ecs.NewWorld()
	defer func(world *ecs.World) {
		if r := recover(); r == nil {
			t.Errorf("code should panic")
		}
		world.Destroy()
	}(w)
	f := ecs.GetFilterWithExc[ecs.Inc1[C1], ecs.Exc1[C2]](w)
	p1 := ecs.GetPool[C1](w)
	p2 := ecs.GetPool[C2](w)
	p3 := ecs.GetPool[C3](w)
	e := w.NewEntity()
	p1.Add(e)
	p3.Add(e)
	for it := f.Iter(); it.Next(); {
		p2.Add(it.GetEntity())
		p2.Del(it.GetEntity())
	}
	t.Errorf("code should panic")
}

func TestFilterInvalidAddExcludedComponent(t *testing.T) {
	w := ecs.NewWorld()
	defer func(world *ecs.World) {
		if r := recover(); r == nil {
			t.Errorf("code should panic")
		}
		world.Destroy()
	}(w)
	f := ecs.GetFilterWithExc[ecs.Inc1[C1], ecs.Exc1[C2]](w)
	p1 := ecs.GetPool[C1](w)
	p2 := ecs.GetPool[C2](w)
	e := w.NewEntity()
	p1.Add(e)
	for it := f.Iter(); it.Next(); {
		e1 := w.NewEntity()
		p1.Add(e1)
		p2.Add(e1)
	}
	t.Errorf("code should panic")
}

func TestFilterInvalidRemoveIncludedComponent(t *testing.T) {
	w := ecs.NewWorld()
	defer func(world *ecs.World) {
		if r := recover(); r == nil {
			t.Errorf("code should panic")
		}
		// cant cleanup broken world.
		// world.Destroy()
	}(w)
	f := ecs.GetFilter[ecs.Inc1[C1]](w)
	p1 := ecs.GetPool[C1](w)
	e := w.NewEntity()
	p1.Add(e)
	for it := f.Iter(); it.Next(); {
		e1 := w.NewEntity()
		p1.Add(e1)
		p1.Del(e1)
	}
	t.Errorf("code should panic")
}

func BenchmarkFilterWithOneEmptyInc(b *testing.B) {
	w := ecs.NewWorld()
	p := ecs.GetPool[C1](w)
	for i := 0; i < 100000; i++ {
		p.Add(w.NewEntity())
	}
	f := ecs.GetFilter[ecs.Inc1[C1]](w)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for it := f.Iter(); it.Next(); {
			_ = p.Get(it.GetEntity())
		}
	}
	b.StopTimer()
	w.Destroy()
}

func BenchmarkFilterWithOneNonEmptyInc(b *testing.B) {
	w := ecs.NewWorld()
	p := ecs.GetPool[C2](w)
	for i := 0; i < 100000; i++ {
		p.Add(w.NewEntity())
	}
	f := ecs.GetFilter[ecs.Inc1[C2]](w)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for it := f.Iter(); it.Next(); {
			_ = p.Get(it.GetEntity())
		}
	}
	b.StopTimer()
	w.Destroy()
}

func BenchmarkFilterWithTwoInc(b *testing.B) {
	w := ecs.NewWorld()
	p1 := ecs.GetPool[C1](w)
	p2 := ecs.GetPool[C2](w)
	for i := 0; i < 100000; i++ {
		e := w.NewEntity()
		p1.Add(e)
		p2.Add(e)
	}
	f := ecs.GetFilter[ecs.Inc2[C1, C2]](w)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for it := f.Iter(); it.Next(); {
			_ = p1.Get(it.GetEntity())
		}
	}
	b.StopTimer()
	w.Destroy()
}

func BenchmarkFilterWithTreeInc(b *testing.B) {
	w := ecs.NewWorld()
	p1 := ecs.GetPool[C1](w)
	p2 := ecs.GetPool[C2](w)
	p3 := ecs.GetPool[C3](w)
	for i := 0; i < 100000; i++ {
		e := w.NewEntity()
		p1.Add(e)
		p2.Add(e)
		p3.Add(e)
	}
	f := ecs.GetFilter[ecs.Inc3[C1, C2, C3]](w)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for it := f.Iter(); it.Next(); {
			_ = p1.Get(it.GetEntity())
		}
	}
	b.StopTimer()
	w.Destroy()
}

func BenchmarkFilterWithFourInc(b *testing.B) {
	w := ecs.NewWorld()
	p1 := ecs.GetPool[C1](w)
	p2 := ecs.GetPool[C2](w)
	p3 := ecs.GetPool[C3](w)
	p4 := ecs.GetPool[C4](w)
	for i := 0; i < 100000; i++ {
		e := w.NewEntity()
		p1.Add(e)
		p2.Add(e)
		p3.Add(e)
		p4.Add(e)
	}
	f := ecs.GetFilter[ecs.Inc4[C1, C2, C3, C4]](w)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for it := f.Iter(); it.Next(); {
			_ = p1.Get(it.GetEntity())
		}
	}
	b.StopTimer()
	w.Destroy()
}

func BenchmarkFilterWithTwoIncAndOneExc(b *testing.B) {
	w := ecs.NewWorld()
	p1 := ecs.GetPool[C1](w)
	p2 := ecs.GetPool[C2](w)
	for i := 0; i < 100000; i++ {
		e := w.NewEntity()
		p1.Add(e)
		p2.Add(e)
	}
	q := ecs.GetFilterWithExc[ecs.Inc2[C1, C2], ecs.Exc1[C3]](w)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for it := q.Iter(); it.Next(); {
			_ = p1.Get(it.GetEntity())
		}
	}
	b.StopTimer()
	w.Destroy()
}
