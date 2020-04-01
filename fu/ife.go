package fu

/*
Ife returns x if expr == true else returns y
*/
func Ife(expr bool, x interface{}, y interface{}) interface{} {
	if expr {
		return x
	}
	return y
}

/*
Ifei returns x if expr == true else returns y
*/
func Ifei(expr bool, x int, y int) int {
	if expr {
		return x
	}
	return y
}

/*
Ifel returns x if expr == true else returns y
*/
func Ifel(expr bool, x int64, y int64) int64 {
	if expr {
		return x
	}
	return y
}

/*
Ifer returns x if expr == true else returns y
*/
func Ifer(expr bool, x float32, y float32) float32 {
	if expr {
		return x
	}
	return y
}

/*
Ifed returns x if expr == true else returns y
*/
func Ifed(expr bool, x float64, y float64) float64 {
	if expr {
		return x
	}
	return y
}

/*
Ifes returns x if expr == true else returns y
*/
func Ifes(expr bool, x string, y string) string {
	if expr {
		return x
	}
	return y
}
