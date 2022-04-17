// ----------------------------------------------------------------------------
// The Proprietary or MIT-Red License
// Copyright (c) 2012-2022 Leopotam <leopotam@yandex.ru>
// ----------------------------------------------------------------------------

package goecs

import (
	"math"
)

type Query struct {
	inc   []IEcsPool
	locks int
	iter  Iter
}

type QueryWithExc struct {
	inc   []IEcsPool
	exc   []IEcsPool
	locks int
	iter  IterWithExc
}

type IInc interface {
	Fill(w *World) []IEcsPool
}

type IExc interface {
	Fill(w *World) []IEcsPool
}

func NewQuery[I IInc](w *World) *Query {
	q := &Query{}
	var i I
	q.inc = i.Fill(w)
	q.iter.q = q
	q.iter.sparseIncs = make([][]int, len(q.inc))
	return q
}

func NewQueryWithExc[I IInc, E IExc](w *World) *QueryWithExc {
	q := &QueryWithExc{}
	var i I
	var e E
	q.inc = i.Fill(w)
	q.exc = e.Fill(w)
	q.iter.q = q
	q.iter.sparseIncs = make([][]int, len(q.inc))
	q.iter.sparseExcs = make([][]int, len(q.exc))
	return q
}

type Iter struct {
	q          *Query
	poolIdx    int
	indices    []int
	length     int
	sparseIncs [][]int
	idx        int
	entity     int
}

type IterWithExc struct {
	q          *QueryWithExc
	poolIdx    int
	indices    []int
	length     int
	sparseIncs [][]int
	sparseExcs [][]int
	idx        int
	entity     int
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
		q.iter.poolIdx = poolIdx
		q.iter.indices = q.inc[poolIdx].GetIndices()
		q.iter.length = len(q.iter.indices)
	}
	q.locks++
	iter := q.iter
	iter.idx = -1
	iter.entity = -1
	return iter
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
			for _, sparse := range i.sparseIncs {
				// if idx != i.poolIdx && sparse[i.entity] <= 0 {
				if sparse[i.entity] <= 0 {
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
		q.iter.poolIdx = poolIdx
		q.iter.indices = q.inc[poolIdx].GetIndices()
		q.iter.length = len(q.iter.indices)
	}
	q.locks++
	iter := q.iter
	iter.idx = -1
	iter.entity = -1
	return iter
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
			for _, sparse := range i.sparseIncs {
				// if idx != i.poolIdx && sparse[i.entity] <= 0 {
				if sparse[i.entity] <= 0 {
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

type Inc1[I1 any] struct{}

func (i Inc1[I1]) Fill(w *World) []IEcsPool {
	return []IEcsPool{GetPool[I1](w)}
}

type Inc2[I1 any, I2 any] struct{}

func (i Inc2[I1, I2]) Fill(w *World) []IEcsPool {
	return []IEcsPool{GetPool[I1](w), GetPool[I2](w)}
}

type Inc3[I1 any, I2 any, I3 any] struct{}

func (i Inc3[I1, I2, I3]) Fill(w *World) []IEcsPool {
	return []IEcsPool{GetPool[I1](w), GetPool[I2](w), GetPool[I3](w)}
}

type Inc4[I1 any, I2 any, I3 any, I4 any] struct{}

func (i Inc4[I1, I2, I3, I4]) Fill(w *World) []IEcsPool {
	return []IEcsPool{GetPool[I1](w), GetPool[I2](w), GetPool[I3](w), GetPool[I4](w)}
}

type Exc1[E1 any] struct{}
type Exc2[E1 any, E2 any] struct{}

func (i Exc1[E1]) Fill(w *World) []IEcsPool {
	return []IEcsPool{GetPool[E1](w)}
}

func (i Exc2[E1, E2]) Fill(w *World) []IEcsPool {
	return []IEcsPool{GetPool[E1](w), GetPool[E2](w)}
}
