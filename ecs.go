package secs

import (
	"fmt"
	"sync"
)

type SystemRunTime int

const (
	SystemOnStart SystemRunTime = iota
	SystemOnExit
	SystemOnUpdate
)

type SystemID = int
type ECS[SysStateT any] struct {
	Components     map[ComponentName]map[Entity]Component
	SystemsIndices map[SystemRunTime][]SystemID
	Systems        []System[SysStateT]
	EntitiesCount  int
	// Events
	onEntityCreated    map[SystemID]func(*ECS[SysStateT], Entity, ...Component)
	onEntityRemoved    map[SystemID]func(*ECS[SysStateT], Entity)
	onComponentAdded   map[SystemID]func(*ECS[SysStateT], Entity, Component)
	onComponentRemoved map[SystemID]func(*ECS[SysStateT], Entity, Component)
}

func NewECS[SysStateT any]() *ECS[SysStateT] {
	result := &ECS[SysStateT]{
		Components:         make(map[ComponentName]map[Entity]Component),
		SystemsIndices:     make(map[SystemRunTime][]int),
		Systems:            []System[SysStateT]{},
		EntitiesCount:      0,
		onEntityCreated:    make(map[SystemID]func(*ECS[SysStateT], Entity, ...Component)),
		onEntityRemoved:    make(map[SystemID]func(*ECS[SysStateT], Entity)),
		onComponentAdded:   make(map[SystemID]func(*ECS[SysStateT], Entity, Component)),
		onComponentRemoved: make(map[SystemID]func(*ECS[SysStateT], Entity, Component)),
	}
	result.SystemsIndices[SystemOnStart] = []int{}
	result.SystemsIndices[SystemOnUpdate] = []int{}
	result.SystemsIndices[SystemOnExit] = []int{}
	return result
}

func (e *ECS[SysStateT]) OnEntityCreated(entity Entity, components ...Component) {
	var wg sync.WaitGroup
	for _, handler := range e.onEntityCreated {
		if handler != nil {
			wg.Add(1)
			go func() {
				defer wg.Done()
				handler(e, entity, components...)
			}()
		}
	}
	wg.Wait()

}
func (e *ECS[SysStateT]) OnEntityRemoved(entity Entity) {
	var wg sync.WaitGroup
	for _, handler := range e.onEntityRemoved {
		if handler != nil {
			wg.Add(1)
			go func() {
				handler(e, entity)
			}()
		}
	}
	wg.Wait()

}
func (e *ECS[SysStateT]) OnComponentAdded(entity Entity, component Component) {
	var wg sync.WaitGroup
	for _, handler := range e.onComponentAdded {
		if handler != nil {
			wg.Add(1)
			go func() {
				defer wg.Done()
				handler(e, entity, component)
			}()
		}
	}
	wg.Wait()

}
func (e *ECS[SysStateT]) OnComponentRemoved(entity Entity, component Component) {
	var wg sync.WaitGroup
	for _, handler := range e.onComponentRemoved {
		if handler != nil {
			wg.Add(1)
			go func() {
				defer wg.Done()
				handler(e, entity, component)
			}()
		}
	}
	wg.Wait()
}

func (e *ECS[SysStateT]) AssignOnEntityCreated(system System[SysStateT], handler func(*ECS[SysStateT], Entity, ...Component)) {
	e.onEntityCreated[SystemID(len(e.Systems))] = handler
}
func (e *ECS[SysStateT]) AssignOnEntityRemoved(system System[SysStateT], handler func(*ECS[SysStateT], Entity)) {
	e.onEntityRemoved[SystemID(len(e.Systems))] = handler
}
func (e *ECS[SysStateT]) AssignOnComponentAdded(system System[SysStateT], handler func(*ECS[SysStateT], Entity, Component)) {
	e.onComponentAdded[SystemID(len(e.Systems))] = handler
}
func (e *ECS[SysStateT]) AssignOnComponentRemoved(system System[SysStateT], handler func(*ECS[SysStateT], Entity, Component)) {
	e.onComponentRemoved[SystemID(len(e.Systems))] = handler
}

func (e *ECS[SysStateT]) RegisterComponent(component Component) bool {
	_, ok := e.Components[component.GetName()]
	if !ok {
		e.Components[component.GetName()] = make(map[Entity]Component)
	}
	return !ok
}

func (e *ECS[SysStateT]) GetComponent(entity Entity, name ComponentName) Component {
	return e.Components[name][entity]
}

func (e *ECS[SysStateT]) UpdateComponent(entity Entity, component Component) {
	if _, ok := e.Components[component.GetName()][entity]; !ok {
		e.AddComponent(entity, component)
	}
	e.Components[component.GetName()][entity] = component
}

func (e *ECS[SysStateT]) AddComponent(entity Entity, component Component) {
	e.OnComponentAdded(entity, component)
	e.Components[component.GetName()][entity] = component
}

func (e *ECS[SysStateT]) HasComponent(entity Entity, componentName ComponentName) bool {

	_, ok := e.Components[componentName][entity]
	return ok
}

func (e *ECS[SysStateT]) HasComponents(entity Entity, componentNames ...ComponentName) bool {

	for _, componentName := range componentNames {
		if !e.HasComponent(entity, componentName) {
			return false
		}
	}
	return true
}

func (e *ECS[SysStateT]) RegisterSystem(time SystemRunTime, system System[SysStateT]) *ECS[SysStateT] {
	e.Systems = append(e.Systems, system)
	var systemID = SystemID(len(e.Systems) - 1)
	e.SystemsIndices[time] = append(e.SystemsIndices[time], systemID)
	// e.onComponentAdded[systemID] = true
	// e.onComponentRemoved[systemID] = true
	// e.onEntityCreated[systemID] = true
	// e.onEntityRemoved[systemID] = true
	e.Systems[systemID].Init(e)
	return e
}

func (e *ECS[SysStateT]) OnStart(systems ...System[SysStateT]) *ECS[SysStateT] {
	for _, system := range systems {
		e.RegisterSystem(SystemOnStart, system)
	}
	return e
}

func (e *ECS[SysStateT]) OnUpdate(systems ...System[SysStateT]) *ECS[SysStateT] {
	for _, system := range systems {
		e.RegisterSystem(SystemOnUpdate, system)
	}
	return e
}

func (e *ECS[SysStateT]) OnExit(systems ...System[SysStateT]) *ECS[SysStateT] {
	for _, system := range systems {
		e.RegisterSystem(SystemOnExit, system)
	}
	return e
}

func (e *ECS[SysStateT]) NewEntity(components ...Component) Entity {
	e.EntitiesCount += 1
	id := e.EntitiesCount
	for i := range components {
		e.RegisterComponent(components[i])
		e.Components[components[i].GetName()][id] = components[i]
	}
	e.OnEntityCreated(id, components...)

	return id
}

// / NOTE: not a good way of searching entity with multiple components
func (e *ECS[SysStateT]) EntitiesWithComponent(componentName ComponentName) []Entity {
	entities := []Entity{}
	for entity := range e.Components[componentName] {
		entities = append(entities, entity)
	}
	return entities
}

func (e *ECS[SysStateT]) EntitiesWithComponents(componentNames ...ComponentName) []Entity {

	if len(componentNames) < 1 {
		return []Entity{}
	}

	if len(componentNames) == 1 {
		return e.EntitiesWithComponent(componentNames[0])
	}

	nameOfShortest := componentNames[0]
	for _, componentName := range componentNames {
		if len(e.Components[componentName]) < len(e.Components[nameOfShortest]) {
			nameOfShortest = componentName
		}
	}
	// iterating over shortest, and checking for presence others
	entities := []Entity{}
	for entity := range e.Components[nameOfShortest] {
		fits := true
		for _, componentName := range componentNames {
			if _, ok := e.Components[componentName][entity]; !ok {
				fits = false
				break
			}
		}
		if fits {
			entities = append(entities, entity)
		}
	}
	return entities
}

func (e *ECS[SysStateT]) RunSystems(sysTime SystemRunTime, args SysStateT) bool {
	if _, value := e.SystemsIndices[sysTime]; !value {
		fmt.Printf("No systems to run on %d\n", sysTime)
		return false
	}
	ok := true
	for _, systemID := range e.SystemsIndices[sysTime] {
		ok = ok && e.Systems[systemID].Run(e, args)
	}
	return ok
}

func (e *ECS[SysStateT]) Start(args SysStateT) bool {
	return e.RunSystems(SystemOnStart, args)
}

func (e *ECS[SysStateT]) Update(args SysStateT) bool {
	return e.RunSystems(SystemOnUpdate, args)
}

func (e *ECS[SysStateT]) Exit(args SysStateT) bool {
	return e.RunSystems(SystemOnExit, args)
}

func (e *ECS[SysStateT]) Run(args SysStateT) error {
	// for i := range e.Systems {
	// e.Systems[i].Init(e)

	// }
	if !e.Start(args) {
		return fmt.Errorf("Something went wrong on ECS.Start()")
	}
	for e.Update(args) {
	}
	if !e.Exit(args) {
		return fmt.Errorf("Something went wrong on ECS.Exit()")
	}
	return nil

}

type Entity = int

type Component interface {
	GetName() ComponentName
}

type ComponentName = string

// / idea of system being not just a function, but an object that holds references to entities
type System[SysStateT any] interface {
	/// This function is a great chance to pick all entities once
	Init(e *ECS[SysStateT]) bool
	Run(e *ECS[SysStateT], args SysStateT) bool
	/// This is used to filter Entities that does not have some components that Sysytem requires
	// RequiredComponents() []ComponentName
}

type SystemBase[SysStateT any] struct {
	System[SysStateT]
	RequiredComponents []ComponentName
	Entities           map[Entity]bool
}

//
// func (s *SystemBase[SysArgsT]) RequiredComponents() []ComponentName {
// 	return []ComponentName{}
// }

func (s *SystemBase[SysArgsT]) AssignToAllEvents(e *ECS[SysArgsT]) bool {
	// s.entities = make(map[Entity]bool)
	e.AssignOnComponentAdded(s, s.OnComponentAdded)
	e.AssignOnComponentRemoved(s, s.OnComponentRemoved)
	e.AssignOnEntityCreated(s, s.OnEntityCreated)
	e.AssignOnEntityRemoved(s, s.OnEntityRemoved)
	// for _, entity := range e.EntitiesWithComponents(s.RequiredComponents()...) {
	// 	s.entities[entity] = true
	// }
	return true
}

// onComponentAdded implements secs.System.
func (s *SystemBase[SysArgsT]) OnComponentAdded(e *ECS[SysArgsT], entity Entity, component Component) {
	if s.Entities[entity] {
		return
	}
	if e.HasComponents(entity, s.RequiredComponents...) {
		s.Entities[entity] = true
	}
}

// onComponentRemoved implements secs.System.
func (s *SystemBase[SysArgsT]) OnComponentRemoved(e *ECS[SysArgsT], entity Entity, component Component) {
	if !s.Entities[entity] {
		return
	}
	for _, componentName := range s.RequiredComponents {
		if componentName == component.GetName() {
			delete(s.Entities, entity)
		}
	}
}

// onEntityCreated implements secs.System.
func (s *SystemBase[SysArgsT]) OnEntityCreated(e *ECS[SysArgsT], entity Entity, components ...Component) {
	if e.HasComponents(entity, s.RequiredComponents...) {
		s.Entities[entity] = true
	}
}

// onEntityRemoved implements secs.System.
func (s *SystemBase[SysArgsT]) OnEntityRemoved(e *ECS[SysArgsT], entity Entity) {
	delete(s.Entities, entity)
}
