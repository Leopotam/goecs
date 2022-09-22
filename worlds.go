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
	EntityComponentsSize      int
}

const (
	RawEntityOffsetComponentsCount int = 0
	RawEntityOffsetGen             int = 1
	RawEntityOffsetComponents      int = 2
)

const defaultWorldEntitiesSize int = 512
const defaultWorldEntitiesRecycledSize int = 512
const defaultWorldPoolsSize int = 128
const defaultPoolDenseSize int = 512
const defaultPoolRecycledSize int = 512
const defaultEntityComponentsSize int = 8

type IWorldEventListener interface {
	OnEntityCreated(entity int)
	OnEntityChanged(entity int)
	OnEntityDestroyed(entity int)
	OnWorldResized(newSize int)
	OnWorldDestroyed(world *World)
}

type World struct {
	config WorldConfig
	// componentsCount, gen, c1, c2, ..., [next]
	entities            []int16
	entitiesItemSize    int
	entitiesRecycled    []int
	pools               []IPool
	poolsHashes         map[reflect.Type]IPool
	filterMaskCache     [][]int16
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
	if config.EntityComponentsSize <= 0 {
		config.EntityComponentsSize = defaultEntityComponentsSize
	}
	w := &World{}
	w.config = config
	w.entitiesItemSize = RawEntityOffsetComponents + config.EntityComponentsSize
	w.entities = make([]int16, 0, config.WorldEntitiesSize*w.entitiesItemSize)
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
	for i, iMax := 0, len(w.entities)/w.entitiesItemSize; i < iMax; i++ {
		if w.entities[w.GetRawEntityOffset(i)+RawEntityOffsetComponentsCount] > 0 {
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

func (w *World) GetRawEntityOffset(entity int) int {
	return entity * w.entitiesItemSize
}

func (w *World) NewEntity() int {
	var entity int
	l := len(w.entitiesRecycled)
	if l > 0 {
		entity = w.entitiesRecycled[l-1]
		w.entitiesRecycled = w.entitiesRecycled[:l-1]
		w.entities[w.GetRawEntityOffset(entity)+RawEntityOffsetGen] *= -1
	} else {
		// new entity.
		entity = len(w.entities) / w.entitiesItemSize
		oldCap := cap(w.entities)
		// add new entity entities.
		for i := 0; i < w.entitiesItemSize; i++ {
			w.entities = append(w.entities, 0)
		}
		w.entities[w.GetRawEntityOffset(entity)+RawEntityOffsetGen] = 1
		newCap := cap(w.entities)
		if oldCap != newCap {
			newCap /= w.entitiesItemSize
			// resize filters and component pools.
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
		if entity < 0 || (entity*w.entitiesItemSize) >= len(w.entities) {
			panic("cant touch invalid entity")
		}
	}
	entityOffset := w.GetRawEntityOffset(entity)
	componentsCount := int(w.entities[entityOffset+RawEntityOffsetComponentsCount])
	entityGen := w.entities[entityOffset+RawEntityOffsetGen]
	// dead entity.
	if entityGen < 0 {
		return
	}
	// kill components.
	if componentsCount > 0 {
		for i := entityOffset + RawEntityOffsetComponents + componentsCount - 1; i >= entityOffset+RawEntityOffsetComponents; i-- {
			w.pools[w.entities[i]].Del(entity)
		}
	} else {
		if entityGen == math.MaxInt16 {
			entityGen = 1
		} else {
			entityGen = entityGen + 1
		}
		w.entities[entityOffset+RawEntityOffsetGen] = -entityGen
		w.entitiesRecycled = append(w.entitiesRecycled, entity)
		if DEBUG {
			for _, l := range w.debugEventListeners {
				l.OnEntityDestroyed(entity)
			}
		}
	}
}

func (w *World) GetEntityGen(entity int) int16 {
	return w.entities[w.GetRawEntityOffset(entity)+RawEntityOffsetGen]
}

func (w *World) GetEntityComponentsCount(entity int) int16 {
	return w.entities[w.GetRawEntityOffset(entity)+RawEntityOffsetComponentsCount]
}

func (w *World) GetRawEntityItemSize() int {
	return w.entitiesItemSize
}

func (w *World) GetComponentValues(entity int, list []any) []any {
	entityOffset := w.GetRawEntityOffset(entity)
	itemsCount := int(w.entities[entityOffset+RawEntityOffsetComponentsCount])
	if itemsCount > 0 {
		dataOffset := entityOffset + RawEntityOffsetComponents
		for i := 0; i < itemsCount; i++ {
			list = append(list, w.pools[w.entities[dataOffset+i]].GetRaw(entity))
		}
	}
	return list
}

func (w *World) CopyEntity(srcEntity, dstEntity int) {
	entityOffset := w.GetRawEntityOffset(srcEntity)
	itemsCount := int(w.entities[entityOffset+RawEntityOffsetComponentsCount])
	if itemsCount > 0 {
		dataOffset := entityOffset + RawEntityOffsetComponents
		for i := 0; i < itemsCount; i++ {
			w.pools[w.entities[dataOffset+i]].Copy(srcEntity, dstEntity)
		}
	}
}

func (w *World) GetComponentTypes(entity int, list []reflect.Type) []reflect.Type {
	entityOffset := w.GetRawEntityOffset(entity)
	itemsCount := int(w.entities[entityOffset+RawEntityOffsetComponentsCount])
	if itemsCount > 0 {
		dataOffset := entityOffset + RawEntityOffsetComponents
		for i := 0; i < itemsCount; i++ {
			list = append(list, w.pools[w.entities[dataOffset+i]].GetItemType())
		}
	}
	return list
}

func (w *World) GetWorldSize() int {
	return cap(w.entities) / w.entitiesItemSize
}

func (w *World) GetRawEntities() []int16 {
	return w.entities
}

func DebugGetPoolsPtr(w *World) *[]IPool {
	return &w.pools
}

func (w *World) addComponentToRawEntity(entity int, poolId int16) {
	offset := w.GetRawEntityOffset(entity)
	dataCount := int(w.entities[offset+RawEntityOffsetComponentsCount])
	if dataCount+RawEntityOffsetComponents == w.entitiesItemSize {
		// resize entities.
		w.extendEntitiesCache()
		offset = w.GetRawEntityOffset(entity)
	}
	w.entities[offset+RawEntityOffsetComponentsCount]++
	w.entities[offset+RawEntityOffsetComponents+dataCount] = poolId
}

func (w *World) removeComponentFromRawEntity(entity int, poolId int16) {
	offset := w.GetRawEntityOffset(entity)
	dataCount := int(w.entities[offset+RawEntityOffsetComponentsCount])
	dataCount--
	w.entities[offset+RawEntityOffsetComponentsCount] = int16(dataCount)
	dataOffset := offset + RawEntityOffsetComponents
	for i := 0; i <= dataCount; i++ {
		if w.entities[dataOffset+i] == poolId {
			if i < dataCount {
				// fill gap with last item.
				w.entities[dataOffset+i] = w.entities[dataOffset+dataCount]
			}
			break
		}
	}
}

func (w *World) extendEntitiesCache() {
	entitiesCount := len(w.entities) / w.entitiesItemSize
	newItemSize := RawEntityOffsetComponents + ((w.entitiesItemSize - RawEntityOffsetComponents) << 1)
	newEntities := make([]int16, entitiesCount*newItemSize, w.GetWorldSize()*newItemSize)
	oldOffset := 0
	newOffset := 0
	for i := 0; i < entitiesCount; i++ {
		// amount of entity data (components + header).
		entityDataLen := int(w.entities[oldOffset+RawEntityOffsetComponentsCount]) + RawEntityOffsetComponents
		for j := 0; j < entityDataLen; j++ {
			newEntities[newOffset+j] = w.entities[oldOffset+j]
		}
		oldOffset += w.entitiesItemSize
		newOffset += newItemSize
	}
	w.entitiesItemSize = newItemSize
	w.entities = newEntities
}

func GetPool[T any](w *World) *Pool[T] {
	itemType := reflect.TypeOf((*T)(nil))
	if pool, ok := w.poolsHashes[itemType]; ok {
		return pool.(*Pool[T])
	}
	if DEBUG {
		if len(w.pools) == math.MaxInt16 {
			panic("no more room for new component into this world")
		}
	}
	pool := newPool[T](w, int16(len(w.pools)), w.config.PoolDenseSize, w.GetWorldSize(), w.config.PoolRecycledSize)
	w.poolsHashes[itemType] = pool
	w.pools = append(w.pools, pool)
	w.filtersByIncludes = append(w.filtersByIncludes, nil)
	w.filtersByExcludes = append(w.filtersByExcludes, nil)
	return pool
}

func debugCheckWorldForLeakedEntities(w *World) bool {
	if len(w.debugLeakedEntities) > 0 {
		for _, leakedEntity := range w.debugLeakedEntities {
			entityData := w.GetRawEntityOffset(leakedEntity)
			if w.entities[entityData+RawEntityOffsetGen] > 0 && w.entities[entityData+RawEntityOffsetComponentsCount] == 0 {
				w.debugLeakedEntities = w.debugLeakedEntities[:0]
				return true
			}
		}
		w.debugLeakedEntities = w.debugLeakedEntities[:0]
	}
	return false
}

func (w *World) checkEntityAlive(entity int) bool {
	return entity >= 0 && (entity*w.entitiesItemSize) < len(w.entities) && w.entities[w.GetRawEntityOffset(entity)+RawEntityOffsetGen] > 0
}
