package main

import "fmt"

type ECS struct {
	Components       map[ComponentName]map[Entity]Component
	Systems          []System
	EntitiesCount    int
	EntitiesRegistry map[Entity][]ComponentName
}

func NewECS() *ECS {
	return &ECS{
		Components:    make(map[ComponentName]map[Entity]Component),
		EntitiesCount: 0,
	}
}

func (e *ECS) RegisterComponent(component Component) bool {
	_, ok := e.Components[component.GetName()]
	if !ok {
		e.Components[component.GetName()] = make(map[Entity]Component)
	}
	return !ok
}

func (e *ECS) RegisterSystem(system System) {
	e.Systems = append(e.Systems, system)
}

func (e *ECS) NewEntity(components ...Component) Entity {
	id := Entity(e.EntitiesCount)
	e.EntitiesCount += 1
	for i := range components {
		e.RegisterComponent(components[i])
		e.Components[components[i].GetName()][id] = components[i]
	}
	return id
}

func (e *ECS) EntitieWithComponent(componentName ComponentName) []Entity {
	entities := []Entity{}
	for entity := range e.Components[componentName] {
		entities = append(entities, entity)
	}
	return entities
}

func (e *ECS) EntitiesWithComponents(componentNames ...ComponentName) []Entity {
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

func (e *ECS) Update() {
	for _, system := range e.Systems {
		system(e)
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

type System func(*ECS)

func SayHiSystem(ecs *ECS) {
	component_name := NameComponent{}.GetName()
	for _, nameComponent := range ecs.Components[component_name] {
		fmt.Printf("Hi %s\n", nameComponent.(NameComponent).name)
	}
}

func IntroSystem(ecs *ECS) {
	nameComponentName, ageComponentName := NameComponent{}.GetName(), AgeComponent{}.GetName()
	for _, entity := range ecs.EntitiesWithComponents(nameComponentName, ageComponentName) {
		nameComponent := ecs.Components[nameComponentName][entity].(NameComponent)
		ageComponent := ecs.Components[ageComponentName][entity].(AgeComponent)
		fmt.Printf("Let me introduce my friend, %s, he is %d years old\n", nameComponent.name, ageComponent.age)
	}
}

func main() {
	ecs := NewECS()
	ecs.RegisterSystem(IntroSystem)
	ecs.RegisterSystem(SayHiSystem)
	for _, person := range []struct {
		name string
		age  int
	}{{"John", 32}, {"Michael", 47}, {"Roxxie", 45}, {"Diane", 39}} {
		ecs.NewEntity(NameComponent{person.name}, AgeComponent{person.age})
	}
	for _, name := range []string{"Guy without age"} {
		ecs.NewEntity(NameComponent{name})
	}
	ecs.Update()
}
