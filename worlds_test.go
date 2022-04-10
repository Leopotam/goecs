// ----------------------------------------------------------------------------
// The Proprietary or MIT-Red License
// Copyright (c) 2012-2022 Leopotam <leopotam@yandex.ru>
// ----------------------------------------------------------------------------

package ecs_test

import (
	"ecs"
	"testing"
)

type C1 struct{}
type C2 struct{ ID int }
type C3 struct{}
type C4 struct{}

func (c2 *C2) Reset() {
	c2.ID = -1
}

func TestWorld(t *testing.T) {
	w := ecs.NewWorld()
	w.Destroy()
}
