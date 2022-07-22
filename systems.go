// ----------------------------------------------------------------------------
// The Proprietary or MIT-Red License
// Copyright (c) 2012-2022 Leopotam <leopotam@yandex.ru>
// ----------------------------------------------------------------------------

package ecs // import "leopotam.com/go/ecs"

import (
	"fmt"
	"reflect"
)

type IPreInitSystem interface {
	PreInit(systems ISystems)
}

type IInitSystem interface {
	Init(systems ISystems)
}

type IRunSystem interface {
	Run(systems ISystems)
}

type IDestroySystem interface {
	Destroy(systems ISystems)
}

type IPostDestroySystem interface {
	PostDestroy(systems ISystems)
}

type ISystems interface {
	Add(system any) ISystems
	GetAllSystems() []any
	AddWorld(world *World, name string) ISystems
	GetWorld(name string) *World
	GetNamedWorlds() map[string]*World
	Init()
	Run()
	Destroy()
}

type systems struct {
	defWorld    *World
	namedWorlds map[string]*World
	all         []any
	run         []IRunSystem
}

func NewSystems(world *World) ISystems {
	return &systems{
		defWorld:    world,
		namedWorlds: make(map[string]*World, 4),
		all:         make([]any, 0, 128),
		run:         make([]IRunSystem, 0, 128),
	}
}

func (s *systems) Add(system any) ISystems {
	if DEBUG {
		switch system.(type) {
		case IPreInitSystem:
		case IInitSystem:
		case IRunSystem:
		case IDestroySystem:
		case IPostDestroySystem:
		default:
			panic(fmt.Sprintf("invalid system type \"%s\"", reflect.TypeOf(system).String()))
		}
	}
	s.all = append(s.all, system)
	if runSystem, ok := system.(IRunSystem); ok {
		s.run = append(s.run, runSystem)
	}
	return s
}

func (s *systems) GetAllSystems() []any {
	return s.all
}

func (s *systems) AddWorld(world *World, name string) ISystems {
	if DEBUG {
		if _, ok := s.namedWorlds[name]; ok {
			panic(fmt.Sprintf("world with name \"%s\" already added", name))
		}
	}
	s.namedWorlds[name] = world
	return s
}

func (s *systems) GetWorld(name string) *World {
	if len(name) == 0 {
		return s.defWorld
	}
	if world, ok := s.namedWorlds[name]; ok {
		return world
	}
	return nil
}

func (s *systems) GetNamedWorlds() map[string]*World {
	return s.namedWorlds
}

func (s *systems) Init() {
	for _, system := range s.all {
		if preInitSystem, ok := system.(IPreInitSystem); ok {
			preInitSystem.PreInit(s)
			if DEBUG {
				worldName := debugCheckSystemsForLeakedEntities(s)
				if len(worldName) > 0 {
					panic(fmt.Sprintf("empty entity detected in world \"%s\" after {%s}.PreInit()", worldName, reflect.TypeOf(system).String()))
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
					panic(fmt.Sprintf("empty entity detected in world \"%s\" after {%s}.Init()", worldName, reflect.TypeOf(system).String()))
				}
			}
		}
	}
}

func (s *systems) Run() {
	for _, system := range s.run {
		system.Run(s)
		if DEBUG {
			worldName := debugCheckSystemsForLeakedEntities(s)
			if len(worldName) > 0 {
				panic(fmt.Sprintf("empty entity detected in world \"%s\" after %s.Run()", worldName, reflect.TypeOf(system).String()))
			}
		}
	}
}

func (s *systems) Destroy() {
	for i := len(s.all) - 1; i >= 0; i-- {
		if destroySystem, ok := s.all[i].(IDestroySystem); ok {
			destroySystem.Destroy(s)
			if DEBUG {
				worldName := debugCheckSystemsForLeakedEntities(s)
				if len(worldName) > 0 {
					panic(fmt.Sprintf("empty entity detected in world \"%s\" after %s.Destroy()", worldName, reflect.TypeOf(destroySystem).String()))
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
					panic(fmt.Sprintf("empty entity detected in world \"%s\" after %s.PostDestroy()", worldName, reflect.TypeOf(postDestroySystem).String()))
				}
			}
		}
	}
	for k := range s.namedWorlds {
		delete(s.namedWorlds, k)
	}
	s.all = s.all[:0]
	s.run = s.run[:0]
}

func debugCheckSystemsForLeakedEntities(s *systems) string {
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
