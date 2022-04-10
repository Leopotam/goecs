// ----------------------------------------------------------------------------
// The Proprietary or MIT-Red License
// Copyright (c) 2012-2022 Leopotam <leopotam@yandex.ru>
// ----------------------------------------------------------------------------

package ecs

import (
	"fmt"
	"reflect"
)

type IPreInitSystem interface {
	PreInit(systems *Systems)
}

type IInitSystem interface {
	Init(systems *Systems)
}

type IRunSystem interface {
	Run(systems *Systems)
}

type IDestroySystem interface {
	Destroy(systems *Systems)
}

type IPostDestroySystem interface {
	PostDestroy(systems *Systems)
}

type Systems struct {
	defWorld *World
	all      []any
	run      []IRunSystem
}

func NewSystems(world *World) *Systems {
	return &Systems{
		defWorld: world,
		all:      make([]any, 0, 128),
		run:      make([]IRunSystem, 0, 128),
	}
}

func (s *Systems) Add(system any) *Systems {
	if DEBUG {
		switch system.(type) {
		case IPreInitSystem:
		case IInitSystem:
		case IRunSystem:
		case IDestroySystem:
		case IPostDestroySystem:
		default:
			panic(fmt.Sprintf("invalid system type \"%s\".", reflect.TypeOf(system).Name()))
		}
	}
	s.all = append(s.all, system)
	if runSystem, ok := system.(IRunSystem); ok {
		s.run = append(s.run, runSystem)
	}
	return s
}

func (s *Systems) Init() {
	for _, system := range s.all {
		if preInitSystem, ok := system.(IPreInitSystem); ok {
			preInitSystem.PreInit(s)
		}
	}
	for _, system := range s.all {
		if initSystem, ok := system.(IInitSystem); ok {
			initSystem.Init(s)
		}
	}
}

func (s *Systems) Run() {
	for _, system := range s.run {
		system.Run(s)
	}
}

func (s *Systems) Destroy() {
	for i := len(s.all) - 1; i >= 0; i-- {
		if destroySystem, ok := s.all[i].(IDestroySystem); ok {
			destroySystem.Destroy(s)
		}
	}
	for i := len(s.all) - 1; i >= 0; i-- {
		if postDestroySystem, ok := s.all[i].(IPostDestroySystem); ok {
			postDestroySystem.PostDestroy(s)
		}
	}
}
