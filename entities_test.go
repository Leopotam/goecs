// ----------------------------------------------------------------------------
// The Proprietary or MIT-Red License
// Copyright (c) 2012-2022 Leopotam <leopotam@yandex.ru>
// ----------------------------------------------------------------------------

package ecs_test

import (
	"testing"

	"leopotam.com/go/ecs"
)

func TestPackedEntity(t *testing.T) {
	w := ecs.NewWorld()
	e := w.NewEntity()
	packedEntity := w.PackEntity(e)
	if unpackedEntity, ok := packedEntity.Unpack(w); !ok || unpackedEntity != e {
		t.Errorf("invalid packed entity")
	}
	w.DelEntity(e)
	if _, ok := packedEntity.Unpack(w); ok {
		t.Errorf("invalid packed entity after removing")
	}
	w.Destroy()
}

func TestPackedEntityWithWorld(t *testing.T) {
	w := ecs.NewWorld()
	e := w.NewEntity()
	packedEntity := w.PackEntityWithWorld(e)
	if unpackedWorld, unpackedEntity, ok := packedEntity.Unpack(); !ok || unpackedWorld != w || unpackedEntity != e {
		t.Errorf("invalid packed entity")
	}
	w.DelEntity(e)
	if _, _, ok := packedEntity.Unpack(); ok {
		t.Errorf("invalid packed entity after removing")
	}
	w.Destroy()
}
