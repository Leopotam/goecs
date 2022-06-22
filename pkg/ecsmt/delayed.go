package ecsmt

import (
	"sync"

	"leopotam.com/go/ecs"
)

type opType int

const (
	newEntity    opType = 0
	delEntity    opType = 1
	addComponent opType = 2
	delComponent opType = 3
)

const defaultBufferCapacity int = 1024
const defaultPoolCapacity int = 512

type delayedOp struct {
	op       opType
	entity   int
	gen      int16
	pool     int
	poolItem int
}

type IDelayedBuffer interface {
	NewEntity() int
	DelEntity(entity int)
	Process()
	addComponent(entity int, poolID, itemID int)
	delComponent(entity, poolID int)
}

type DelayedPool[T any] struct {
	sync   sync.Mutex
	buffer IDelayedBuffer
	world  *ecs.World
	pool   *ecs.Pool[T]
	id     int
	items  []T
}

type iDelayedPool interface {
	link(buffer IDelayedBuffer, id int, world *ecs.World)
	processAdd(entity, itemID int)
	processDel(entity int)
	reset()
}

type delayedBuffer struct {
	sync          sync.Mutex
	world         *ecs.World
	ops           []delayedOp
	pools         []iDelayedPool
	entitiesAdded []int
}

func NewDelayedPool[T any]() *DelayedPool[T] {
	return NewDelayedPoolWithCapacity[T](defaultPoolCapacity)
}

func NewDelayedPoolWithCapacity[T any](capacity int) *DelayedPool[T] {
	return &DelayedPool[T]{items: make([]T, 0, capacity)}
}

func NewDelayedBuffer(world *ecs.World, pools ...iDelayedPool) IDelayedBuffer {
	return NewDelayedBufferWithCapacity(world, defaultBufferCapacity, pools...)
}

func NewDelayedBufferWithCapacity(world *ecs.World, capacity int, pools ...iDelayedPool) IDelayedBuffer {
	b := &delayedBuffer{world: world, ops: make([]delayedOp, 0, capacity), pools: pools}
	for k, v := range pools {
		v.link(b, k, world)
	}
	return b
}

func (b *delayedBuffer) NewEntity() int {
	b.sync.Lock()
	op := delayedOp{
		op:     newEntity,
		entity: -(len(b.entitiesAdded) + 1),
	}
	b.entitiesAdded = append(b.entitiesAdded, -1)
	b.ops = append(b.ops, op)
	b.sync.Unlock()
	return op.entity
}

func (b *delayedBuffer) DelEntity(entity int) {
	if ecs.DEBUG && entity < 0 {
		panic("cant delete delayed entity")
	}
	b.sync.Lock()
	op := delayedOp{
		op:     delEntity,
		entity: entity,
		gen:    b.world.GetEntityGen(entity),
	}
	b.ops = append(b.ops, op)
	b.sync.Unlock()
}

func (b *delayedBuffer) Process() {
	b.sync.Lock()
	defer b.sync.Unlock()
	for _, v := range b.ops {
		switch v.op {
		case newEntity:
			b.entitiesAdded[-(v.entity + 1)] = b.world.NewEntity()
		case delEntity:
			if ecs.DEBUG && b.world.GetEntityGen(v.entity) != v.gen {
				panic("cant delete non-exist entity")
			}
			b.world.DelEntity(v.entity)
		case addComponent:
			if v.entity < 0 {
				v.entity = b.entitiesAdded[-(v.entity + 1)]
			}
			b.pools[v.pool].processAdd(v.entity, v.poolItem)
		case delComponent:
			b.pools[v.pool].processDel(v.entity)
		}
	}
	b.ops = b.ops[:0]
	b.entitiesAdded = b.entitiesAdded[:0]
	for _, v := range b.pools {
		v.reset()
	}
}

func (b *delayedBuffer) addComponent(entity int, poolID, itemID int) {
	op := delayedOp{
		op:       addComponent,
		entity:   entity,
		pool:     poolID,
		poolItem: itemID,
	}
	b.sync.Lock()
	b.ops = append(b.ops, op)
	b.sync.Unlock()
}

func (b *delayedBuffer) delComponent(entity, poolID int) {
	op := delayedOp{
		op:     delComponent,
		entity: entity,
		pool:   poolID,
	}
	b.sync.Lock()
	b.ops = append(b.ops, op)
	b.sync.Unlock()
}

func (p *DelayedPool[T]) link(buffer IDelayedBuffer, id int, world *ecs.World) {
	if ecs.DEBUG && p.buffer != nil {
		panic("already attached to buffer")
	}
	p.buffer = buffer
	p.id = id
	p.world = world
	p.pool = ecs.GetPool[T](world)
}

func (p *DelayedPool[T]) processAdd(entity, itemID int) {
	*p.pool.Add(entity) = p.items[itemID]
}

func (p *DelayedPool[T]) processDel(entity int) {
	p.pool.Del(entity)
}

func (p *DelayedPool[T]) reset() {
	p.items = p.items[:0]
}

func (p *DelayedPool[T]) Add(entity int, v T) {
	if ecs.DEBUG && p.buffer == nil {
		panic("not linked with buffer")
	}
	p.sync.Lock()
	itemID := len(p.items)
	p.items = append(p.items, v)
	p.sync.Unlock()
	p.buffer.addComponent(entity, p.id, itemID)
}

func (p *DelayedPool[T]) Del(entity int) {
	if ecs.DEBUG && p.buffer == nil {
		panic("not linked with buffer")
	}
	if ecs.DEBUG && entity < 0 {
		panic("cant delete delayed component")
	}
	p.buffer.delComponent(entity, p.id)
}

func (p *DelayedPool[T]) Get(entity int) *T {
	if ecs.DEBUG && p.buffer == nil {
		panic("not linked with buffer")
	}
	if ecs.DEBUG && entity < 0 {
		panic("cant get component on delayed entity")
	}
	return p.pool.Get(entity)
}

func (p *DelayedPool[T]) Has(entity int) bool {
	if ecs.DEBUG && p.buffer == nil {
		panic("not linked with buffer")
	}
	if ecs.DEBUG && entity < 0 {
		panic("cant check component on delayed entity")
	}
	return p.pool.Has(entity)
}
