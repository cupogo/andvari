package utils

func Refer[V comparable](v V) *V {
	return &v
}
