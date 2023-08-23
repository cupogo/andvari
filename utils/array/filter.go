package array

func FilterFunc[S ~[]E, E any](s S, f func(E) bool) (o S) {
	for i := range s {
		v := s[i]
		if f(v) {
			o = append(o, v)
		}
	}
	return
}
