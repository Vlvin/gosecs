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

func SayHiSystem(ecs *ECS[*StateT], args *StateT) bool {
	component_name := NameComponent{}.GetName()
	for _, nameComponent := range ecs.Components[component_name] {
		fmt.Printf("%s %s\n", args.greet, nameComponent.(NameComponent).name)
	}
	return true
}

func IntroSystem(ecs *ECS[*StateT], args *StateT) bool {
	nameComponentName, ageComponentName := NameComponent{}.GetName(), AgeComponent{}.GetName()
	for _, entity := range ecs.EntitiesWithComponents(nameComponentName, ageComponentName) {
		nameComponent := ecs.Components[nameComponentName][entity].(NameComponent)
		ageComponent := ecs.Components[ageComponentName][entity].(AgeComponent)
		fmt.Printf("Let me introduce my friend, %s, he is %d years old\n", nameComponent.name, ageComponent.age)
	}
	return true
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
		OnUpdate(SystemOfAShutDownCheckCreate[*StateT]()).
		OnExit(IntroSystem, SayHiSystem)
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
