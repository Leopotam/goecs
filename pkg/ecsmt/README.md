# GoECS MT - Поддержка многопоточной обработки
Обеспечивает поддержку обработки сущностей в несколько системных потоков для GoECS.

# Содержание
* [Социальные ресурсы](#Социальные-ресурсы)
* [Установка](#Установка)
* [Интеграция](#Интеграция)
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

# Интеграция
```go
// Стартовый код.
systems := ecs.NewSystems(ecs.NewWorld())
systems.
    Add (&System1{}).
    Init()

// Компонент.
type c1 struct {
    id int
}

// Система.
type System1 struct {
    filter *ecs.Filter
    pool *ecs.Pool[c1]
}
func (s *System1) Init(systems *ecs.Systems) {
    w := systems.GetWorld()
    s.filter = GetFilter[ecs.Inc1[C1]](w)
    s.pool = GetPool[C1](w)
}
func (s *System1) Run(systems *ecs.Systems) {
    // Размер блока данных, после которого наступает разделение.
    chunkSize := 1000
    ecsmt.Run(s, s.filter, chunkSize)
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

> **ВАЖНО!** Внутри обработчика **запрещено** изменять состояние мира: нельзя создавать / удалять сущности, нельзя добавлять / удалять компоненты на сущности. Допускается только модификация данных внутри существующих компонентов.

# Лицензия
Фреймворк выпускается под двумя лицензиями, [подробности тут](./../../LICENSE.md).

В случаях лицензирования по условиям MIT-Red не стоит расчитывать на
персональные консультации или какие-либо гарантии.