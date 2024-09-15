package genstrument

func SetStringAttribute[S ~string](str S, setter AttributeSetter) {
	setter.String(string(str))
}

func SetIntAttribute[I ~int | ~int8 | ~int16 | ~int32 | ~int64](i I, setter AttributeSetter) {
	setter.Int64(int64(i))
}

func SetBoolAttribute[B ~bool](b B, setter AttributeSetter) {
	setter.Bool(bool(b))
}

func SetFloatAttribute[F ~float32 | ~float64](f F, setter AttributeSetter) {
	setter.Float64(float64(f))
}

func SetErrorAttribute(err error, setter AttributeSetter) {
	if err == nil {
		return
	}
	setter.Error(err)
}
