// ----------------------------------------------------------------------------
// The Proprietary or MIT-Red License
// Copyright (c) 2012-2022 Leopotam <leopotam@yandex.ru>
// ----------------------------------------------------------------------------

package goecs

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
	defWorld    *World
	namedWorlds map[string]*World
	all         []any
	run         []IRunSystem
}

func NewSystems(world *World) *Systems {
	return &Systems{
		defWorld:    world,
		namedWorlds: make(map[string]*World, 4),
		all:         make([]any, 0, 128),
		run:         make([]IRunSystem, 0, 128),
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
			panic(fmt.Sprintf("invalid system type \"%s\".", reflect.TypeOf(system).String()))
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
			if DEBUG {
				worldName := debugCheckSystemsForLeakedEntities(s)
				if len(worldName) > 0 {
					panic(fmt.Sprintf("Empty entity detected in world \"%s\" after {%s}.PreInit().", worldName, reflect.TypeOf(system).String()))
				}
			}
		}
	}
	for _, system := range s.all {
		if initSystem, ok := system.(IInitSystem); ok {
			initSystem.Init(s)
			if DEBUG {
				worldName := debugCheckSystemsForLeakedEntities(s)
				if len(worldName) > 0 {
					panic(fmt.Sprintf("Empty entity detected in world \"%s\" after {%s}.Init().", worldName, reflect.TypeOf(system).String()))
				}
			}
		}
	}
}

func (s *Systems) Run() {
	for _, system := range s.run {
		system.Run(s)
		if DEBUG {
			worldName := debugCheckSystemsForLeakedEntities(s)
			if len(worldName) > 0 {
				panic(fmt.Sprintf("Empty entity detected in world \"%s\" after %s.Run().", worldName, reflect.TypeOf(system).String()))
			}
		}
	}
}

func (s *Systems) Destroy() {
	for i := len(s.all) - 1; i >= 0; i-- {
		if destroySystem, ok := s.all[i].(IDestroySystem); ok {
			destroySystem.Destroy(s)
			if DEBUG {
				worldName := debugCheckSystemsForLeakedEntities(s)
				if len(worldName) > 0 {
					panic(fmt.Sprintf("Empty entity detected in world \"%s\" after %s.Destroy().", worldName, reflect.TypeOf(destroySystem).String()))
				}
			}
		}
	}
	for i := len(s.all) - 1; i >= 0; i-- {
		if postDestroySystem, ok := s.all[i].(IPostDestroySystem); ok {
			postDestroySystem.PostDestroy(s)
			if DEBUG {
				worldName := debugCheckSystemsForLeakedEntities(s)
				if len(worldName) > 0 {
					panic(fmt.Sprintf("Empty entity detected in world \"%s\" after %s.PostDestroy().", worldName, reflect.TypeOf(postDestroySystem).String()))
				}
			}
		}
	}
	s.all = s.all[:0]
	s.run = s.run[:0]
}

func (s *Systems) AddWorld(world *World, name string) *Systems {
	if DEBUG {
		if _, ok := s.namedWorlds[name]; ok {
			panic(fmt.Sprintf("World with name \"%s\" already added.", name))
		}
	}
	s.namedWorlds[name] = world
	return s
}

func (s *Systems) GetWorld() *World {
	return s.defWorld
}

func (s *Systems) GetWorldWithName(name string) *World {
	if world, ok := s.namedWorlds[name]; ok {
		return world
	}
	return nil
}

func debugCheckSystemsForLeakedEntities(s *Systems) string {
	if DEBUG {
		if debugCheckWorldForLeakedEntities(s.defWorld) {
			return "default"
		}
		for name, world := range s.namedWorlds {
			if debugCheckWorldForLeakedEntities(world) {
				return name
			}
		}
	}
	return ""
}
