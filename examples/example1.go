package main

import (
	"fmt"
	. "github.com/Vlvin/gosecs"
	"os"
	"os/signal"
	"syscall"
)

type StateT struct {
	greet string
}

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

type SayHiSystem struct {
	entities map[Entity]bool
}

// Init implements secs.System.
func (s *SayHiSystem) Init(e *ECS[*StateT]) bool {
	s.entities = make(map[Entity]bool)
	e.AssignOnComponentAdded(s, s.OnComponentAdded)
	e.AssignOnComponentRemoved(s, s.OnComponentRemoved)
	e.AssignOnEntityCreated(s, s.OnEntityCreated)
	e.AssignOnEntityRemoved(s, s.OnEntityRemoved)
	for _, entity := range e.EntitiesWithComponents(s.RequiredComponents()...) {
		s.entities[entity] = true
	}
	return true
}

// RequiredComponents implements secs.System.
func (s SayHiSystem) RequiredComponents() []ComponentName {
	return []ComponentName{
		NameComponent{}.GetName(),
	}
}

// Run implements secs.System.
func (s SayHiSystem) Run(e *ECS[*StateT], args *StateT) bool {
	component_name := NameComponent{}.GetName()
	for entity, alive := range s.entities {
		if !alive {
			continue
		}
		fmt.Printf("%s %s\n", args.greet, e.GetComponent(entity, component_name).(NameComponent).name)
	}
	return true
}

// onComponentAdded implements secs.System.
func (s SayHiSystem) OnComponentAdded(e *ECS[*StateT], entity Entity, component Component) {
	if s.entities[entity] {
		return
	}
	if e.HasComponents(entity, s.RequiredComponents()...) {
		s.entities[entity] = true
	}
}

// onComponentRemoved implements secs.System.
func (s SayHiSystem) OnComponentRemoved(e *ECS[*StateT], entity Entity, component Component) {
	if !s.entities[entity] {
		return
	}
	for _, componentName := range s.RequiredComponents() {
		if componentName == component.GetName() {
			delete(s.entities, entity)
		}
	}
}

// onEntityCreated implements secs.System.
func (s *SayHiSystem) OnEntityCreated(e *ECS[*StateT], entity Entity, components ...Component) {
	if e.HasComponents(entity, s.RequiredComponents()...) {
		s.entities[entity] = true
	}
}

// onEntityRemoved implements secs.System.
func (s *SayHiSystem) OnEntityRemoved(e *ECS[*StateT], entity Entity) {
	delete(s.entities, entity)
}

type IntroSystem struct {
	entities map[Entity]bool
}

// Init implements secs.System.
func (s *IntroSystem) Init(e *ECS[*StateT]) bool {
	s.entities = make(map[Entity]bool)
	e.AssignOnComponentAdded(s, s.OnComponentAdded)
	e.AssignOnComponentRemoved(s, s.OnComponentRemoved)
	e.AssignOnEntityCreated(s, s.OnEntityCreated)
	e.AssignOnEntityRemoved(s, s.OnEntityRemoved)
	for _, entity := range e.EntitiesWithComponents(s.RequiredComponents()...) {
		s.entities[entity] = true
	}
	return true
}

// RequiredComponents implements secs.System.
func (s IntroSystem) RequiredComponents() []ComponentName {
	return []ComponentName{
		NameComponent{}.GetName(),
		AgeComponent{}.GetName(),
	}
}

// Run implements secs.System.
func (s IntroSystem) Run(e *ECS[*StateT], args *StateT) bool {
	nameComponentName, ageComponentName := NameComponent{}.GetName(), AgeComponent{}.GetName()
	for _, entity := range e.EntitiesWithComponents(nameComponentName, ageComponentName) {
		nameComponent := e.Components[nameComponentName][entity].(NameComponent)
		ageComponent := e.Components[ageComponentName][entity].(AgeComponent)
		fmt.Printf("Let me introduce my friend, %s, he is %d years old\n", nameComponent.name, ageComponent.age)
	}
	return true
}

// onComponentAdded implements secs.System.
func (s IntroSystem) OnComponentAdded(e *ECS[*StateT], entity Entity, component Component) {
	if s.entities[entity] {
		return
	}
	if e.HasComponents(entity, s.RequiredComponents()...) {
		s.entities[entity] = true
	}
}

// onComponentRemoved implements secs.System.
func (s IntroSystem) OnComponentRemoved(e *ECS[*StateT], entity Entity, component Component) {
	if !s.entities[entity] {
		return
	}
	for _, componentName := range s.RequiredComponents() {
		if componentName == component.GetName() {
			delete(s.entities, entity)
		}
	}
}

// onEntityCreated implements secs.System.
func (s *IntroSystem) OnEntityCreated(e *ECS[*StateT], entity Entity, components ...Component) {
	if e.HasComponents(entity, s.RequiredComponents()...) {
		s.entities[entity] = true
	}
}

// onEntityRemoved implements secs.System.
func (s *IntroSystem) OnEntityRemoved(e *ECS[*StateT], entity Entity) {
	delete(s.entities, entity)
}

type SystemOfAShutDown[SysStateT any] struct {
	shutDown bool
}

func (s *SystemOfAShutDown[SysStateT]) Init(e *ECS[SysStateT]) bool {
	s.shutDown = false
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		signal := <-c
		fmt.Println()
		fmt.Println(signal)
		s.shutDown = true
	}()
	return true
}
func (s SystemOfAShutDown[SysStateT]) Run(e *ECS[SysStateT], args SysStateT) bool {
	return !s.shutDown
}
func (s SystemOfAShutDown[SysStateT]) RequiredComponents() []ComponentName {
	return nil
}
func SystemOfAShutDownCheckCreate[SysStateT any]() func(*ECS[SysStateT], SysStateT) bool {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	terminate := false
	go func() {
		signal := <-c
		fmt.Println()
		fmt.Println(signal)
		terminate = true
	}()
	return func(ecs *ECS[SysStateT], args SysStateT) bool {
		return !terminate
	}
}

func main() {
	state := StateT{"Aloha"}
	ecs := NewECS[*StateT]()
	ecs.
		OnStart(&IntroSystem{}).
		OnUpdate(&SystemOfAShutDown[*StateT]{}, &SayHiSystem{})
	for _, person := range []struct {
		name string
		age  int
	}{{"John", 32}, {"Michael", 47}, {"Roxxie", 45}, {"Diane", 39}} {
		ecs.NewEntity(NameComponent{person.name}, AgeComponent{person.age})
	}
	for _, name := range []string{"Guy without age"} {
		ecs.NewEntity(NameComponent{name})
	}
	ecs.Run(&state)
}
