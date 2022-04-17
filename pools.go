// ----------------------------------------------------------------------------
// The Proprietary or MIT-Red License
// Copyright (c) 2012-2022 Leopotam <leopotam@yandex.ru>
// ----------------------------------------------------------------------------

package goecs

import (
	"fmt"
	"reflect"
)

type IComponentReset interface {
	Reset()
}

type IPool interface {
	GetWorld() *World
	Resize(capacity int)
	Has(entity int) bool
	Del(entity int)
	GetIndices() []int
	GetSparseIndices() []int
	GetItemsCount() int
}

type Pool[T any] struct {
	world           *World
	items           []T
	itemsCount      int
	denseIndices    []int
	sparseIndices   []int
	recycledIndices []int
}

func newPool[T any](world *World, denseCapacity int, sparseCapacity int, recycledCapacity int) *Pool[T] {
	p := &Pool[T]{}
	p.world = world
	p.items = make([]T, 1, denseCapacity+1)
	p.itemsCount = 0
	p.denseIndices = make([]int, 1, denseCapacity+1)
	p.sparseIndices = make([]int, sparseCapacity)
	p.recycledIndices = make([]int, 0, denseCapacity+1)
	return p
}

func (p *Pool[T]) Add(entity int) *T {
	if DEBUG {
		if p.Has(entity) {
			panic(fmt.Sprintf("Component \"%s\" already attached to entity.", reflect.TypeOf(p.items).Elem().String()))
		}
	}
	l := len(p.recycledIndices)
	isNew := l == 0
	var denseIdx int
	if isNew {
		denseIdx = len(p.denseIndices)
		p.denseIndices = append(p.denseIndices, entity)
		var defaultT T
		if r, ok := any(&defaultT).(IComponentReset); ok {
			r.Reset()
		}
		p.items = append(p.items, defaultT)
	} else {
		denseIdx = p.recycledIndices[l-1]
		p.recycledIndices = p.recycledIndices[:l-1]
		p.denseIndices[denseIdx] = entity
	}
	p.sparseIndices[entity] = denseIdx
	p.itemsCount++
	p.world.entities[entity].ComponentsCount++
	return &p.items[denseIdx]
}

func (p *Pool[T]) Get(entity int) *T {
	if DEBUG {
		if !p.Has(entity) {
			panic(fmt.Sprintf("Component \"%s\" not attached to entity.", reflect.TypeOf(p.items).Elem().String()))
		}
	}
	return &p.items[p.sparseIndices[entity]]
}

func (p *Pool[T]) GetWorld() *World {
	return p.world
}

func (p *Pool[T]) Resize(capacity int) {
	ss := make([]int, capacity)
	copy(ss, p.sparseIndices)
	p.sparseIndices = ss
}

func (p *Pool[T]) Has(entity int) bool {
	if DEBUG {
		if !p.world.checkEntityAlive(entity) {
			panic("Cant touch destroyed entity.")
		}
	}
	return p.sparseIndices[entity] > 0
}

func (p *Pool[T]) Del(entity int) {
	if DEBUG {
		if !p.world.checkEntityAlive(entity) {
			panic("Cant touch destroyed entity.")
		}
	}
	if p.sparseIndices[entity] <= 0 {
		return
	}
	denseIdx := p.sparseIndices[entity]
	p.denseIndices[denseIdx] = -1
	p.sparseIndices[entity] = 0
	p.recycledIndices = append(p.recycledIndices, denseIdx)

	if r, ok := any(&p.items[denseIdx]).(IComponentReset); ok {
		r.Reset()
	} else {
		var defaultT T
		p.items[denseIdx] = defaultT
	}
	p.itemsCount--
	entityData := &p.world.entities[entity]
	entityData.ComponentsCount--
	if entityData.ComponentsCount == 0 {
		p.world.DelEntity(entity)
	}
}

func (p *Pool[T]) GetIndices() []int {
	return p.denseIndices[1:]
}

func (p *Pool[T]) GetSparseIndices() []int {
	return p.sparseIndices
}

func (p *Pool[T]) GetItemsCount() int {
	return p.itemsCount
}
