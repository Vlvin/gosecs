package secs

import "fmt"

type SystemRunTime int

const (
	SystemOnStart SystemRunTime = iota
	SystemOnExit
	SystemOnUpdate
)

type SystemID = int
type ECS[SysStateT any] struct {
	Components     map[ComponentName]map[Entity]Component
	SystemsIndices map[SystemRunTime][]int
	Systems        []System[SysStateT]
	EntitiesCount  int
	// Events
	onComponentAdded   map[SystemID]func(*ECS[SysStateT], Entity, Component)
	onComponentRemoved map[SystemID]func(*ECS[SysStateT], Entity, Component)
}

func NewECS[SysStateT any]() *ECS[SysStateT] {
	result := &ECS[SysStateT]{
		Components:     make(map[ComponentName]map[Entity]Component),
		SystemsIndices: make(map[SystemRunTime][]int),
		Systems:        []System[SysStateT]{},
		EntitiesCount:  0,
	}
	result.SystemsIndices[SystemOnStart] = []int{}
	result.SystemsIndices[SystemOnUpdate] = []int{}
	result.SystemsIndices[SystemOnExit] = []int{}
	return result
}

func (e *ECS[SysStateT]) AssignOnComponentAdded(id SystemID, f func(*ECS[SysStateT], Entity, Component)) {
	e.onComponentAdded[id] = f
}

func (e *ECS[SysStateT]) OnComponentAdded(entity Entity, component Component) {
	for _, f := range e.onComponentAdded {
		if f != nil {
			f(e, entity, component)
		}
	}
}

func (e *ECS[SysStateT]) AssignOnComponentRemoved(id SystemID, f func(*ECS[SysStateT], Entity, Component)) {
	e.onComponentRemoved[id] = f
}

func (e *ECS[SysStateT]) OnComponentRemoved(entity Entity, component Component) {
	for _, f := range e.onComponentRemoved {
		if f != nil {
			f(e, entity, component)
		}
	}
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

func (e *ECS[SysStateT]) RegisterSystem(time SystemRunTime, system System[SysStateT]) *ECS[SysStateT] {
	e.Systems = append(e.Systems, system)
	e.SystemsIndices[time] = append(e.SystemsIndices[time], len(e.Systems)-1)
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
	for i, system := range e.Systems {
		ok := true
		for _, comp := range system.RequiredComponents() {
			if _, ok = e.Components[comp][id]; !ok {
				break
			}
		}
		if ok {
			e.Systems[i].AddEntity(id)
		}
	}
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
	//
	// s_entities := map[Entity]bool{}
	// for _, entity := range e.EntitiesWithComponent(componentNames[0]) {
	// 	s_entities[entity] = true
	// }
	// for _, componentName := range componentNames {
	// 	s_intersection := map[Entity]bool{}
	// 	for _, entity := range e.EntitiesWithComponent(componentName) {
	// 		s_intersection[entity] = s_entities[entity]
	// 	}
	// 	s_entities = s_intersection
	// }
	// entities := []Entity{}
	// for entity, exists := range s_entities {
	// 	if exists {
	// 		entities = append(entities, entity)
	// 	}
	// }
	// return entities
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
	RequiredComponents() []ComponentName
	/// Adds entity to the systems
	AddEntity(entity Entity)
	/// Removes entity from the system
	RemoveEntity(entity Entity)
}
