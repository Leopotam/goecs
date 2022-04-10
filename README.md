# GoEcs - Легковесный Golang Entity Component System фреймворк
Относительно неплохая производительность, нулевые или минимальные аллокации, минимизация использования памяти, отсутствие зависимостей - это основные цели данного фреймворка.

> **ВАЖНО!** Фреймворк не готов к использованию на реальных проектах, API будет меняться!

> **ВАЖНО!** Не забывайте использовать `DEBUG`-версии билдов для разработки и `RELEASE`-версии билдов для релизов: все внутренние проверки/исключения будут работать только в `DEBUG`-версиях и удалены для увеличения производительности в `RELEASE`-версиях.
`DEBUG`-версии компилируются по умолчанию, для `RELEASE`-версии билд необходимо собирать
с тегом `RELEASE`:
```
go build -ldflags "-w -s" -tags "RELEASE" .
```

> **ВАЖНО!** GoEcs-фрейморк **не потокобезопасен** и никогда не будет таким! Если вам нужна поддержка goroutines - вы должны реализовать ее самостоятельно и интегрировать синхронизацию в виде ecs-системы.

# Содержание
* [Социальные ресурсы](#Социальные-ресурсы)
* [Установка](#Установка)
* [Основные понятия](#Основные-понятия)
    * [Сущность](#Сущность)
    * [Компонент](#Компонент)
    * [Система](#Система)
* [Специальные типы](#Специальные-типы)
    * [World](#World)
    * [Pool](#Pool)
    * [Systems](#Systems)
    * [Query](#Query)
* [Лицензия](#Лицензия)

# Социальные ресурсы
[![discord](https://img.shields.io/discord/404358247621853185.svg?label=enter%20to%20discord%20server&style=for-the-badge&logo=discord)](https://discord.gg/5GZVde6)

# Установка
Поддерживается установка штатных модулем:
```
go get -u github.com/leopotam/goecs
```
По умолчанию используется последняя релизная версия. Если требуется версия "в разработке" с актуальными изменениями - следует переключиться на ветку `develop`:
```
go get -u github.com/leopotam/goecs@develop
```

# Основные понятия

## Сущность
Сама по себе ничего не значит и не существует, является исключительно контейнером для компонентов. Реализована как `int`:
```go
// Создаем новую сущность в мире.
entity := world.NewEntity()

// Любая сущность может быть удалена, при этом сначала все компоненты будут автоматически удалены и только потом энтити будет считаться уничтоженной. 
world.DelEntity(entity)
```

> **ВАЖНО!** Сущности не могут существовать без компонентов и будут автоматически уничтожаться при удалении последнего компонента на них.

## Компонент
Является контейнером для данных пользователя и не должен содержать логику (допускаются минимальные хелперы, но не куски основной логики):
```go
type Component1 struct {
    ID int
    Name string
}
```
Компоненты могут быть добавлены, запрошены или удалены через [компонентные пулы](#ecspool).

## Система
Является контейнером для основной логики для обработки отфильтрованных сущностей. Существует в виде пользовательского класса, реализующего как минимум один из `IInitSystem`, `IDestroySystem`, `IRunSystem` (и прочих поддерживаемых) интерфейсов:
```go
type PreInitSystem1 struct {}
type InitSystem1 struct {}
type RunSystem1 struct {}
type DestroySystem1 struct {}
type PostDestroySystem1 struct {}

func (s *PreInitSystem1) PreInit(systems *ecs.Systems) {
	// Будет вызван один раз в момент работы Systems.Init() и до срабатывания IInitSystem.Init().
}
func (s *InitSystem1) Init(systems *ecs.Systems) {
	// Будет вызван один раз в момент работы Systems.Init() и после срабатывания IPreInitSystem.PreInit().
}
func (s *RunSystem1) Run(systems *ecs.Systems) {
	// Будет вызван один раз в момент работы Systems.Run().
}
func (s *DestroySystem1) Destroy(systems *ecs.Systems) {
	// Будет вызван один раз в момент работы Systems.Destroy() и до срабатывания IPostDestroySystem.PostDestroy().
}
func (s *PostDestroySystem1) PostDestroy(systems *ecs.Systems) {
	// Будет вызван один раз в момент работы Systems.Destroy() и после срабатывания IDestroySystem.Destroy().
}
```

# Специальные типы

## World
Является контейнером для всех сущностей, компонентых пулов и фильтров, данные каждого экземпляра уникальны и изолированы от других миров:
```go
w := ecs.NewWorld()
// Работа с миром.
w.Destroy()
```

> **ВАЖНО!** Необходимо вызывать `World.Destroy()` у экземпляра мира если он больше не нужен.

## Pool
Является контейнером для компонентов, предоставляет апи для добавления / запроса / удаления компонентов на сущности:
```go
entity := world.NewEntity()
pool := ecs.GetPool[Component1](world)

// Add() добавляет компонент к сущности. Если компонент уже существует - будет брошено исключение в DEBUG-версии.
c1 := pool.Add(entity)

// Get() возвращает существующий на сущности компонент. Если компонент не существует - будет брошено исключение в DEBUG-версии.
c1 = pool.Get(entity)

// Has() проверяет наличие компонента на сущности.
if pool.Has(entity) {
    // Компонент присутствует
}

// Del() удаляет компонент с сущности. Если компонента не было - никаких ошибок не будет. Если это был последний компонент - сущность будет удалена автоматически.
pool.Del(entity)
```

> **ВАЖНО!** После удаления, компонент будет помещен в пул для последующего переиспользования. Все поля компонента будут сброшены в значения по умолчанию автоматически.

## Systems
Является контейнером для систем, которыми будет обрабатываться `World`-экземпляр мира:
```go

var world *ecs.World
var systems *ecs.Systems

func main() {
    // Создаем окружение, подключаем системы.
    world = ecs.NewWorld()
    systems = ecs.NewSystems(world)
    systems.
        Add(&System1{}).
        Add(&System2{}).
        Init()
    
    // Выполняем все подключенные системы, этот метод надо вызывать
    // в каждом цикле обновления
    systems.Run()

    // Уничтожаем подключенные системы.
    if systems != nil {
        systems.Destroy()
        systems = nil
    }
    // Очищаем окружение.
    if world != nil {
        world.Destroy()
        world = nil
    }
}
```

> **ВАЖНО!** Необходимо вызывать `Systems.Destroy()` у экземпляра группы систем если он больше не нужен.

## Query
Представляют собой механизм итерирования по сущностям, выбранным на основе определенных требований к компонентам (наличию или отсутствию):
```go
type C1{}
type C2{}
type C3{}

w := ecs.NewWorld()
// В выборку попадут все сущности с компонентом C1.
q1 := ecs.NewQuery[ecs.Inc1[C1]](w)
// В выборку попадут все сущности с компонентами C1 и C2 одновременно.
q2 := ecs.NewQuery[ecs.Inc2[C1, C2]](w)
// В выборку попадут все сущности с компонентами C1 и C2 и без C3 одновременно.
q3 := ecs.NewQueryWithExc[ecs.Inc2[C1, C2], ecs.Exc1[C3]](w)

// Способ обработки у всех запросов одинаков:
for it := q1.Iter(); it.Next(); {
    entity := it.GetEntity()
    // Дальнейшая работа с сущностью.
}
```

# Лицензия
Фреймворк выпускается под двумя лицензиями, [подробности тут](./LICENSE.md).

В случаях лицензирования по условиям MIT-Red не стоит расчитывать на
персональные консультации или какие-либо гарантии.