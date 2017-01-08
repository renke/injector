# injector

injector is a constructor-based dependency injection library for Go.

Some features and remarks:

* Only supports constructor injection.
* Injection based on concrete type, interface or slice of interface.
* No need to use exported struct fields.
* Injected dependencies are always singletons.
* Encourages a lot of constructor functions. Probably not idiomatic Go.
* Contains a few known bugs. Mostly likely has more.
* Panics when something bad happens.
* Lacks useful documentation.

# Installation

```
$ go get github.com/renke/injector
```

# Usage

```go
package main

import (
	"fmt"

	"github.com/renke/injector"
)

type Foo struct {
	name string
}

func NewFoo() *Foo {
	return &Foo{
		name: "Foo",
	}
}

type Bar struct {
	name string
}

func NewBar() *Bar {
	return &Bar{
		name: "Bar",
	}
}

type Baz struct {
	foo *Foo
	bar *Bar
}

func NewBaz(foo *Foo, bar *Bar) *Baz {
	return &Baz{
		foo: foo,
		bar: bar,
	}
}

func (baz *Baz) Print() {
	fmt.Println(baz.foo.name, baz.bar.name, "Baz")
}

type App struct {
	Baz *Baz
}

func main() {
	container := injector.NewContainer()

	container.Register(NewFoo)
	container.Register(NewBar)
	container.Register(NewBaz)

	var app App

	container.Resolve(&app)

	app.Baz.Print()
}
```

See [test cases](injector_test.go) for more "examples".

# Feedback

I appreciate any kind of feedback. Just create an issue or drop me a mail. Thanks!

# License

See [LICENSE](LICENSE).
