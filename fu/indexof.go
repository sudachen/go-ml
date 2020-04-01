package fu

func IndexOf(a string, b []string) int {
	for i, v := range b {
		if v == a {
			return i
		}
	}
	return -1
}
