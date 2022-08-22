// ----------------------------------------------------------------------------
// The Proprietary or MIT-Red License
// Copyright (c) 2012-2022 Leopotam <leopotam@yandex.ru>
// ----------------------------------------------------------------------------

package ecs // import "leopotam.com/go/ecs"

import (
	"sort"
)

type int16Slice []int16

func (x int16Slice) Len() int           { return len(x) }
func (x int16Slice) Less(i, j int) bool { return x[i] < x[j] }
func (x int16Slice) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

type mask struct {
	include []int16
	exclude []int16
	hash    int
}

type delayedOp struct {
	added  bool
	entity int
}

type IInc interface {
	FillIncludes(w *World, list []int16) []int16
	FillPools(w *World) IInc
}

type IExc interface {
	FillExcludes(w *World, list []int16) []int16
}

type FilterIter struct {
	f      *Filter
	locked bool
	denses []int
	length int
	idx    int
	entity int
}

func (i *FilterIter) Next() bool {
	i.idx++
	if i.idx >= i.length {
		i.Destroy()
		return false
	}
	i.entity = i.denses[i.idx]
	return true
}

func (i *FilterIter) GetEntity() int {
	return i.entity
}

func (i *FilterIter) Destroy() {
	if i.locked {
		i.locked = false
		i.f.unlock()
	}
}

type Filter struct {
	world   *World
	mask    *mask
	densed  []int
	sparsed []int
	delayed []delayedOp
	locks   int
}

func newFilter(w *World, mask *mask, denseCapacity int, sparseCapacity int) *Filter {
	f := &Filter{
		world:   w,
		mask:    mask,
		densed:  make([]int, 0, denseCapacity),
		sparsed: make([]int, sparseCapacity),
		delayed: make([]delayedOp, 0, 512),
	}
	w.filtersHashes[mask.hash] = f
	for _, v := range mask.include {
		l := w.filtersByIncludes[v]
		l = append(l, f)
		w.filtersByIncludes[v] = l
	}
	for _, v := range mask.exclude {
		l := w.filtersByExcludes[v]
		l = append(l, f)
		w.filtersByExcludes[v] = l
	}
	// scan exist entities for compatibility with new filter.
	for i, iMax := 0, len(w.entities)/w.entitiesItemSize; i < iMax; i++ {
		if w.entities[w.GetRawEntityOffset(i)+RawEntityOffsetComponentsCount] > 0 && w.isMaskCompatible(mask, i) {
			f.addEntity(i)
		}
	}
	return f
}

func (w *World) isMaskCompatible(m *mask, entity int) bool {
	for _, v := range m.include {
		if !w.pools[v].Has(entity) {
			return false
		}
	}
	for _, v := range m.exclude {
		if w.pools[v].Has(entity) {
			return false
		}
	}
	return true
}

func (w *World) isMaskCompatibleWithout(m *mask, entity int, componentID int16) bool {
	for _, typeID := range m.include {
		if typeID == componentID || !w.pools[typeID].Has(entity) {
			return false
		}
	}
	for _, typeID := range m.exclude {
		if typeID != componentID && w.pools[typeID].Has(entity) {
			return false
		}
	}
	return true
}

func (w *World) onEntityChange(entity int, componentID int16, added bool) {
	var includeList = w.filtersByIncludes[componentID]
	var excludeList = w.filtersByExcludes[componentID]
	if added {
		// add component.
		for _, filter := range includeList {
			if w.isMaskCompatible(filter.getMask(), entity) {
				if DEBUG && filter.sparsed[entity] > 0 {
					panic("entity already in filter")
				}
				filter.addEntity(entity)
			}
		}
		for _, filter := range excludeList {
			if w.isMaskCompatibleWithout(filter.getMask(), entity, componentID) {
				if DEBUG && filter.sparsed[entity] == 0 {
					panic("entity not in filter")
				}
				filter.removeEntity(entity)
			}
		}
	} else {
		// remove component.
		for _, filter := range includeList {
			if w.isMaskCompatible(filter.getMask(), entity) {
				if DEBUG && filter.sparsed[entity] == 0 {
					panic("entity not in filter")
				}
				filter.removeEntity(entity)
			}
		}
		for _, filter := range excludeList {
			if w.isMaskCompatibleWithout(filter.getMask(), entity, componentID) {
				if DEBUG && filter.sparsed[entity] > 0 {
					panic("entity already in filter")
				}
				filter.addEntity(entity)
			}
		}
	}
}

func (f *Filter) resizeSparseIndex(capacity int) {
	ss := make([]int, capacity)
	copy(ss, f.sparsed)
	f.sparsed = ss
}

func (f *Filter) addEntity(entity int) {
	if f.locks == 0 {
		f.densed = append(f.densed, entity)
		f.sparsed[entity] = len(f.densed)
	} else {
		f.delayed = append(f.delayed, delayedOp{added: true, entity: entity})
	}
}

func (f *Filter) removeEntity(entity int) {
	if f.locks == 0 {
		idx := f.sparsed[entity] - 1
		f.sparsed[entity] = 0
		l := len(f.densed) - 1
		if idx < l {
			f.densed[idx] = f.densed[l]
			f.sparsed[f.densed[idx]] = idx + 1
		}
		f.densed = f.densed[:l]
	} else {
		f.delayed = append(f.delayed, delayedOp{added: false, entity: entity})
	}
}

func (f *Filter) unlock() {
	f.locks--
	if f.locks == 0 && len(f.delayed) > 0 {
		for _, op := range f.delayed {
			if op.added {
				f.addEntity(op.entity)
			} else {
				f.removeEntity(op.entity)
			}
		}
		f.delayed = f.delayed[:0]
	}
}

func (f *Filter) getMask() *mask {
	return f.mask
}

func (f *Filter) Iter() FilterIter {
	f.locks++
	return FilterIter{
		f:      f,
		locked: true,
		denses: f.densed,
		length: len(f.densed),
		idx:    -1,
	}
}

func (f *Filter) GetWorld() *World {
	return f.world
}

func (f *Filter) GetEntitiesCount() int {
	return len(f.densed)
}

func (f *Filter) GetRawEntities() []int {
	return f.densed
}

func (f *Filter) GetSparseIndices() []int {
	return f.sparsed
}

func (w *World) requestMaskCache() []int16 {
	var c []int16
	if l := len(w.filterMaskCache); l > 0 {
		l--
		c = w.filterMaskCache[l]
		w.filterMaskCache[l] = nil
		w.filterMaskCache = w.filterMaskCache[:l]
	} else {
		c = make([]int16, 0, 16)
	}
	return c
}

func (w *World) recycleMaskCache(c []int16) {
	c = c[:0]
	w.filterMaskCache = append(w.filterMaskCache, c)
}

func GetFilter[I IInc](w *World) *Filter {
	var i I
	inc := i.FillIncludes(w, w.requestMaskCache())
	sort.Sort(int16Slice(inc))
	hash := len(inc)
	for _, v := range inc {
		hash = hash*314159 + int(v)
	}
	if f, ok := w.filtersHashes[hash]; ok {
		w.recycleMaskCache(inc)
		return f
	}
	return newFilter(w, &mask{include: inc, hash: hash}, w.config.PoolDenseSize, w.GetWorldSize())
}

func GetFilterWithExc[I IInc, E IExc](w *World) *Filter {
	var i I
	var e E
	inc := i.FillIncludes(w, w.requestMaskCache())
	exc := e.FillExcludes(w, w.requestMaskCache())
	sort.Sort(int16Slice(inc))
	sort.Sort(int16Slice(exc))
	hash := len(inc) + len(exc)
	for _, v := range inc {
		hash = hash*314159 + int(v)
	}
	for _, v := range exc {
		hash = hash*314159 - int(v)
	}
	if f, ok := w.filtersHashes[hash]; ok {
		w.recycleMaskCache(inc)
		w.recycleMaskCache(exc)
		return f
	}
	return newFilter(w, &mask{include: inc, exclude: exc, hash: hash}, w.config.PoolDenseSize, w.GetWorldSize())
}

type Inc1[I1 any] struct {
	Inc1 *Pool[I1]
}

func (i Inc1[I1]) FillIncludes(w *World, list []int16) []int16 {
	return append(list, GetPool[I1](w).GetID())
}

func (i Inc1[I1]) FillPools(w *World) IInc {
	return &Inc1[I1]{
		Inc1: GetPool[I1](w),
	}
}

type Inc2[I1 any, I2 any] struct {
	Inc1 *Pool[I1]
	Inc2 *Pool[I2]
}

func (i Inc2[I1, I2]) FillIncludes(w *World, list []int16) []int16 {
	list = append(list, GetPool[I1](w).GetID())
	return append(list, GetPool[I2](w).GetID())
}

func (i Inc2[I1, I2]) FillPools(w *World) IInc {
	return &Inc2[I1, I2]{
		Inc1: GetPool[I1](w),
		Inc2: GetPool[I2](w),
	}
}

type Inc3[I1 any, I2 any, I3 any] struct {
	Inc1 *Pool[I1]
	Inc2 *Pool[I2]
	Inc3 *Pool[I3]
}

func (i Inc3[I1, I2, I3]) FillIncludes(w *World, list []int16) []int16 {
	list = append(list, GetPool[I1](w).GetID())
	list = append(list, GetPool[I2](w).GetID())
	return append(list, GetPool[I3](w).GetID())
}

func (i Inc3[I1, I2, I3]) FillPools(w *World) IInc {
	return &Inc3[I1, I2, I3]{
		Inc1: GetPool[I1](w),
		Inc2: GetPool[I2](w),
		Inc3: GetPool[I3](w),
	}
}

type Inc4[I1 any, I2 any, I3 any, I4 any] struct {
	Inc1 *Pool[I1]
	Inc2 *Pool[I2]
	Inc3 *Pool[I3]
	Inc4 *Pool[I4]
}

func (i Inc4[I1, I2, I3, I4]) FillIncludes(w *World, list []int16) []int16 {
	list = append(list, GetPool[I1](w).GetID())
	list = append(list, GetPool[I2](w).GetID())
	list = append(list, GetPool[I3](w).GetID())
	return append(list, GetPool[I4](w).GetID())
}

func (i Inc4[I1, I2, I3, I4]) FillPools(w *World) IInc {
	return &Inc4[I1, I2, I3, I4]{
		Inc1: GetPool[I1](w),
		Inc2: GetPool[I2](w),
		Inc3: GetPool[I3](w),
		Inc4: GetPool[I4](w),
	}
}

type Inc5[I1 any, I2 any, I3 any, I4 any, I5 any] struct {
	Inc1 *Pool[I1]
	Inc2 *Pool[I2]
	Inc3 *Pool[I3]
	Inc4 *Pool[I4]
	Inc5 *Pool[I5]
}

func (i Inc5[I1, I2, I3, I4, I5]) FillIncludes(w *World, list []int16) []int16 {
	list = append(list, GetPool[I1](w).GetID())
	list = append(list, GetPool[I2](w).GetID())
	list = append(list, GetPool[I3](w).GetID())
	list = append(list, GetPool[I4](w).GetID())
	return append(list, GetPool[I5](w).GetID())
}

func (i Inc5[I1, I2, I3, I4, I5]) FillPools(w *World) IInc {
	return &Inc5[I1, I2, I3, I4, I5]{
		Inc1: GetPool[I1](w),
		Inc2: GetPool[I2](w),
		Inc3: GetPool[I3](w),
		Inc4: GetPool[I4](w),
		Inc5: GetPool[I5](w),
	}
}

type Inc6[I1 any, I2 any, I3 any, I4 any, I5 any, I6 any] struct {
	Inc1 *Pool[I1]
	Inc2 *Pool[I2]
	Inc3 *Pool[I3]
	Inc4 *Pool[I4]
	Inc5 *Pool[I5]
	Inc6 *Pool[I6]
}

func (i Inc6[I1, I2, I3, I4, I5, I6]) FillIncludes(w *World, list []int16) []int16 {
	list = append(list, GetPool[I1](w).GetID())
	list = append(list, GetPool[I2](w).GetID())
	list = append(list, GetPool[I3](w).GetID())
	list = append(list, GetPool[I4](w).GetID())
	list = append(list, GetPool[I5](w).GetID())
	return append(list, GetPool[I6](w).GetID())
}

func (i Inc6[I1, I2, I3, I4, I5, I6]) FillPools(w *World) IInc {
	return &Inc6[I1, I2, I3, I4, I5, I6]{
		Inc1: GetPool[I1](w),
		Inc2: GetPool[I2](w),
		Inc3: GetPool[I3](w),
		Inc4: GetPool[I4](w),
		Inc5: GetPool[I5](w),
		Inc6: GetPool[I6](w),
	}
}

type Exc1[E1 any] struct{}

func (e Exc1[E1]) FillExcludes(w *World, list []int16) []int16 {
	return append(list, GetPool[E1](w).GetID())
}

type Exc2[E1 any, E2 any] struct{}

func (e Exc2[E1, E2]) FillExcludes(w *World, list []int16) []int16 {
	list = append(list, GetPool[E1](w).GetID())
	return append(list, GetPool[E2](w).GetID())
}

type Exc3[E1 any, E2 any, E3 any] struct{}

func (e Exc3[E1, E2, E3]) FillExcludes(w *World, list []int16) []int16 {
	list = append(list, GetPool[E1](w).GetID())
	list = append(list, GetPool[E2](w).GetID())
	return append(list, GetPool[E3](w).GetID())
}
