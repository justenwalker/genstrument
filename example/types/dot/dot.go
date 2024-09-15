package dot

import "github.com/justenwalker/genstrument"

type Type1Dot struct {
	Foo string
}

func (t *Type1Dot) String() string {
	return t.Foo
}

type Type2Dot struct {
	Bar string
}

func (t *Type2Dot) String() string {
	return t.Bar
}

func Type1Attr(t Type1Dot, setter genstrument.AttributeSetter) {
	setter.String(t.Foo)
}

func Type2Attr(t Type2Dot, setter genstrument.AttributeSetter) {
	setter.String(t.Bar)
}
