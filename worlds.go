// ----------------------------------------------------------------------------
// The Proprietary or MIT-Red License
// Copyright (c) 2012-2022 Leopotam <leopotam@yandex.ru>
// ----------------------------------------------------------------------------

package ecs // import "leopotam.com/go/ecs"

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

type IWorldEventListener interface {
	OnEntityCreated(entity int)
	OnEntityChanged(entity int)
	OnEntityDestroyed(entity int)
	OnWorldResized(newSize int)
	OnWorldDestroyed(world *World)
}

type World struct {
	config              WorldConfig
	entities            []entityData
	entitiesRecycled    []int
	pools               []IPool
	poolsHashes         map[reflect.Type]IPool
	filterMaskCache     [][]int
	filtersHashes       map[int]*Filter
	filtersByIncludes   [][]*Filter
	filtersByExcludes   [][]*Filter
	debugLeakedEntities []int
	debugEventListeners []IWorldEventListener
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
	w.entitiesRecycled = make([]int, 0, config.WorldEntitiesRecycledSize)
	w.pools = make([]IPool, 0, config.WorldPoolsSize)
	w.poolsHashes = make(map[reflect.Type]IPool, config.WorldPoolsSize)
	w.filtersHashes = make(map[int]*Filter, config.WorldPoolsSize)
	w.filtersByIncludes = make([][]*Filter, config.WorldPoolsSize)
	w.filtersByExcludes = make([][]*Filter, config.WorldPoolsSize)
	if DEBUG {
		w.debugLeakedEntities = make([]int, 0, 512)
	}
	return w
}

func (w *World) Destroy() {
	if DEBUG {
		if debugCheckWorldForLeakedEntities(w) {
			panic("empty entity detected before EcsWorld.Destroy()")
		}
	}
	for i := 0; i < len(w.entities); i++ {
		if w.entities[i].ComponentsCount > 0 {
			w.DelEntity(i)
		}
	}
	w.entities = w.entities[:0]
	for k := range w.poolsHashes {
		delete(w.poolsHashes, k)
	}
	w.pools = w.pools[:0]
	w.entitiesRecycled = w.entitiesRecycled[:0]
	for k := range w.filtersHashes {
		delete(w.filtersHashes, k)
	}
	w.filtersByIncludes = w.filtersByIncludes[:0]
	w.filtersByExcludes = w.filtersByExcludes[:0]
	if DEBUG {
		for _, l := range w.debugEventListeners {
			l.OnWorldDestroyed(w)
		}
	}
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
			for _, f := range w.filtersHashes {
				f.resizeSparseIndex(newCap)
			}
			if DEBUG {
				for _, l := range w.debugEventListeners {
					l.OnWorldResized(entity)
				}
			}
		}
	}
	if DEBUG {
		w.debugLeakedEntities = append(w.debugLeakedEntities, entity)
		for _, l := range w.debugEventListeners {
			l.OnEntityCreated(entity)
		}
	}
	return entity
}

func (w *World) DelEntity(entity int) {
	if DEBUG {
		if entity < 0 || entity >= len(w.entities) {
			panic("cant touch invalid entity")
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
	if DEBUG {
		for _, l := range w.debugEventListeners {
			l.OnEntityDestroyed(entity)
		}
	}
}

func (w *World) GetEntityGen(entity int) int16 {
	return w.entities[entity].gen
}

func (w *World) GetComponentValues(entity int, list []any) []any {
	if w.entities[entity].ComponentsCount > 0 {
		for _, p := range w.pools {
			if p.Has(entity) {
				list = append(list, p.GetRaw(entity))
			}
		}
	}
	return list
}

func (w *World) GetComponentTypes(entity int, list []reflect.Type) []reflect.Type {
	if w.entities[entity].ComponentsCount > 0 {
		for _, p := range w.pools {
			if p.Has(entity) {
				list = append(list, p.GetItemType())
			}
		}
	}
	return list
}

func GetPool[T any](w *World) *Pool[T] {
	itemType := reflect.TypeOf((*T)(nil))
	if pool, ok := w.poolsHashes[itemType]; ok {
		return pool.(*Pool[T])
	}
	pool := newPool[T](w, len(w.pools), w.config.PoolDenseSize, cap(w.entities), w.config.PoolRecycledSize)
	w.poolsHashes[itemType] = pool
	w.pools = append(w.pools, pool)
	w.filtersByIncludes = append(w.filtersByIncludes, nil)
	w.filtersByExcludes = append(w.filtersByExcludes, nil)
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

func (w *World) checkEntityAlive(entity int) bool {
	return entity >= 0 && entity < len(w.entities) && w.entities[entity].gen > 0
}
