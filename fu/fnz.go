package fu

import "reflect"

/*
Fnz returns the first non zero value
*/
func Fnz(a ...interface{}) interface{} {
	for _, i := range a {
		if !reflect.ValueOf(i).IsZero() {
			return i
		}
	}
	return 0
}

/*
Fnzi returns the first non integer zero value
*/
func Fnzi(a ...int) int {
	for _, i := range a {
		if i != 0 {
			return i
		}
	}
	return 0
}

/*
Fnzl returns the first non zero long integer value
*/
func Fnzl(a ...int64) int64 {
	for _, i := range a {
		if i != 0 {
			return i
		}
	}
	return 0
}

/*
Fnzr returns the first non zero float value
*/
func Fnzr(a ...float32) float32 {
	for _, i := range a {
		if i != 0 {
			return i
		}
	}
	return 0
}

/*
Fnzd returns the first non zero double value
*/
func Fnzd(a ...float64) float64 {
	for _, i := range a {
		if i != 0 {
			return i
		}
	}
	return 0
}

/*
Fnze returns the first non nil error
*/
func Fnze(e ...error) error {
	for _, i := range e {
		if i != nil {
			return i
		}
	}
	return nil
}

/*
Fnzs returns the first non empty string
*/
func Fnzs(e ...string) string {
	for _, i := range e {
		if i != "" {
			return i
		}
	}
	return ""
}

/*
Fnzb returns the first non false bool
*/
func Fnzb(e ...bool) bool {
	for _, i := range e {
		if i {
			return i
		}
	}
	return false
}
