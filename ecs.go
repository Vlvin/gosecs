package ecs

import "fmt"

type SystemRunTime int

const (
	SystemOnStart SystemRunTime = iota
	SystemOnExit
	SystemOnUpdate
)

type ECS[SysStateT any] struct {
	Components       map[ComponentName]map[Entity]Component
	Systems          map[SystemRunTime][]System[SysStateT]
	EntitiesCount    int
	EntitiesRegistry map[Entity][]ComponentName
}

func NewECS[SysStateT any]() *ECS[SysStateT] {
	return &ECS[SysStateT]{
		Components:    make(map[ComponentName]map[Entity]Component),
		Systems:       make(map[SystemRunTime][]System[SysStateT]),
		EntitiesCount: 0,
	}
}

func (e *ECS[SysStateT]) RegisterComponent(component Component) bool {
	_, ok := e.Components[component.GetName()]
	if !ok {
		e.Components[component.GetName()] = make(map[Entity]Component)
	}
	return !ok
}

func (e *ECS[SysStateT]) RegisterSystem(time SystemRunTime, system System[SysStateT]) *ECS[SysStateT] {
	e.Systems[time] = append(e.Systems[time], system)
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
	id := Entity(e.EntitiesCount)
	e.EntitiesCount += 1
	for i := range components {
		e.RegisterComponent(components[i])
		e.Components[components[i].GetName()][id] = components[i]
	}
	return id
}

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
	s_entities := map[Entity]bool{}
	for _, entity := range e.EntitiesWithComponent(componentNames[0]) {
		s_entities[entity] = true
	}
	for _, componentName := range componentNames {
		s_intersection := map[Entity]bool{}
		for _, entity := range e.EntitiesWithComponent(componentName) {
			s_intersection[entity] = s_entities[entity]
		}
		s_entities = s_intersection
	}
	entities := []Entity{}
	for entity, exists := range s_entities {
		if exists {
			entities = append(entities, entity)
		}
	}
	return entities
}

func (e *ECS[SysStateT]) Start(args SysStateT) bool {
	ok := true
	for _, system := range e.Systems[SystemOnStart] {
		ok = ok && system(e, args)
	}
	return ok
}

func (e *ECS[SysStateT]) Update(args SysStateT) bool {
	run := true
	for _, system := range e.Systems[SystemOnUpdate] {
		run = run && system(e, args)
	}
	return run
}

func (e *ECS[SysStateT]) Exit(args SysStateT) bool {
	ok := true
	for _, system := range e.Systems[SystemOnExit] {
		ok = ok && system(e, args)
	}
	return ok
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

type Entity int

type Component interface {
	GetName() ComponentName
}

type ComponentName string

type System[SysStateT any] func(*ECS[SysStateT], SysStateT) bool
