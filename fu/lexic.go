package fu

type Lexic []func(string) bool

func (lx Lexic) Accepted(s string, dflt ...bool) bool {
	if len(lx) == 0 {
		return Fnzb(dflt...)
	}
	for _, x := range lx {
		if x(s) {
			return true
		}
	}
	return false
}
