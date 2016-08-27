package injector

import (
	"container/list"
	"fmt"
	"reflect"
)

// Container …
type Container struct {
	dependencies []*dependency
}

type dependency struct {
	ConstructorValue reflect.Value
	ParameterTypes   []reflect.Type
	ReturnType       reflect.Type
}

// NewContainer …
func NewContainer() *Container {
	return &Container{}
}

// Register …
func (container *Container) Register(constructors ...interface{}) {
	for _, constructor := range constructors {
		constructorType := reflect.TypeOf(constructor)

		if constructorType.Kind() != reflect.Func {
			panic("Provided constructor is not a function")
		}

		if constructorType.NumOut() != 1 {
			panic("Provided must have single return value")
		}

		constructorValue := reflect.ValueOf(constructor)

		returnType := constructorType.Out(0)

		var paramTypes []reflect.Type

		for i := 0; i < constructorType.NumIn(); i++ {
			paramType := constructorType.In(i)
			paramTypes = append(paramTypes, paramType)
		}

		dep := &dependency{
			ConstructorValue: constructorValue,
			ParameterTypes:   paramTypes,
			ReturnType:       returnType,
		}

		container.dependencies = append(container.dependencies, dep)
	}
}

func (container *Container) findDependencies(targetType reflect.Type) []*dependency {
	var deps []*dependency

	for _, dep := range container.dependencies {
		exactType := dep.ReturnType == targetType

		if exactType {
			deps = append(deps, dep)
			continue
		}

		implementingType := exactType || (targetType.Kind() == reflect.Interface && dep.ReturnType.Implements(targetType))

		if implementingType {
			deps = append(deps, dep)
			continue
		}
	}

	return deps
}

func typeResolved(valuesByType valuesByTypeMap, targetType reflect.Type) bool {
	_, ok := valuesByType[targetType]
	return ok
}

type valuesByTypeMap map[reflect.Type][]reflect.Value

func innerType(rawTargetType reflect.Type) reflect.Type {
	if rawTargetType.Kind() == reflect.Slice {
		return rawTargetType.Elem()
	}

	return rawTargetType
}

func dependencyResolved(container *Container, valuesByType valuesByTypeMap, dep *dependency) bool {
	for _, rawParamType := range dep.ParameterTypes {
		paramType := innerType(rawParamType)

		if paramType.Kind() == reflect.Slice {
			if len(valuesByType[paramType]) != len(container.findDependencies(paramType)) {
				return false
			}

			continue
		}

		if !typeResolved(valuesByType, paramType) {
			return false
		}
	}

	return true
}

func dependenciesResolved(container *Container, valuesByType valuesByTypeMap, deps []*dependency) bool {
	for _, dep := range deps {
		if !dependencyResolved(container, valuesByType, dep) {
			return false
		}
	}

	return true
}

func resolveDependency(valuesByType valuesByTypeMap, dep *dependency, targetType reflect.Type) {
	var arguments []reflect.Value

	for _, rawParamType := range dep.ParameterTypes {
		paramType := innerType(rawParamType)

		if rawParamType.Kind() == reflect.Slice {
			paramSliceValue := reflect.MakeSlice(reflect.SliceOf(paramType), 0, 0)
			paramSliceValue = reflect.Append(paramSliceValue, valuesByType[paramType]...)
			arguments = append(arguments, paramSliceValue)
			continue
		}

		if len(valuesByType[paramType]) > 1 {
			panic(fmt.Sprintf("Ambiguous dependency '%s'", paramType))
		}

		arguments = append(arguments, valuesByType[paramType][0])
	}

	value := dep.ConstructorValue.Call(arguments)[0]
	valuesByType[targetType] = append(valuesByType[targetType], value)
}

func resolveDependencies(valuesByType valuesByTypeMap, deps []*dependency, targetType reflect.Type) {
	for _, dep := range deps {
		resolveDependency(valuesByType, dep, targetType)
	}
}

// Resolve …
func (container *Container) Resolve(root interface{}) {
	rootType := reflect.TypeOf(root).Elem()

	queue := list.New()

	for i := 0; i < rootType.NumField(); i++ {
		structField := rootType.Field(i)
		structFieldType := structField.Type
		queue.PushBack(structFieldType)
	}

	valuesByType := make(valuesByTypeMap)
	visitedTypes := make(map[reflect.Type]bool)

	for queue.Len() > 0 {
		targetTypeElement := queue.Front()

		rawTargetType := targetTypeElement.Value.(reflect.Type)
		targetType := innerType(rawTargetType)

		visitedTypes[targetType] = true

		// Resolve type

		targetDeps := container.findDependencies(targetType)

		if len(targetDeps) == 0 && rawTargetType.Kind() != reflect.Slice {
			panic(fmt.Sprintf("No constructor defined for dependency '%s'", targetType))
		}

		if targetType.Kind() != reflect.Slice {
			if dependenciesResolved(container, valuesByType, targetDeps) {
				resolveDependencies(valuesByType, targetDeps, targetType)
				queue.Remove(targetTypeElement)
				continue
			}
		}

		// Resolve dependencies of type

		for _, dep := range targetDeps {
			for _, paramType := range dep.ParameterTypes {
				if visitedTypes[paramType] {
					panic(fmt.Sprintf("Cycle detected for dependency '%s'", paramType))
				}

				queue.PushBack(paramType)
			}
		}

		queue.Remove(targetTypeElement)
		queue.PushBack(rawTargetType)
	}

	for i := 0; i < rootType.NumField(); i++ {
		structField := rootType.Field(i)
		structFieldType := structField.Type

		rootValue := reflect.ValueOf(root).Elem()
		rootValue.Field(i).Set(valuesByType[structFieldType][0])
	}
}
