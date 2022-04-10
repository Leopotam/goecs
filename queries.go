// ----------------------------------------------------------------------------
// The Proprietary or MIT-Red License
// Copyright (c) 2012-2022 Leopotam <leopotam@yandex.ru>
// ----------------------------------------------------------------------------

package ecs

import (
	"math"
)

type Query struct {
	inc []IEcsPool
}

type QueryWithExc struct {
	inc []IEcsPool
	exc []IEcsPool
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
	return q
}

func NewQueryWithExc[I IInc, E IExc](w *World) *QueryWithExc {
	q := &QueryWithExc{}
	var i I
	var e E
	q.inc = i.Fill(w)
	q.exc = e.Fill(w)
	return q
}

type Iter struct {
	q       *Query
	poolIdx int
	indices []int
	idx     int
	entity  int
}

type IterWithExc struct {
	q       *QueryWithExc
	poolIdx int
	indices []int
	idx     int
	entity  int
}

func (q *Query) Iter() Iter {
	itemsCount := math.MaxInt
	var poolIdx int
	var indices []int
	for idx, pool := range q.inc {
		newMin := pool.GetItemsCount()
		if newMin < itemsCount {
			itemsCount = newMin
			indices = pool.GetIndices()
			poolIdx = idx
		}
	}
	return Iter{
		q:       q,
		poolIdx: poolIdx,
		indices: indices,
		idx:     -1,
		entity:  -1,
	}
}

func (i *Iter) Next() bool {
	for {
		i.idx++
		if i.idx >= len(i.indices) {
			i.unlock()
			return false
		}
		i.entity = i.indices[i.idx]
		if i.entity < 0 {
			continue
		}
		for idx, pool := range i.q.inc {
			if idx == i.poolIdx {
				continue
			}
			if !pool.Has(i.entity) {
				i.entity = -1
				break
			}
		}
		if i.entity >= 0 {
			return true
		}
	}
}

func (i *Iter) GetEntity() int {
	return i.entity
}

func (i *Iter) unlock() {

}

func (q *QueryWithExc) Iter() IterWithExc {
	itemsCount := math.MaxInt
	var poolIdx int
	var indices []int
	for idx, pool := range q.inc {
		newMin := pool.GetItemsCount()
		if newMin < itemsCount {
			itemsCount = newMin
			indices = pool.GetIndices()
			poolIdx = idx
		}
	}
	return IterWithExc{
		q:       q,
		poolIdx: poolIdx,
		indices: indices,
		idx:     -1,
		entity:  -1,
	}
}

func (i *IterWithExc) Next() bool {
	for {
		i.idx++
		if i.idx >= len(i.indices) {
			i.unlock()
			return false
		}
		i.entity = i.indices[i.idx]
		if i.entity < 0 {
			continue
		}
		for idx, pool := range i.q.inc {
			if idx == i.poolIdx {
				continue
			}
			if !pool.Has(i.entity) {
				i.entity = -1
				break
			}
		}
		if i.entity >= 0 {
			for _, pool := range i.q.exc {
				if pool.Has(i.entity) {
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

func (i *IterWithExc) GetEntity() int {
	return i.entity
}

func (i *IterWithExc) unlock() {

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
