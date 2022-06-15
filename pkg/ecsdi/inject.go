// ----------------------------------------------------------------------------
// The Proprietary or MIT-Red License
// Copyright (c) 2012-2022 Leopotam <leopotam@yandex.ru>
// ----------------------------------------------------------------------------

package ecsdi // import "leopotam.com/go/ecs/pkg/ecsdi"

import (
	"fmt"
	"reflect"

	"leopotam.com/go/ecs"
)

type iBuiltinInject interface {
	fill(systems *ecs.Systems, tag string)
}
type iCustomInject interface {
	fill(injects []any)
}

type World struct {
	Value *ecs.World
}

func (w *World) fill(systems *ecs.Systems, tag string) {
	if len(tag) > 0 {
		w.Value = systems.GetWorldWithName(tag)
	} else {
		w.Value = systems.GetWorld()
	}
}

type Pool[T any] struct {
	Value *ecs.Pool[T]
}

func (p *Pool[T]) fill(systems *ecs.Systems, tag string) {
	var w *ecs.World
	if len(tag) > 0 {
		w = systems.GetWorldWithName(tag)
		if ecs.DEBUG {
			if w == nil {
				panic(fmt.Sprintf("cant get Filter[%s] from undefined world with name \"%s\"", reflect.TypeOf((*T)(nil)).Elem().String(), tag))
			}
		}
	} else {
		w = systems.GetWorld()
	}
	p.Value = ecs.GetPool[T](w)
}

func (p *Pool[T]) NewEntity() (*T, int) {
	e := p.Value.GetWorld().NewEntity()
	return p.Value.Add(e), e
}

type Filter[Inc ecs.IInc] struct {
	Value *ecs.Filter
	Pools *Inc
}

//lint:ignore U1000 called with reflection
func (f *Filter[Inc]) fill(systems *ecs.Systems, tag string) {
	var w *ecs.World
	if len(tag) > 0 {
		w = systems.GetWorldWithName(tag)
		if ecs.DEBUG {
			if w == nil {
				panic(fmt.Sprintf("cant get Filter[%s] from undefined world with name \"%s\"", reflect.TypeOf((*Inc)(nil)).Elem().String(), tag))
			}
		}
	} else {
		w = systems.GetWorld()
	}
	var inc Inc
	f.Pools = any(inc.FillPools(w)).(*Inc)
	f.Value = ecs.GetFilter[Inc](w)
}

type FilterWithExc[Inc ecs.IInc, Exc ecs.IExc] struct {
	Value *ecs.Filter
	Pools *Inc
}

//lint:ignore U1000 called with reflection
func (q *FilterWithExc[Inc, Exc]) fill(systems *ecs.Systems, tag string) {
	var w *ecs.World
	if len(tag) > 0 {
		w = systems.GetWorldWithName(tag)
		if ecs.DEBUG {
			if w == nil {
				panic(fmt.Sprintf(
					"cant get QueryWithExc[%s, %s] from undefined world with name \"%s\"",
					reflect.TypeOf((*Inc)(nil)).Elem().String(),
					reflect.TypeOf((*Exc)(nil)).Elem().String(),
					tag))
			}
		}
	} else {
		w = systems.GetWorld()
	}
	var inc Inc
	q.Pools = any(inc.FillPools(w)).(*Inc)
	q.Value = ecs.GetFilterWithExc[Inc, Exc](w)
}

type Custom[T any] struct {
	Value *T
}

//lint:ignore U1000 called with reflection
func (c *Custom[T]) fill(injects []any) {
	for _, inj := range injects {
		if casted, ok := inj.(*T); ok {
			c.Value = casted
			break
		}
	}
}

func Inject(systems *ecs.Systems, injects ...any) *ecs.Systems {
	injectsLen := len(injects)
	for _, s := range systems.GetAllSystems() {
		sType := reflect.TypeOf(s).Elem()
		sValue := reflect.ValueOf(s)
		for i := 0; i < sType.NumField(); i++ {
			fValue := sValue.Elem().Field(i)
			if fValue.CanAddr() && fValue.CanInterface() {
				fValuePtr := fValue.Addr().Interface()
				fType := sType.Field(i)
				if inj, ok := fValuePtr.(iBuiltinInject); ok {
					inj.fill(systems, fType.Tag.Get("ecsdi"))
					continue
				}
				if injectsLen > 0 {
					if inj, ok := fValuePtr.(iCustomInject); ok {
						inj.fill(injects)
						continue
					}
				}
			}
		}
	}
	return systems
}
