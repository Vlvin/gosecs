package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

type SystemRunTime int

const (
	SystemOnStart SystemRunTime = iota
	SystemOnExit
	SystemOnUpdate
)

type ECS[SysArgs any] struct {
	Components       map[ComponentName]map[Entity]Component
	Systems          map[SystemRunTime][]System[SysArgs]
	EntitiesCount    int
	EntitiesRegistry map[Entity][]ComponentName
}

func NewECS[SysArgs any]() *ECS[SysArgs] {
	return &ECS[SysArgs]{
		Components:    make(map[ComponentName]map[Entity]Component),
		Systems:       make(map[SystemRunTime][]System[SysArgs]),
		EntitiesCount: 0,
	}
}

func (e *ECS[SysArgs]) RegisterComponent(component Component) bool {
	_, ok := e.Components[component.GetName()]
	if !ok {
		e.Components[component.GetName()] = make(map[Entity]Component)
	}
	return !ok
}

func (e *ECS[SysArgs]) RegisterSystem(time SystemRunTime, system System[SysArgs]) {
	e.Systems[time] = append(e.Systems[time], system)
}

func (e *ECS[SysArgs]) NewEntity(components ...Component) Entity {
	id := Entity(e.EntitiesCount)
	e.EntitiesCount += 1
	for i := range components {
		e.RegisterComponent(components[i])
		e.Components[components[i].GetName()][id] = components[i]
	}
	return id
}

func (e *ECS[SysArgs]) EntitieWithComponent(componentName ComponentName) []Entity {
	entities := []Entity{}
	for entity := range e.Components[componentName] {
		entities = append(entities, entity)
	}
	return entities
}

func (e *ECS[SysArgs]) EntitiesWithComponents(componentNames ...ComponentName) []Entity {
	if len(componentNames) < 1 {
		return []Entity{}
	}
	s_entities := map[Entity]bool{}
	for _, entity := range e.EntitieWithComponent(componentNames[0]) {
		s_entities[entity] = true
	}
	for _, componentName := range componentNames {
		s_intersection := map[Entity]bool{}
		for _, entity := range e.EntitieWithComponent(componentName) {
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

func (e *ECS[SysArgs]) Start(args SysArgs) {
	for _, system := range e.Systems[SystemOnStart] {
		system(e, args)
	}
}

func (e *ECS[SysArgs]) Update(args SysArgs) bool {
	run := true
	for _, system := range e.Systems[SystemOnUpdate] {
		run = run && system(e, args)
	}
	return run
}

func (e *ECS[SysArgs]) Exit(args SysArgs) {
	for _, system := range e.Systems[SystemOnExit] {
		system(e, args)
	}
}

func (e *ECS[SysArgs]) Run(args SysArgs) {
	e.Start(args)
	defer e.Exit(args)
	for e.Update(args) {

	}
}

type Entity int

type Component interface {
	GetName() ComponentName
}

type ComponentName string

type NameComponent struct {
	name string
}

func (nc NameComponent) GetName() ComponentName {
	return "NameComponent"
}

type AgeComponent struct {
	age int
}

func (nc AgeComponent) GetName() ComponentName {
	return "AgeComponent"
}

type System[SysArgs any] func(*ECS[SysArgs], SysArgs) bool

func SayHiSystem(ecs *ECS[struct{}], args struct{}) bool {
	component_name := NameComponent{}.GetName()
	for _, nameComponent := range ecs.Components[component_name] {
		fmt.Printf("Hi %s\n", nameComponent.(NameComponent).name)
	}
	return true
}

func IntroSystem(ecs *ECS[struct{}], args struct{}) bool {
	nameComponentName, ageComponentName := NameComponent{}.GetName(), AgeComponent{}.GetName()
	for _, entity := range ecs.EntitiesWithComponents(nameComponentName, ageComponentName) {
		nameComponent := ecs.Components[nameComponentName][entity].(NameComponent)
		ageComponent := ecs.Components[ageComponentName][entity].(AgeComponent)
		fmt.Printf("Let me introduce my friend, %s, he is %d years old\n", nameComponent.name, ageComponent.age)
	}
	return true
}

func SystemOfAShutDownCreate() func(*ECS[struct{}], struct{}) bool {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	terminate := false
	go func() {
		signal := <-c
		fmt.Println()
		fmt.Println(signal)
		terminate = true
	}()
	return func(ecs *ECS[struct{}], args struct{}) bool {
		return !terminate
	}

}

func main() {
	ecs := NewECS[struct{}]()
	ecs.RegisterSystem(SystemOnExit, IntroSystem)
	ecs.RegisterSystem(SystemOnExit, SayHiSystem)
	ecs.RegisterSystem(SystemOnUpdate, SystemOfAShutDownCreate())
	for _, person := range []struct {
		name string
		age  int
	}{{"John", 32}, {"Michael", 47}, {"Roxxie", 45}, {"Diane", 39}} {
		ecs.NewEntity(NameComponent{person.name}, AgeComponent{person.age})
	}
	for _, name := range []string{"Guy without age"} {
		ecs.NewEntity(NameComponent{name})
	}
	ecs.Run(struct{}{})
}
