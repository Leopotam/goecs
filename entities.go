// ----------------------------------------------------------------------------
// The Proprietary or MIT-Red License
// Copyright (c) 2012-2022 Leopotam <leopotam@yandex.ru>
// ----------------------------------------------------------------------------

package goecs

type PackedEntity struct {
	id  int
	gen int16
}

type PackedEntityWithWorld struct {
	id  int
	gen int16
	w   *World
}

func (w *World) PackEntity(entity int) PackedEntity {
	return PackedEntity{id: entity, gen: w.GetEntityGen(entity)}
}

func (pe PackedEntity) Unpack(w *World) (int, bool) {
	if !w.checkEntityAlive(pe.id) || w.GetEntityGen(pe.id) != int16(pe.gen) {
		return 0, false
	}
	return pe.id, true
}

func (w *World) PackEntityWithWorld(entity int) PackedEntityWithWorld {
	return PackedEntityWithWorld{id: entity, gen: w.GetEntityGen(entity), w: w}
}

func (pe PackedEntityWithWorld) Unpack() (*World, int, bool) {
	if pe.w == nil || !pe.w.checkEntityAlive(pe.id) || pe.w.GetEntityGen(pe.id) != pe.gen {
		return nil, 0, false
	}
	return pe.w, pe.id, true
}
