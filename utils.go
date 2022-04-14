// ----------------------------------------------------------------------------
// The Proprietary or MIT-Red License
// Copyright (c) 2012-2022 Leopotam <leopotam@yandex.ru>
// ----------------------------------------------------------------------------

package ecs

type sparseSet struct {
	dense    []int
	sparse   []int
	recycled []int
}

func newSparseSet(denseCapacity int, sparseCapacity int) *sparseSet {
	return &sparseSet{
		dense:    make([]int, 1, denseCapacity+1),
		sparse:   make([]int, sparseCapacity),
		recycled: make([]int, 0, denseCapacity+1),
	}
}

func (s *sparseSet) Resize(capacity int) {
	ss := make([]int, capacity)
	copy(ss, s.sparse)
	s.sparse = ss
}

func (s *sparseSet) Set(idx int) (denseIdx int, isNew bool) {
	if DEBUG {
		if idx < 0 || idx >= len(s.sparse) {
			panic("invalid idx")
		}
		if s.sparse[idx] != 0 {
			panic("already set")
		}
	}
	l := len(s.recycled)
	isNew = l == 0
	if isNew {
		denseIdx = len(s.dense)
		s.dense = append(s.dense, idx)
	} else {
		denseIdx = s.recycled[l-1]
		s.recycled = s.recycled[:l-1]
		s.dense[denseIdx] = idx
	}
	s.sparse[idx] = denseIdx
	return
}

func (s *sparseSet) Get(idx int) int {
	if DEBUG {
		if idx < 0 || idx >= len(s.sparse) {
			panic("Cant touch invalid entity.")
		}
		if s.sparse[idx] == 0 {
			panic("not set")
		}
	}
	return s.sparse[idx]
}

func (s *sparseSet) Has(idx int) bool {
	if DEBUG {
		if idx < 0 || idx >= len(s.sparse) {
			panic("Cant touch invalid entity.")
		}
	}
	return s.sparse[idx] > 0
}

func (s *sparseSet) Del(idx int) (denseIdx int) {
	if DEBUG {
		if idx < 0 || idx >= len(s.sparse) {
			panic("invalid idx")
		}
		if s.sparse[idx] == 0 {
			panic("not set")
		}
	}
	denseIdx = s.sparse[idx]
	s.dense[denseIdx] = -1
	s.sparse[idx] = 0
	s.recycled = append(s.recycled, denseIdx)
	return
}

func (s *sparseSet) GetIndices() []int {
	return s.dense[1:]
}
