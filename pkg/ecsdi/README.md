# GoECS DI - Поддержка автоматической инъекции данных в поля ECS-систем
Обеспечивает поддержку инъекции пользовательских и ECS-данных в поля ECS-систем для GoECS.

# Содержание
* [Социальные ресурсы](#Социальные-ресурсы)
* [Установка](#Установка)
* [Интеграция](#Интеграция)
* [Специальные типы](#Специальные-типы)
    * [World](#World)
    * [Pool](#Pool)
    * [Filter](#Filter)
    * [Custom](#Custom)
* [Лицензия](#Лицензия)

# Социальные ресурсы
[![discord](https://img.shields.io/discord/404358247621853185.svg?label=enter%20to%20discord%20server&style=for-the-badge&logo=discord)](https://discord.gg/5GZVde6)

# Установка
Поддерживается установка штатным модулем:
```
go get -u leopotam.com/go/ecs/pkg/ecsdi
```
По умолчанию используется последняя релизная версия. Если требуется версия "в разработке" с актуальными изменениями - следует скопировать хеш нужного коммита из ветки `develop` и подставить в командную строку. Например:
```
go get -u leopotam.com/go/ecs/pkg/ecsdi@830f682
```
После скачивания пакет будет доступен как `"leopotam.com/go/ecs/pkg/ecsdi"`.

# Интеграция
```go
systems := ecs.NewSystems(ecs.NewWorld())
systems.
    Add (&System1{})
    .AddWorld(ecs.NewWorld{}, "events")
// Вызов Inject() должен быть размещен после регистрации
// всех систем и миров, но до вызова Init().
ecsdi.Inject(systems).Init()
```

# Специальные типы

> **ВАЖНО!** Инъекция идет только в публичные поля систем.

## World
```go
type TestSystem1 struct {
    // Поле будет содержать ссылку на мир "по умолчанию".
    World       ecsdi.World
    // Поле будет содержать ссылку на мир "events".
    EventsWorld ecsdi.World `ecsdi:"events"`
}
```

## Pool
```go
type TestSystem1 struct {
    // Поле будет содержать ссылку на пул из мира "по умолчанию".
    C1Pool       ecsdi.Pool[C1]
    // Поле будет содержать ссылку на пул из мира "events".
    EventsC1Pool ecsdi.Pool[C1] `ecsdi:"events"`
}
```

## Filter
```go
type TestSystem1 struct {
    // Поле будет содержать ссылку на выборку (с C1) из мира "по умолчанию".
    Filter1       ecsdi.Filter[ecs.Inc1[C1]]
    // Поле будет содержать ссылку на выборку (с C1 и C2) из мира "по умолчанию".
    Filter2       ecsdi.Filter[ecs.Inc2[C1, C2]]
    // Поле будет содержать ссылку на выборку (с C1, но без C2) из мира "по умолчанию".
    Filter3       ecsdi.FilterWithExc[ecs.Inc1[C1], ecs.Exc[C2]]
    // Поле будет содержать ссылку на выборку (с C1, но без C2) из мира "events".
    Filter4       ecsdi.FilterWithExc[ecs.Inc1[C1], ecs.Exc[C2]] `ecsdi:"events"`
}
```
Пулы компонентов, использующиеся в качестве `Include`-ограничений, доступны через поле `Pools`:
```go
Filter2 ecsdi.Filter[ecs.Inc2[C1, C2]]
//...
for it := Filter2.Value.Iter(); it.Next(); {
    entity := it.GetEntity()
    c1 := Filter2.Pools.Inc1.Get(entity)
    c2 := Filter2.Pools.Inc2.Get(entity)
}
```

## Custom
```go
custom1 := Custom1{ID : 1}
custom2 := Custom2{Name : "test"}
systems.Add(&TestSystem1{})
ecsdi.Inject(systems, &custom1, &custom2).Init()
type TestSystem1 struct {
    // Поле будет содержать ссылку на объект совместимого типа, переданого в вызов ecsdi.Inject(xxx).
    Data1 ecsdi.Custom[Custom1]
    // Поле будет содержать ссылку на объект совместимого типа, переданого в вызов ecsdi.Inject(xxx).
    Data2 ecsdi.Custom[Custom2]
}
```

# Лицензия
Фреймворк выпускается под двумя лицензиями, [подробности тут](./../../LICENSE.md).

В случаях лицензирования по условиям MIT-Red не стоит расчитывать на
персональные консультации или какие-либо гарантии.