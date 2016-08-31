package injector

import (
	"testing"
	"time"
)

type App struct {
	Foo *Foo
}

type Foo struct {
	bar *Bar
}

type Bar struct {
	value string
}

func NewFoo(bar *Bar) *Foo {
	return &Foo{
		bar: bar,
	}
}

func NewBar() *Bar {
	return &Bar{
		value: "bar",
	}
}

func Test(t *testing.T) {
	container := NewContainer()

	container.Register(NewFoo, NewBar)

	app := &App{}
	container.Resolve(app)

	if app.Foo.bar.value != "bar" {
		t.Errorf("Foo could not be resolved")
	}
}

type CycleApp struct {
	Foo *CycleFoo
}

type CycleFoo struct {
	bar *CycleBar
}

type CycleBar struct {
	foo   *CycleFoo
	value string
}

func NewCycleFoo(bar *CycleBar) *CycleFoo {
	return &CycleFoo{
		bar: bar,
	}
}

func NewCycleBar(foo *CycleFoo) *CycleBar {
	return &CycleBar{
		value: "bar",
	}
}

func TestCycle(t *testing.T) {
	container := NewContainer()

	container.Register(NewCycleFoo, NewCycleBar)

	app := &CycleApp{}

	done := make(chan struct{})

	go func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Cycle was not detected")
			}

			done <- struct{}{}
		}()

		container.Resolve(app)
	}()

	select {
	case <-done:
		break
	case <-time.After(time.Second):
		t.Errorf("Cycle was not detected")
	}
}

type MissingApp struct {
	Foo *MissingFoo
}

type MissingFoo struct {
	bar *MissingBar
}

type MissingBar struct {
	value string
}

func NewMissingFoo(bar *MissingBar) *MissingFoo {
	return &MissingFoo{
		bar: bar,
	}
}

func TestMissing(t *testing.T) {
	container := NewContainer()

	container.Register(NewMissingFoo)

	app := &MissingApp{}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Missing dependency was not detected")
		}
	}()

	container.Resolve(app)
}

type InterfaceApp struct {
	Foo *InterfaceFoo
}

type InterfaceFoo struct {
	bar InterfaceBar
}

type InterfaceBar interface {
	Bar() string
}

type InterfaceActualBar struct {
}

func (bar *InterfaceActualBar) Bar() string {
	return "bar"
}

func NewInterfaceFoo(bar InterfaceBar) *InterfaceFoo {
	return &InterfaceFoo{
		bar: bar,
	}
}

func NewInterfaceActualBar() *InterfaceActualBar {
	return &InterfaceActualBar{}
}

func TestInterface(t *testing.T) {
	container := NewContainer()

	container.Register(NewInterfaceFoo, NewInterfaceActualBar)

	app := &InterfaceApp{}
	container.Resolve(app)

	if app.Foo.bar.Bar() != "bar" {
		t.Errorf("Foo could not be resolved")
	}
}

type ImplementionApp struct {
	Foo *ImplementionFoo
}

type ImplementionFoo struct {
	bars []ImplementionBar
}

type ImplementionBar interface {
	Bar() string
}

type ImplementionFirstBar struct {
}

func (bar *ImplementionFirstBar) Bar() string {
	return "first_bar"
}

type ImplementionSecondBar struct {
}

func (bar *ImplementionSecondBar) Bar() string {
	return "second_bar"
}

func NewImplementionFoo(bars []ImplementionBar) *ImplementionFoo {
	return &ImplementionFoo{
		bars: bars,
	}
}

func NewImplementionFirstBar() *ImplementionFirstBar {
	return &ImplementionFirstBar{}
}

func NewImplementionSecondBar() *ImplementionSecondBar {
	return &ImplementionSecondBar{}
}

func TestImplemention(t *testing.T) {
	container := NewContainer()

	container.Register(NewImplementionFoo, NewImplementionFirstBar, NewImplementionSecondBar)

	app := &ImplementionApp{}
	container.Resolve(app)

	bars := app.Foo.bars

	if len(bars) != 2 || bars[0].Bar() != "first_bar" || bars[1].Bar() != "second_bar" {
		t.Errorf("Foo could not be resolved")
	}
}

type AmbiguousApp struct {
	Foo *AmbiguousFoo
}

type AmbiguousFoo struct {
	bar AmbiguousBar
}

type AmbiguousBar interface {
	Bar() string
}

type AmbiguousFirstBar struct {
}

func (bar *AmbiguousFirstBar) Bar() string {
	return "first_bar"
}

type AmbiguousSecondBar struct {
}

func (bar *AmbiguousSecondBar) Bar() string {
	return "second_bar"
}

func NewAmbiguousFoo(bar AmbiguousBar) *AmbiguousFoo {
	return &AmbiguousFoo{
		bar: bar,
	}
}

func NewAmbiguousFirstBar() *AmbiguousFirstBar {
	return &AmbiguousFirstBar{}
}

func NewAmbiguousSecondBar() *AmbiguousSecondBar {
	return &AmbiguousSecondBar{}
}

func TestAmbiguous(t *testing.T) {
	container := NewContainer()

	container.Register(NewAmbiguousFoo, NewAmbiguousFirstBar, NewAmbiguousSecondBar)

	app := &AmbiguousApp{}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Ambiguous dependency was not detected")
		}
	}()

	container.Resolve(app)
}

type PointerApp struct {
	Foo PointerFoo
}

type PointerFoo struct {
	bar PointerBar
}

type PointerBar struct {
	value string
}

func NewPointerFoo(bar PointerBar) PointerFoo {
	return PointerFoo{
		bar: bar,
	}
}

func NewPointerBar() PointerBar {
	return PointerBar{
		value: "bar",
	}
}

func TestPointer(t *testing.T) {
	container := NewContainer()

	container.Register(NewPointerFoo, NewPointerBar)

	app := &PointerApp{}
	container.Resolve(app)

	if app.Foo.bar.value != "bar" {
		t.Errorf("Foo could not be resolved")
	}
}

type SimpleRootApp struct {
	Foo *SimpleRootFoo
	Bar *SimpleRootBar
	Baz *SimpleRootBaz
}

type SimpleRootFoo struct {
	value string
}

type SimpleRootBar struct {
	value string
}

type SimpleRootBaz struct {
	foo   *SimpleRootFoo
	bar   *SimpleRootBar
	value string
}

func NewSimpleRootFoo() *SimpleRootFoo {
	return &SimpleRootFoo{
		value: "foo",
	}
}

func NewSimpleRootBar() *SimpleRootBar {
	return &SimpleRootBar{
		value: "bar",
	}
}

func NewSimpleRootBaz(foo *SimpleRootFoo, bar *SimpleRootBar) *SimpleRootBaz {
	return &SimpleRootBaz{
		foo:   foo,
		bar:   bar,
		value: "baz",
	}
}

func TestSimpleRoot(t *testing.T) {
	container := NewContainer()

	container.Register(NewSimpleRootBaz, NewSimpleRootBar, NewSimpleRootFoo)

	app := &SimpleRootApp{}
	container.Resolve(app)

	if app.Foo.value != "foo" {
		t.Errorf("Foo could not be resolved")
	}

	if app.Bar.value != "bar" {
		t.Errorf("Bar could not be resolved")
	}

	if app.Baz.value != "baz" || app.Baz.foo != app.Foo || app.Baz.bar != app.Bar {
		t.Errorf("Baz could not be resolved")
	}
}

type PolymorphApp struct {
	Foo      *PolymorphFoo
	OtherFoo PolymorphFooInterface
}

type PolymorphFoo struct {
	value string
}

func (foo *PolymorphFoo) Value() string {
	return "foo"
}

type PolymorphFooInterface interface {
	Value() string
}

func NewPolymorphFoo() *PolymorphFoo {
	return &PolymorphFoo{
		value: "foo",
	}
}

func TestPolymorph(t *testing.T) {
	container := NewContainer()

	container.Register(NewPolymorphFoo)

	app := &PolymorphApp{}
	container.Resolve(app)

	if app.Foo.value != "foo" {
		t.Errorf("Foo could not be resolved")
	}

	if app.OtherFoo.Value() != "foo" {
		t.Errorf("OtherFoo could not be resolved")
	}
}

type ComplexRootApp struct {
	Foo *ComplexRootFoo
	Bar *ComplexRootBar
	Baz *ComplexRootBaz
}

type ComplexRootFoo struct {
	value string
}

type ComplexRootBar struct {
	foo   *ComplexRootFoo
	value string
}

type ComplexRootBaz struct {
	foo   *ComplexRootFoo
	value string
}

func NewComplexRootFoo() *ComplexRootFoo {
	return &ComplexRootFoo{
		value: "foo",
	}
}

func NewComplexRootBar(foo *ComplexRootFoo) *ComplexRootBar {
	return &ComplexRootBar{
		foo:   foo,
		value: "bar",
	}
}

func NewComplexRootBaz(foo *ComplexRootFoo) *ComplexRootBaz {
	return &ComplexRootBaz{
		foo:   foo,
		value: "baz",
	}
}

func TestComplexRoot(t *testing.T) {
	container := NewContainer()

	container.Register(NewComplexRootBaz, NewComplexRootBar, NewComplexRootFoo)

	app := &ComplexRootApp{}
	container.Resolve(app)

	if app.Foo.value != "foo" {
		t.Errorf("Foo could not be resolved")
	}

	if app.Bar.value != "bar" || app.Bar.foo != app.Foo {
		t.Errorf("Bar could not be resolved")
	}

	if app.Baz.value != "baz" || app.Baz.foo != app.Foo {
		t.Errorf("Baz could not be resolved")
	}
}

type MultiApp struct {
	Foo *MultiFoo
}

type MultiFoo struct {
	bars []MultiBar
}

type MultiBar interface {
	Bar() string
}

type MultiFirstBar struct {
}

func (bar *MultiFirstBar) Bar() string {
	return "first_bar"
}

type MultiSecondBar struct {
}

func (bar *MultiSecondBar) Bar() string {
	return "second_bar"
}

func NewMultiFoo(bars []MultiBar) *MultiFoo {
	return &MultiFoo{
		bars: bars,
	}
}

func NewMultiFirstBar() *MultiFirstBar {
	return &MultiFirstBar{}
}

func NewMultiSecondBar() *MultiSecondBar {
	return &MultiSecondBar{}
}

func TestMulti(t *testing.T) {
	container := NewContainer()

	container.Register(NewMultiFoo, NewMultiFirstBar, NewMultiSecondBar)

	app := &MultiApp{}
	container.Resolve(app)

	bars := app.Foo.bars

	if len(bars) != 2 || bars[0].Bar() != "first_bar" || bars[1].Bar() != "second_bar" {
		t.Errorf("Foo could not be resolved")
	}
}
