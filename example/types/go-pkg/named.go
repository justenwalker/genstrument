package gopkg

import "github.com/justenwalker/genstrument"

type GoType1 struct {
	Foo string
}

type GoType2 struct {
	Bar string
}

func GoType1Type1Attr(t GoType1, setter genstrument.AttributeSetter) {
	setter.String(t.Foo)
}

func GoType2Type2Attr(t GoType2, setter genstrument.AttributeSetter) {
	setter.String(t.Bar)
}
