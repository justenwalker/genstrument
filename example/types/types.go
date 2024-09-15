package types

import "github.com/justenwalker/genstrument"

type MyType struct {
	Foo    string
	Bar    string
	Secret string
}

func MyTypeAttr(mt MyType, setter genstrument.AttributeSetter) {
	setter.Attribute("foo").String(mt.Foo)
	setter.Attribute("bar").String(mt.Bar)
	setter.Attribute("secret").String("MASKED")
}
