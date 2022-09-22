// ----------------------------------------------------------------------------
// The Proprietary or MIT-Red License
// Copyright (c) 2012-2022 Leopotam <leopotam@yandex.ru>
// ----------------------------------------------------------------------------

package ecs // import "leopotam.com/go/ecs"

import (
	"fmt"
	"reflect"
)

type IComponentReset interface {
	Reset()
}

type IComponentCopy[T any] interface {
	Copy(src *T)
}

type IPool interface {
	GetID() int16
	GetWorld() *World
	Resize(capacity int)
	Has(entity int) bool
	Del(entity int)
	GetSparseIndices() []int
	GetRaw(entity int) any
	GetItemType() reflect.Type
	Copy(srcEntity, dstEntity int)
}

type Pool[T any] struct {
	id              int16
	world           *World
	itemType        reflect.Type
	items           []T
	sparseIndices   []int
	recycledIndices []int
}

func newPool[T any](world *World, id int16, denseCapacity int, sparseCapacity int, recycledCapacity int) *Pool[T] {
	p := &Pool[T]{}
	p.id = id
	p.world = world
	p.itemType = reflect.TypeOf(p.items).Elem()
	p.items = make([]T, 1, denseCapacity+1)
	p.sparseIndices = make([]int, sparseCapacity)
	p.recycledIndices = make([]int, 0, denseCapacity+1)
	return p
}

func (p *Pool[T]) GetID() int16 {
	return p.id
}

func (p *Pool[T]) Add(entity int) *T {
	if DEBUG {
		if p.Has(entity) {
			panic(fmt.Sprintf("component \"%s\" already attached to entity", reflect.TypeOf(p.items).Elem().String()))
		}
	}
	l := len(p.recycledIndices)
	isNew := l == 0
	var denseIdx int
	if isNew {
		denseIdx = len(p.items)
		var defaultT T
		if r, ok := any(&defaultT).(IComponentReset); ok {
			r.Reset()
		}
		p.items = append(p.items, defaultT)
	} else {
		denseIdx = p.recycledIndices[l-1]
		p.recycledIndices = p.recycledIndices[:l-1]
	}
	p.sparseIndices[entity] = denseIdx
	p.world.onEntityChange(entity, p.id, true)
	p.world.addComponentToRawEntity(entity, p.id)
	if DEBUG {
		for _, l := range p.world.debugEventListeners {
			l.OnEntityChanged(entity)
		}
	}
	return &p.items[denseIdx]
}

func (p *Pool[T]) Get(entity int) *T {
	if DEBUG {
		if !p.Has(entity) {
			panic(fmt.Sprintf("component \"%s\" not attached to entity", reflect.TypeOf(p.items).Elem().String()))
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
			panic("cant touch destroyed entity")
		}
	}
	return p.sparseIndices[entity] > 0
}

func (p *Pool[T]) Del(entity int) {
	if DEBUG {
		if !p.world.checkEntityAlive(entity) {
			panic("cant touch destroyed entity")
		}
	}
	if p.sparseIndices[entity] <= 0 {
		return
	}
	p.world.onEntityChange(entity, p.id, false)
	denseIdx := p.sparseIndices[entity]
	p.sparseIndices[entity] = 0
	p.recycledIndices = append(p.recycledIndices, denseIdx)

	if r, ok := any(&p.items[denseIdx]).(IComponentReset); ok {
		r.Reset()
	} else {
		var defaultT T
		p.items[denseIdx] = defaultT
	}
	p.world.removeComponentFromRawEntity(entity, p.id)
	componentsCount := p.world.GetEntityComponentsCount(entity)
	if DEBUG {
		for _, l := range p.world.debugEventListeners {
			l.OnEntityChanged(entity)
		}
	}
	if componentsCount == 0 {
		p.world.DelEntity(entity)
	}
}

func (p *Pool[T]) GetSparseIndices() []int {
	return p.sparseIndices
}

func (p *Pool[T]) GetRaw(entity int) any {
	return p.Get(entity)
}

func (p *Pool[T]) GetItemType() reflect.Type {
	return p.itemType
}

func (p *Pool[T]) Copy(srcEntity, dstEntity int) {
	if DEBUG {
		if !p.world.checkEntityAlive(srcEntity) {
			panic("cant touch destroyed src-entity")
		}
		if !p.world.checkEntityAlive(dstEntity) {
			panic("cant touch destroyed dst-entity")
		}
	}
	if p.Has(srcEntity) {
		srcData := p.Get(srcEntity)
		if !p.Has(dstEntity) {
			p.Add(dstEntity)
		}
		dstData := p.Get(dstEntity)
		if c, ok := any(dstData).(IComponentCopy[T]); ok {
			c.Copy(srcData)
		} else {
			*dstData = *srcData
		}
	}
}
