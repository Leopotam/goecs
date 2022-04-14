// ----------------------------------------------------------------------------
// The Proprietary or MIT-Red License
// Copyright (c) 2012-2022 Leopotam <leopotam@yandex.ru>
// ----------------------------------------------------------------------------

package ecs

import (
	"fmt"
	"reflect"
)

type IEcsReset interface {
	Reset()
}

type IEcsPool interface {
	GetWorld() *World
	Resize(capacity int)
	Has(entity int) bool
	Del(entity int)
	GetIndices() []int
	GetItemsCount() int
}

type Pool[T any] struct {
	world      *World
	items      []T
	sparseSet  *sparseSet
	itemsCount int
}

func newPool[T any](world *World, denseCapacity int, sparseCapacity int, recycledCapacity int) *Pool[T] {
	p := &Pool[T]{}
	p.world = world
	p.items = make([]T, 1, denseCapacity+1)
	p.sparseSet = newSparseSet(denseCapacity, sparseCapacity)
	p.itemsCount = 0
	return p
}

func (p *Pool[T]) Add(entity int) *T {
	if DEBUG {
		if p.Has(entity) {
			panic(fmt.Sprintf("Component \"%s\" already attached to entity.", reflect.TypeOf((*T)(nil)).Name()))
		}
	}
	idx, isNew := p.sparseSet.Set(entity)
	if isNew {
		if DEBUG {
			if len(p.items) != idx {
				panic("something wrong here.")
			}
		}
		var defaultT T
		if r, ok := any(&defaultT).(IEcsReset); ok {
			r.Reset()
		}
		p.items = append(p.items, defaultT)
	}
	p.itemsCount++
	p.world.entities[entity].ComponentsCount++
	return &p.items[idx]
}

func (p *Pool[T]) Get(entity int) *T {
	if DEBUG {
		if !p.Has(entity) {
			panic(fmt.Sprintf("Component \"%s\" not attached to entity.", reflect.TypeOf((*T)(nil)).Name()))
		}
	}
	return &p.items[p.sparseSet.Get(entity)]
}

func (p *Pool[T]) GetWorld() *World {
	return p.world
}

func (p *Pool[T]) Resize(capacity int) {
	p.sparseSet.Resize(capacity)
}

func (p *Pool[T]) Has(entity int) bool {
	if DEBUG {
		if !debugCheckEntityAlive(p.world, entity) {
			panic("Cant touch destroyed entity.")
		}
	}
	return p.sparseSet.Has(entity)
}

func (p *Pool[T]) Del(entity int) {
	idx := p.sparseSet.Del(entity)
	if r, ok := any(&p.items[idx]).(IEcsReset); ok {
		r.Reset()
	} else {
		var defaultT T
		p.items[idx] = defaultT
	}
	p.itemsCount--
	entityData := &p.world.entities[entity]
	entityData.ComponentsCount--
	if entityData.ComponentsCount == 0 {
		p.world.DelEntity(entity)
	}
}

func (p *Pool[T]) GetIndices() []int {
	return p.sparseSet.GetIndices()
}

func (p *Pool[T]) GetItemsCount() int {
	return p.itemsCount
}
