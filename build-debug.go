// ----------------------------------------------------------------------------
// The Proprietary or MIT-Red License
// Copyright (c) 2012-2022 Leopotam <leopotam@yandex.ru>
// ----------------------------------------------------------------------------

//go:build !RELEASE

package ecs // import "leopotam.com/go/ecs"

// DEBUG mode constant.
const DEBUG = true

func (w *World) AddEventListener(l IWorldEventListener) {
	w.debugEventListeners = append(w.debugEventListeners, l)
}

func (w *World) RemoveEventListener(l IWorldEventListener) {
	i := -1
	for idx, v := range w.debugEventListeners {
		if v == l {
			i = idx
			break
		}
	}
	if i > -1 {
		copy(w.debugEventListeners[i:], w.debugEventListeners[i+1:])
		w.debugEventListeners[len(w.debugEventListeners)-1] = nil
		w.debugEventListeners = w.debugEventListeners[:len(w.debugEventListeners)-1]
	}
}
