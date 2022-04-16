// ----------------------------------------------------------------------------
// The Proprietary or MIT-Red License
// Copyright (c) 2012-2022 Leopotam <leopotam@yandex.ru>
// ----------------------------------------------------------------------------

package ecs

import (
	"math"
	"reflect"
)

type WorldConfig struct {
	WorldEntitiesSize         int
	WorldEntitiesRecycledSize int
	WorldPoolsSize            int
	PoolDenseSize             int
	PoolRecycledSize          int
}

const defaultWorldEntitiesSize int = 512
const defaultWorldEntitiesRecycledSize int = 512
const defaultWorldPoolsSize int = 128
const defaultPoolDenseSize int = 512
const defaultPoolRecycledSize int = 512

type entityData struct {
	ComponentsCount int
	gen             int16
}

type World struct {
	config              WorldConfig
	entities            []entityData
	pools               map[reflect.Type]IEcsPool
	entitiesRecycled    []int
	debugLeakedEntities []int
}

func NewWorld() *World {
	return NewWorldWithConfig(WorldConfig{})
}

func NewWorldWithConfig(config WorldConfig) *World {
	if config.WorldEntitiesSize <= 0 {
		config.WorldEntitiesSize = defaultWorldEntitiesSize
	}
	if config.WorldEntitiesRecycledSize <= 0 {
		config.WorldEntitiesRecycledSize = defaultWorldEntitiesRecycledSize
	}
	if config.WorldPoolsSize <= 0 {
		config.WorldPoolsSize = defaultWorldPoolsSize
	}
	if config.PoolDenseSize <= 0 {
		config.PoolDenseSize = defaultPoolDenseSize
	}
	if config.PoolRecycledSize <= 0 {
		config.PoolRecycledSize = defaultPoolRecycledSize
	}
	w := &World{}
	w.config = config
	w.entities = make([]entityData, 0, config.WorldEntitiesSize)
	w.pools = make(map[reflect.Type]IEcsPool, config.WorldPoolsSize)
	w.entitiesRecycled = make([]int, 0, config.WorldEntitiesRecycledSize)
	if DEBUG {
		w.debugLeakedEntities = make([]int, 0, 512)
	}
	return w
}

func (w *World) Destroy() {
	if DEBUG {
		if debugCheckWorldForLeakedEntities(w) {
			panic("Empty entity detected before EcsWorld.Destroy().")
		}
	}
	for i := 0; i < len(w.entities); i++ {
		if w.entities[i].ComponentsCount > 0 {
			w.DelEntity(i)
		}
	}
	w.entities = w.entities[:0]
	for k := range w.pools {
		delete(w.pools, k)
	}
	w.entitiesRecycled = w.entitiesRecycled[:0]
}

func (w *World) NewEntity() int {
	var entity int
	l := len(w.entitiesRecycled)
	if l > 0 {
		entity = w.entitiesRecycled[l-1]
		w.entitiesRecycled = w.entitiesRecycled[:l-1]
		entityData := &w.entities[entity]
		entityData.gen = -entityData.gen
	} else {
		// new entity.
		entity = len(w.entities)
		oldCap := cap(w.entities)
		w.entities = append(w.entities, entityData{gen: 1})
		newCap := cap(w.entities)
		if oldCap != newCap {
			// resize entities and component pools.
			for _, p := range w.pools {
				p.Resize(newCap)
			}
		}
	}
	if DEBUG {
		w.debugLeakedEntities = append(w.debugLeakedEntities, entity)
	}
	return entity
}

func (w *World) DelEntity(entity int) {
	if DEBUG {
		if entity < 0 || entity >= len(w.entities) {
			panic("Cant touch invalid entity.")
		}
	}
	entityData := &w.entities[entity]
	if entityData.gen < 0 {
		return
	}
	// kill components.
	if entityData.ComponentsCount > 0 {
		for _, pool := range w.pools {
			if pool.Has(entity) {
				pool.Del(entity)
				if entityData.ComponentsCount == 0 {
					return
				}
			}
		}
	}
	if entityData.gen == math.MaxInt16 {
		entityData.gen = -1
	} else {
		entityData.gen = -(entityData.gen + 1)
	}
	w.entitiesRecycled = append(w.entitiesRecycled, entity)
}

func (w *World) GetEntityGen(entity int) int16 {
	return w.entities[entity].gen
}

func GetPool[T any](w *World) *Pool[T] {
	itemType := reflect.TypeOf((*T)(nil))
	if pool, ok := w.pools[itemType]; ok {
		return pool.(*Pool[T])
	}
	pool := newPool[T](w, w.config.PoolDenseSize, cap(w.entities), w.config.PoolRecycledSize)
	w.pools[itemType] = pool
	return pool
}

func debugCheckWorldForLeakedEntities(w *World) bool {
	if len(w.debugLeakedEntities) > 0 {
		for _, leakedEntity := range w.debugLeakedEntities {
			entityData := w.entities[leakedEntity]
			if entityData.gen > 0 && entityData.ComponentsCount == 0 {
				w.debugLeakedEntities = w.debugLeakedEntities[:0]
				return true
			}
		}
		w.debugLeakedEntities = w.debugLeakedEntities[:0]
	}
	return false
}

func debugCheckEntityAlive(w *World, entity int) bool {
	return entity >= 0 && entity < len(w.entities) && w.entities[entity].gen > 0
}
