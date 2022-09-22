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
    * [Filter](#Filter)
* [Расширения](#Расширения)
* [Лицензия](#Лицензия)
* [ЧаВо](#ЧаВо)

# Социальные ресурсы
[![discord](https://img.shields.io/discord/404358247621853185.svg?label=enter%20to%20discord%20server&style=for-the-badge&logo=discord)](https://discord.gg/5GZVde6)

# Установка
Поддерживается установка штатным модулем:
```
go get -u leopotam.com/go/ecs
```
По умолчанию используется последняя релизная версия. Если требуется версия "в разработке" с актуальными изменениями - следует скопировать хеш нужного коммита из ветки `develop` и подставить в командную строку. Например:
```
go get -u leopotam.com/go/ecs@830f682
```
После скачивания пакет будет доступен как `"leopotam.com/go/ecs"`.

# Основные понятия

## Сущность
Сама по себе ничего не значит и не существует, является исключительно контейнером для компонентов. Реализована как `int`:
```go
// Создаем новую сущность в мире.
entity := world.NewEntity()

// Любая сущность может быть удалена, при этом сначала все компоненты будут автоматически удалены и только потом сущность будет считаться уничтоженной. 
world.DelEntity(entity)

// Компоненты с любой сущности могут быть скопированы на другую. Если исходная или целевая сущность не существует - будет брошено исключение в DEBUG-версии.
world.CopyEntity (srcEntity, dstEntity)
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
Компоненты могут быть добавлены, запрошены или удалены через [компонентные пулы](#pool).

## Система
Является контейнером для основной логики для обработки отфильтрованных сущностей. Существует в виде пользовательского класса, реализующего как минимум один из `IInitSystem`, `IDestroySystem`, `IRunSystem` (и прочих поддерживаемых) интерфейсов:
```go
type PreInitSystem1 struct {}
type InitSystem1 struct {}
type RunSystem1 struct {}
type DestroySystem1 struct {}
type PostDestroySystem1 struct {}

func (s *PreInitSystem1) PreInit(systems ecs.ISystems) {
    // Будет вызван один раз в момент работы ISystems.Init() и до срабатывания IInitSystem.Init().
}
func (s *InitSystem1) Init(systems ecs.ISystems) {
    // Будет вызван один раз в момент работы ISystems.Init() и после срабатывания IPreInitSystem.PreInit().
}
func (s *RunSystem1) Run(systems ecs.ISystems) {
    // Будет вызван один раз в момент работы ISystems.Run().
}
func (s *DestroySystem1) Destroy(systems ecs.ISystems) {
    // Будет вызван один раз в момент работы ISystems.Destroy() и до срабатывания IPostDestroySystem.PostDestroy().
}
func (s *PostDestroySystem1) PostDestroy(systems ecs.ISystems) {
    // Будет вызван один раз в момент работы ISystems.Destroy() и после срабатывания IDestroySystem.Destroy().
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
var systems ecs.ISystems

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

> **ВАЖНО!** Необходимо вызывать `ISystems.Destroy()` у экземпляра группы систем если он больше не нужен.

## Filter
Представляют собой механизм итерирования по сущностям, выбранным на основе определенных требований к компонентам (наличию или отсутствию):
```go
type C1{}
type C2{}
type C3{}

w := ecs.NewWorld()
// В выборку попадут все сущности с компонентом C1.
f1 := ecs.GetFilter[ecs.Inc1[C1]](w)
// В выборку попадут все сущности с компонентами C1 и C2 одновременно.
f2 := ecs.GetFilter[ecs.Inc2[C1, C2]](w)
// В выборку попадут все сущности с компонентами C1 и C2 и без C3 одновременно.
f3 := ecs.GetFilterWithExc[ecs.Inc2[C1, C2], ecs.Exc1[C3]](w)

// Способ обработки у всех запросов одинаков:
for it := f1.Iter(); it.Next(); {
    entity := it.GetEntity()
    // Дальнейшая работа с сущностью.
}
```
> **ВАЖНО!** Необходимо вызывать `it.Destroy()` у итератора, созданного вне цикла, либо если происходит принудительное прерывание цикла до его конца:
```go
for it := f1.Iter(); it.Next(); {
    it.Destroy()
    return
}
```
# Расширения

* [Инъекция зависимостей](https://github.com/leopotam/goecs/tree/master/pkg/ecsdi)
* [Многопоточная обработка](https://github.com/leopotam/goecs/tree/master/pkg/ecsmt)

# Лицензия
Фреймворк выпускается под двумя лицензиями, [подробности тут](./LICENSE.md).

В случаях лицензирования по условиям MIT-Red не стоит расчитывать на
персональные консультации или какие-либо гарантии.

# ЧаВо

## Меня не устраивают значения по умолчанию для полей компонентов. Как я могу это настроить?

Компоненты поддерживают кастомную настройку значений через реализацию интерфейса `IComponentReset`:
```go
type C1 struct{
    ID int
}

func (c *C1) Reset() {
    c.ID = -1
}
```
Этот метод будет автоматически вызываться для всех новых компонентов, а так же для всех только что удаленных, до помещения их в пул.
> **ВАЖНО!** В случае применения IEcsAutoReset все дополнительные очистки/проверки полей компонента отключаются, что может привести к утечкам памяти. Ответственность лежит на пользователе!

### Меня не устраивают значения для полей компонентов при их копировании через World.CopyEntity() или Pool[].Copy(). Как я могу это настроить?

Компоненты поддерживают установку произвольных значений при вызове `World.CopyEntity()` или `Pool[].Copy()` через реализацию интерфейса `IComponentCopy[]`:
```go
type C1 struct{
    ID int
}

func (c *C1) Copy(src *C1) {
    c.ID = src.ID * 123
}
```
> **ВАЖНО!** В случае применения `IComponentCopy[]` никакого копирования по умолчанию не происходит. Ответственность за корректность заполнения данных и за целостность исходных лежит на пользователе!


## Я хочу сохранить ссылку на сущность в компоненте. Как я могу это сделать?

Для сохранения ссылки на сущность ее необходимо упаковать в один из специальных контейнеров (`PackedEntity` или `PackedEntityWithWorld`):
```go
w := ecs.NewWorld()
e := w.NewEntity()
// PackedEntity - контейнер без ссылки на мир.
packedEntity := w.PackEntity(e)
if unpackedEntity1, ok := packedEntity.Unpack(w); ok {
    // unpackedEntity1 - сущность жива и может быть использована.
}

// PackedEntityWithWorld - контейнер со ссылкой на мир.
packedEntityWithWorld := w.PackEntityWithWorld(e)
if unpackedWorld, unpackedEntity2, ok := packedEntityWithWorld.Unpack(); ok {
    // unpackedEntity2 - сущность жива и может быть использована.
}
```

## Мне нужно больше чем 6-"Include" и 3-"Exclude" ограничений для компонентов в фильтре. Как я могу сделать это?
Для расширения списка `include`-требований необходимо создать новый тип, реализующий `IInc`-интерфейс. Например, нужна поддержка 7 компонентов:
```go
type Inc7[I1 any, I2 any, I3 any, I4 any, I5 any, I6 any, I7 any] struct {
	Inc1 *ecs.Pool[I1]
	Inc2 *ecs.Pool[I2]
	Inc3 *ecs.Pool[I3]
	Inc4 *ecs.Pool[I4]
    Inc5 *ecs.Pool[I5]
	Inc6 *ecs.Pool[I6]
	Inc7 *ecs.Pool[I7]
}

func (i Inc7[I1, I2, I3, I4, I5, I6, I7]) FillIncludes(w *ecs.World, list []int16) []int16 {
	list = append(list, ecs.GetPool[I1](w).GetID())
	list = append(list, ecs.GetPool[I2](w).GetID())
	list = append(list, ecs.GetPool[I3](w).GetID())
    list = append(list, ecs.GetPool[I4](w).GetID())
	list = append(list, ecs.GetPool[I5](w).GetID())
	list = append(list, ecs.GetPool[I6](w).GetID())
	return append(list, ecs.GetPool[I7](w).GetID())
}

func (i Inc7[I1, I2, I3, I4, I5, I6, I7]) FillPools(w *ecs.World) ecs.IInc {
	return &Inc7[I1, I2, I3, I4, I5, I6, I7]{
		Inc1: ecs.GetPool[I1](w),
		Inc2: ecs.GetPool[I2](w),
		Inc3: ecs.GetPool[I3](w),
		Inc4: ecs.GetPool[I4](w),
        Inc5: ecs.GetPool[I5](w),
		Inc6: ecs.GetPool[I6](w),
		Inc7: ecs.GetPool[I7](w),
	}
}
```

Для расширения списка `exclude`-требований необходимо создать новый тип, реализующий `IExc`-интерфейс. Например, нужна поддержка 4 компонентов:
```go
type Exc4[E1 any, E2 any, E3 any, E4 any] struct {
	Exc1 *ecs.Pool[E1]
	Exc2 *ecs.Pool[E2]
	Exc3 *ecs.Pool[E3]
    Exc4 *ecs.Pool[E4]
}

func (e Exc4[E1, E2, E3, E4]) FillExcludes(w *ecs.World, list []int16) []int16 {
	list = append(list, ecs.GetPool[E1](w).GetID())
	list = append(list, ecs.GetPool[E2](w).GetID())
	list = append(list, ecs.GetPool[E3](w).GetID())
	return append(list, ecs.GetPool[E4](w).GetID())
}
```