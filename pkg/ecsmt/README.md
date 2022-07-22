# GoECS MT - Поддержка многопоточной обработки
Обеспечивает поддержку обработки сущностей в несколько системных потоков для GoECS.

# Содержание
* [Социальные ресурсы](#Социальные-ресурсы)
* [Установка](#Установка)
* [Специальные типы](#Специальные-типы)
    * [Задачи](#Задачи)
    * [Отложенные операции](#Отложенные-операции)
* [Лицензия](#Лицензия)

# Социальные ресурсы
[![discord](https://img.shields.io/discord/404358247621853185.svg?label=enter%20to%20discord%20server&style=for-the-badge&logo=discord)](https://discord.gg/5GZVde6)

# Установка
Поддерживается установка штатным модулем:
```
go get -u leopotam.com/go/ecs/pkg/ecsmt
```
По умолчанию используется последняя релизная версия. Если требуется версия "в разработке" с актуальными изменениями - следует скопировать хеш нужного коммита из ветки `develop` и подставить в командную строку. Например:
```
go get -u leopotam.com/go/ecs/pkg/ecsmt@830f682
```
После скачивания пакет будет доступен как `"leopotam.com/go/ecs/pkg/ecsmt"`.

# Специальные типы

## Задачи
Предназначены для распараллеливания обработки сущностей в фильтре через вызов `ecsmt.RunTask()`.
```go
// Компонент.
type c1 struct {
    id int
}
// Система.
type System1 struct {
    filter *ecs.Filter
    pool *ecs.Pool[c1]
}
func (s *System1) Init(systems ecs.ISystems) {
    w := systems.GetWorld()
    s.filter = GetFilter[ecs.Inc1[C1]](w)
    s.pool = GetPool[C1](w)
}
func (s *System1) Run(systems ecs.ISystems) {
    // Размер блока данных, после которого наступает разделение на несколько задач.
    chunkSize := 1000
    ecsmt.RunTask(s, s.filter, chunkSize)
}
// Обработчик для каждого блока сущностей.
func (s *System1) Process(entities []int, fromIdx, beforeIdx int) {
    // Можно итерироваться по entities строго от fromIdx и до beforeIdx.
	for idx := fromIdx; idx < beforeIdx; idx++ {
		c1 := s.pool.Get(entities[idx])
        c1.counter = (c1.counter + 1) % 10000
	}
}
```

> **ВАЖНО!** Внутри обработчика **запрещено** изменять состояние мира стандартным апи `World` и `Pool`: нельзя создавать / удалять сущности, нельзя добавлять / удалять компоненты на сущности. Допускается только модификация данных внутри существующих компонентов.

## Отложенные операции
Позволяют модифицировать мир не мгновенно, а с отложенным выполнением, могут быть использованы в [задачах](#Задачи) для создания/удаления сущностей и компонентов.

```go
type c1 struct {}
type c2 struct { Source c1 }
type delayedSystem struct {
	c1DelayedPool *ecsmt.DelayedPool[c1]
    c2DelayedPool *ecsmt.DelayedPool[c2]
	delayedBuffer ecsmt.IDelayedBuffer
	World         ecsdi.World
	Filter        ecsdi.Filter[ecs.Inc1[c1]]
}

func (s *delayedSystem) Init(systems ecs.ISystems) {
    // Подготовим все необходимое для многопоточной обработки данных.

	// Добавим сущность с компонентом, чтобы фильтр не был пустым.
	s.Filter.Pools.Inc1.Add(s.World.Value.NewEntity())
    // Добавлять и удалять компоненты можно только через эти пулы.
	s.c1DelayedPool = ecsmt.NewDelayedPool[c1]()
    s.c2DelayedPool = ecsmt.NewDelayedPool[c2]()
    // Добавлять и удалять сущности можно только через этот буфер команд.
	s.delayedBuffer = ecsmt.NewDelayedBuffer(s.World.Value, s.c1DelayedPool, s.c2DelayedPool)
}

func (s *delayedSystem) Run(systems ecs.ISystems) {
    // Запускаем обработку задач.
	ecsmt.RunTask(s, s.Filter.Value, 10)
    // Применяем все изменения, которые накопились в буфере задач.
	s.delayedBuffer.Process()
}

func (s *delayedSystem) Process(entities []int, from, before int) {
	for i := from; i < before; i++ {
        e := entities[i]
        // Если компонент c1 есть сущности (в данном случае всегда есть)...
        if s.c1DelayedPool.Has(e) {
            // создаем новую сущность...
            evt := s.delayedBuffer.NewEntity()
            // с компонентом c2 на ней...
            c2 := s.c2DelayedPool.Add(evt)
            c2.Source = s.c1DelayedPool[e]
            // и удаляем старую сущность.
            s.delayedBuffer.DelEntity(e)
        }
	}
}
```

> **ВАЖНО!** Без вызова `IDelayedBuffer.Process()` отложенные операции не будут применены, а будут копиться и потреблять память.

# Лицензия
Фреймворк выпускается под двумя лицензиями, [подробности тут](./../../LICENSE.md).

В случаях лицензирования по условиям MIT-Red не стоит расчитывать на
персональные консультации или какие-либо гарантии.