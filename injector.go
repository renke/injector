package injector

import (
	"container/list"
	"fmt"
	"reflect"
)

// Container keeps track of all dependencies that were registered.
type Container struct {
	constructors []*constructor
}

// NewContainer creates a new empty container.
func NewContainer() *Container {
	return &Container{}
}

// Register registers new dependencies based on constructor functions. A
// constructor is a function that takes zero or more parameters and returns
// exactly one dependency as value.
//
//   func NewBaz(foo *Foo, bar *Bar) *Baz {…}
//
// The paramters type can be one of the following:
//
// A pointer to a struct. Exactly one dependency must be registered that returns
// an instance of that struct.
//
//   func NewBar(foo *Foo) *Baz {…} // Inject Foo dependency
//
// An interface. Exactly one dependency must be registered that returns
// an instance of a struct that implements that interface.
//
//   func NewBar(foo Foo) *Baz {…} // Inject dependency that implements the Foo interface
//
// A slice type using an interface. At least one dependency must be registered that returns
// an instance of a struct that implements that interface.
//
//   func NewBar(foos []Foo) *Baz {…} // Inject all dependencies that implement the Foo interface
func (container *Container) Register(constructors ...interface{}) {
	for _, _constructor := range constructors {
		_type := reflect.TypeOf(_constructor)

		if _type.Kind() != reflect.Func {
			panic(fmt.Sprintf("Constructor '%s' is not a function", _type))
		}

		if _type.NumOut() != 1 {
			panic(fmt.Sprintf("Constructor '%s' must have single return value", _type))
		}

		function := reflect.ValueOf(_constructor)

		var params []reflect.Type

		for i := 0; i < _type.NumIn(); i++ {
			param := _type.In(i)
			params = append(params, param)
		}

		returnType := _type.Out(0)

		details := &constructor{
			Function:   function,
			Parameters: params,
			ReturnType: returnType,
		}

		container.constructors = append(container.constructors, details)
	}
}

// Resolve wires together the object graph starting with the fields
// in the given struct instance.
func (container *Container) Resolve(root interface{}) {
	resolver := newResolver(container)

	rootType := reflect.TypeOf(root).Elem()

	// Add all types that should be resolved
	for i := 0; i < rootType.NumField(); i++ {
		structField := rootType.Field(i)
		structFieldType := structField.Type

		resolver.resolveType(structFieldType)
	}

	// Resolve all types that were added initially
	for i := 0; i < rootType.NumField(); i++ {
		structField := rootType.Field(i)
		structFieldType := structField.Type

		rootValue := reflect.ValueOf(root).Elem()
		rootValue.Field(i).Set(resolver.ValuesByType[structFieldType][0])
	}
}

func (container *Container) findConstructors(_type reflect.Type) []*constructor {
	var constructors []*constructor

	for _, constructor := range container.constructors {
		if constructor.ReturnType.AssignableTo(_type) {
			constructors = append(constructors, constructor)
		}
	}

	return constructors
}

type constructor struct {
	Function   reflect.Value
	Parameters []reflect.Type
	ReturnType reflect.Type
}

type valuesByType map[reflect.Type][]reflect.Value

func (_valuesByType valuesByType) hasValue(_type reflect.Type, value reflect.Value) bool {
	for _, value := range _valuesByType[_type] {
		if value == value {
			return true
		}
	}

	return false
}

type resolver struct {
	Container *Container

	ValuesByType       valuesByType
	VisitedTypes       map[reflect.Type]bool
	ValueByConstructor map[*constructor]reflect.Value
}

func (_resolver *resolver) resolveType(_type reflect.Type) {
	stack := list.New()
	stack.PushFront(_type)

	container := _resolver.Container

	for stack.Len() > 0 {
		typeElement := stack.Front()

		rawType := typeElement.Value.(reflect.Type)
		_type := innerType(rawType)

		_resolver.VisitedTypes[_type] = true

		// Resolve type by invoking all its constructors

		constructors := container.findConstructors(_type)

		if len(constructors) == 0 && rawType.Kind() != reflect.Slice {
			panic(fmt.Sprintf("No constructor defined for type '%s'", _type))
		}

		var pendingConstructors []*constructor

		if _type.Kind() != reflect.Slice {
			for _, constructor := range constructors {
				if value, ok := _resolver.constructorInvoked(constructor, _type); ok {
					if !_resolver.ValuesByType.hasValue(_type, value) {
						_resolver.ValuesByType[_type] = append(_resolver.ValuesByType[_type], value)
					}
				} else if _resolver.constructorInvokable(constructor) {
					value := _resolver.invokeConstructor(constructor, _type)
					_resolver.ValuesByType[_type] = append(_resolver.ValuesByType[_type], value)
					_resolver.ValueByConstructor[constructor] = value
				} else {
					pendingConstructors = append(pendingConstructors, constructor)
				}
			}

			if len(pendingConstructors) == 0 {
				stack.Remove(typeElement)
				continue
			}
		}

		// Resolve missing parameters of pending constructors

		for _, dep := range pendingConstructors {

			for _, param := range dep.Parameters {
				visited := _resolver.VisitedTypes[param]
				resolved := _resolver.typeResolved(param)

				if visited && !resolved {
					panic(fmt.Sprintf(
						"Cycle detected for parameter '%s' of constructor '%s' while resolving type '%s'.",
						param, dep.Function.Type(), _type,
					))
				}

				if !visited {
					stack.PushFront(param)
				}
			}
		}
	}
}

func newResolver(container *Container) *resolver {
	return &resolver{
		Container: container,

		ValuesByType:       make(map[reflect.Type][]reflect.Value),
		VisitedTypes:       make(map[reflect.Type]bool),
		ValueByConstructor: make(map[*constructor]reflect.Value),
	}
}

func (_resolver *resolver) typeResolved(_type reflect.Type) bool {
	_, ok := _resolver.ValuesByType[_type]
	return ok
}

func innerType(rawType reflect.Type) reflect.Type {
	if rawType.Kind() == reflect.Slice {
		return rawType.Elem()
	}

	return rawType
}

func (_resolver *resolver) constructorInvokable(constructor *constructor) bool {
	for _, rawParam := range constructor.Parameters {
		param := innerType(rawParam)

		if param.Kind() == reflect.Slice {
			if len(_resolver.ValuesByType[param]) != len(_resolver.Container.findConstructors(param)) {
				return false
			}

			continue
		}

		if !_resolver.typeResolved(param) {
			return false
		}
	}

	return true
}

func (_resolver *resolver) invokeConstructor(constructor *constructor, _type reflect.Type) reflect.Value {
	var arguments []reflect.Value

	for _, rawParam := range constructor.Parameters {
		param := innerType(rawParam)

		if rawParam.Kind() == reflect.Slice {
			paramSliceValue := reflect.MakeSlice(reflect.SliceOf(param), 0, 0)
			paramSliceValue = reflect.Append(paramSliceValue, _resolver.ValuesByType[param]...)
			arguments = append(arguments, paramSliceValue)
			continue
		}

		if len(_resolver.ValuesByType[param]) > 1 {
			panic(fmt.Sprintf("Ambiguity detected for type '%s'", param))
		}

		arguments = append(arguments, _resolver.ValuesByType[param][0])
	}

	value := constructor.Function.Call(arguments)[0]
	return value
}

func (_resolver *resolver) constructorInvoked(constructor *constructor, _type reflect.Type) (reflect.Value, bool) {
	value, ok := _resolver.ValueByConstructor[constructor]
	return value, ok
}
