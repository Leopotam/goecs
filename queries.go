// ----------------------------------------------------------------------------
// The Proprietary or MIT-Red License
// Copyright (c) 2012-2022 Leopotam <leopotam@yandex.ru>
// ----------------------------------------------------------------------------

package ecs // import "leopotam.com/go/ecs"

import (
	"math"
)

type Query struct {
	world *World
	inc   []IPool
	locks int
	iter  Iter
}

type QueryWithExc struct {
	world *World
	inc   []IPool
	exc   []IPool
	locks int
	iter  IterWithExc
}

type IInc interface {
	CopyWithPools(w *World) IInc
	GetPools(w *World) []IPool
}

type IExc interface {
	Fill(w *World) []IPool
}

func NewQuery[I IInc](w *World) *Query {
	q := &Query{world: w}
	var i I
	q.inc = i.GetPools(w)
	incsLen := len(q.inc)
	q.iter = Iter{
		q:            q,
		sparseIncs:   make([][]int, incsLen),
		otherIncsLen: incsLen - 1,
		entity:       -1,
		idx:          -1,
	}
	return q
}

func NewQueryWithExc[I IInc, E IExc](w *World) *QueryWithExc {
	q := &QueryWithExc{world: w}
	var i I
	var e E
	q.inc = i.GetPools(w)
	q.exc = e.Fill(w)
	incsLen := len(q.inc)
	q.iter = IterWithExc{
		q:            q,
		sparseIncs:   make([][]int, incsLen),
		sparseExcs:   make([][]int, len(q.exc)),
		otherIncsLen: incsLen - 1,
		entity:       -1,
		idx:          -1,
	}
	return q
}

type Iter struct {
	q            *Query
	indices      []int
	sparseIncs   [][]int
	length       int
	otherIncsLen int
	idx          int
	entity       int
}

type IterWithExc struct {
	q            *QueryWithExc
	indices      []int
	sparseIncs   [][]int
	sparseExcs   [][]int
	length       int
	otherIncsLen int
	idx          int
	entity       int
}

func (q *Query) GetWorld() *World {
	return q.world
}

func (q *Query) Iter() Iter {
	if q.locks == 0 {
		itemsCount := math.MaxInt
		var poolIdx int
		for idx, pool := range q.inc {
			newMin := pool.GetItemsCount()
			if newMin < itemsCount {
				itemsCount = newMin
				poolIdx = idx
			}
			q.iter.sparseIncs[idx] = pool.GetSparseIndices()
		}
		q.iter.indices = q.inc[poolIdx].GetIndices()
		q.iter.length = len(q.iter.indices)
		if q.iter.otherIncsLen > 0 {
			q.iter.sparseIncs[poolIdx] = q.iter.sparseIncs[q.iter.otherIncsLen]
		}
	}
	q.locks++
	return q.iter
}

func (i *Iter) Next() bool {
	for {
		i.idx++
		if i.idx >= i.length {
			i.q.locks--
			if DEBUG {
				if i.q.locks < 0 {
					panic("Query lock/unlock invalid balance.")
				}
			}
			return false
		}
		i.entity = i.indices[i.idx]
		if i.entity >= 0 {
			for incIdx := 0; incIdx < i.otherIncsLen; incIdx++ {
				if i.sparseIncs[incIdx][i.entity] <= 0 {
					i.entity = -1
					break
				}
			}
			if i.entity >= 0 {
				return true
			}
		}
	}
}

func (i *Iter) GetEntity() int {
	return i.entity
}

func (q *QueryWithExc) GetWorld() *World {
	return q.world
}

func (q *QueryWithExc) Iter() IterWithExc {
	if q.locks == 0 {
		itemsCount := math.MaxInt
		var poolIdx int
		for idx, pool := range q.inc {
			newMin := pool.GetItemsCount()
			if newMin < itemsCount {
				itemsCount = newMin
				poolIdx = idx
			}
			q.iter.sparseIncs[idx] = pool.GetSparseIndices()
		}
		for idx, pool := range q.exc {
			q.iter.sparseExcs[idx] = pool.GetSparseIndices()
		}
		q.iter.indices = q.inc[poolIdx].GetIndices()
		q.iter.length = len(q.iter.indices)
		if q.iter.otherIncsLen > 0 {
			q.iter.sparseIncs[poolIdx] = q.iter.sparseIncs[q.iter.otherIncsLen]
		}
	}
	q.locks++
	return q.iter
}

func (i *IterWithExc) Next() bool {
	for {
		i.idx++
		if i.idx >= i.length {
			i.q.locks--
			if DEBUG {
				if i.q.locks < 0 {
					panic("Query lock/unlock invalid balance.")
				}
			}
			return false
		}
		i.entity = i.indices[i.idx]
		if i.entity >= 0 {
			for incIdx := 0; incIdx < i.otherIncsLen; incIdx++ {
				if i.sparseIncs[incIdx][i.entity] <= 0 {
					i.entity = -1
					break
				}
			}
			if i.entity >= 0 {
				for _, sparse := range i.sparseExcs {
					if sparse[i.entity] > 0 {
						i.entity = -1
						break
					}
				}
			}
			if i.entity >= 0 {
				return true
			}
		}
	}
}

func (i *IterWithExc) GetEntity() int {
	return i.entity
}

type Inc1[I1 any] struct {
	Inc1 *Pool[I1]
}

func (i Inc1[I1]) CopyWithPools(w *World) IInc {
	return &Inc1[I1]{
		Inc1: GetPool[I1](w),
	}
}

func (i Inc1[I1]) GetPools(w *World) []IPool {
	return []IPool{GetPool[I1](w)}
}

type Inc2[I1 any, I2 any] struct {
	Inc1 *Pool[I1]
	Inc2 *Pool[I2]
}

func (i Inc2[I1, I2]) CopyWithPools(w *World) IInc {
	return &Inc2[I1, I2]{
		Inc1: GetPool[I1](w),
		Inc2: GetPool[I2](w),
	}
}

func (i Inc2[I1, I2]) GetPools(w *World) []IPool {
	return []IPool{GetPool[I1](w), GetPool[I2](w)}
}

type Inc3[I1 any, I2 any, I3 any] struct {
	Inc1 *Pool[I1]
	Inc2 *Pool[I2]
	Inc3 *Pool[I3]
}

func (i Inc3[I1, I2, I3]) CopyWithPools(w *World) IInc {
	return &Inc3[I1, I2, I3]{
		Inc1: GetPool[I1](w),
		Inc2: GetPool[I2](w),
		Inc3: GetPool[I3](w),
	}
}

func (i Inc3[I1, I2, I3]) GetPools(w *World) []IPool {
	return []IPool{GetPool[I1](w), GetPool[I2](w), GetPool[I3](w)}
}

type Inc4[I1 any, I2 any, I3 any, I4 any] struct {
	Inc1 *Pool[I1]
	Inc2 *Pool[I2]
	Inc3 *Pool[I3]
	Inc4 *Pool[I4]
}

func (i Inc4[I1, I2, I3, I4]) CopyWithPools(w *World) IInc {
	return &Inc4[I1, I2, I3, I4]{
		Inc1: GetPool[I1](w),
		Inc2: GetPool[I2](w),
		Inc3: GetPool[I3](w),
		Inc4: GetPool[I4](w),
	}
}

func (i Inc4[I1, I2, I3, I4]) GetPools(w *World) []IPool {
	return []IPool{GetPool[I1](w), GetPool[I2](w), GetPool[I3](w), GetPool[I4](w)}
}

type Exc1[E1 any] struct{}
type Exc2[E1 any, E2 any] struct{}

func (i Exc1[E1]) Fill(w *World) []IPool {
	return []IPool{GetPool[E1](w)}
}

func (i Exc2[E1, E2]) Fill(w *World) []IPool {
	return []IPool{GetPool[E1](w), GetPool[E2](w)}
}
