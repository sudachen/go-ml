package mx

import (
	"fmt"
	"github.com/sudachen/go-dnn/fu"
	"github.com/sudachen/go-ml/nn/mx/capi"
	"github.com/sudachen/go-ml/util"
	"runtime"
	"strings"
)

type GraphIdentity [20]byte // SHA1

type Loss interface {
	// out => loss
	Loss(*Symbol) *Symbol
}

type Param struct {
	Data     *NDArray
	Grad     *NDArray
	Shape    Dimension
	Autograd bool
}

type Graph struct {
	Ctx   Context
	Dtype Dtype

	Input   *NDArray  // network input referencing to Params["_input"]
	Output  *NDArray  // referencing to Outputs["_output_output"]
	Loss    *NDArray  // referencing to Outputs["_loss_loss"]
	Label   *NDArray  // loss function label referencing to Params["_label"]

	Outputs  map[string]*NDArray  // referencing to executor outputs except loss
	Params   map[string]*NDArray  // network parameters
	Shapes   map[string]Dimension // predefined param shape
	Autograd map[string]bool      // if param can be trained
	Grads    map[string]*NDArray  // training gradients

	Exec         capi.ExecutorHandle
	Initializers map[string]Inite
	Initialized  bool

	symOut, symLast capi.SymbolHandle

	vars    map[string]capi.SymbolHandle
	symbols map[*Symbol]capi.SymbolHandle
	auxs    []capi.NDArrayHandle

	identity *GraphIdentity
	alias    map[*Symbol]*Symbol
	outputs  map[string]*Symbol
	refs     map[string]capi.SymbolHandle
}

func (g *Graph) symRelease() {
	for _, v := range g.symbols {
		if v != g.symOut && v != g.symLast {
			capi.ReleaseSymbol(v)
		}
	}
	g.symbols = nil
	for _, v := range g.vars {
		capi.ReleaseSymbol(v)
	}
	g.vars = nil
	g.alias = nil
	g.refs = nil
	g.outputs = nil
}

func (g *Graph) Release() {
	runtime.SetFinalizer(g, nil)

	g.symRelease()
	if g.symLast != g.symOut {
		capi.ReleaseSymbol(g.symLast)
		g.symLast = nil
	}
	capi.ReleaseSymbol(g.symOut)
	g.symOut = nil

	capi.ReleaseExecutor(g.Exec)
	g.Exec = nil

	for _, v := range g.Params {
		v.Release()
	}
	g.Params = nil
	for _, v := range g.Grads {
		v.Release()
	}
	g.Grads = nil

	for _, v := range g.auxs {
		capi.ReleaseNDArry(v)
	}
}

func (g *Graph) allocate(shapes map[string][]int) {

	for n, s := range shapes {
		_, ok := g.Params[n]
		if !ok {
			if s2, ok := g.Shapes[n]; ok { s = s2.Slice() }
			a := g.Ctx.Array(g.Dtype, Dim(s...))
			g.Params[n] = a
		}
	}
}

func (g *Graph) GetShapes(withLoss bool) map[string][]int {
	sym := g.symLast
	if withLoss {
		sym = g.symOut
	}

	inter := capi.GetInternals(sym)
	x := map[string][]int{"_input": g.Input.Dim().Slice()}
	n := capi.ListNames(sym, capi.ArgumentsNames)

	for _, name := range n {
		if p, ok := g.Shapes[name]; ok && p.Len != 0 {
			x[name] = p.Slice()
		}
	}

	return capi.InferShapes(inter, x, capi.WithArguments|capi.WithOutputs)
}

func (g *Graph) bind() {
	input := g.Input.Dim()
	x := map[string][]int{"_input": input.Shape[:input.Len]}
	names := capi.ListNames(g.symOut, capi.ArgumentsNames)

	for _, n := range names {
		if p, ok := g.Shapes[n]; ok && p.Len != 0 {
			x[n] = p.Slice()
		}
	}

	shapes := capi.InferShapes(g.symOut, x, capi.WithArguments|capi.WithAuxStates|capi.WithoutOutput)
	g.allocate(shapes)
	args := make([]capi.NDArrayHandle, len(names))
	grads := make([]capi.NDArrayHandle, len(names))
	g.Input  = g.Params["_input"]
	g.Label  = g.Params["_label"]

	for i, name := range names {
		p := g.Params[name]
		if p != nil {
			args[i] = p.handle
			if g.symLast != g.symOut && g.Autograd[name] {
				a := g.Ctx.Array(g.Dtype,p.Dim())
				g.Grads[name] = a
				grads[i] = a.handle
			}
		}
	}

	auxnam := capi.ListNames(g.symOut, capi.AuxNames)
	aux := make([]capi.NDArrayHandle, len(auxnam))
	for i, name := range auxnam {
		if p, ok := g.Params[name]; ok {
			aux[i] = p.handle
		}
	}

	g.Exec = capi.Bind(g.symOut, g.Ctx.DevType(), g.Ctx.DevNo(), args, grads, aux)
	o := capi.GetOutputs(g.Exec)
	names = capi.ListNames(g.symOut, capi.OutputNames)
	g.Outputs = make(map[string]*NDArray)
	for i, n := range names {
		v := o[i]
		if strings.HasSuffix(n,"_output") {
			n = strings.TrimSuffix(n, "_output")
		} else {
			n = strings.TrimSuffix(n, "_loss")
		}
		g.Outputs[n] = &NDArray{handle: v.Handle, ctx: g.Ctx, dim: Dim(v.Dim...), dtype: Dtype(v.Type)}
	}
	if g.symLast != g.symOut {
		g.Loss = g.Outputs["_loss"]
	}
	g.Output = g.Outputs["_output"]
}

func Compose(
	ctx Context,
	sym *Symbol,
	loss Loss,
	input Dimension,
	dtype Dtype) *Graph {

	g := &Graph{
		Ctx:          ctx,
		Dtype:        dtype,
		Params:       make(map[string]*NDArray),
		Grads:        make(map[string]*NDArray),
		Autograd:     make(map[string]bool),
		Shapes:       make(map[string]Dimension),
		symbols:      make(map[*Symbol]capi.SymbolHandle),
		vars:         make(map[string]capi.SymbolHandle),
		refs:         make(map[string]capi.SymbolHandle),
		alias:        make(map[*Symbol]*Symbol),
		outputs:      make(map[string]*Symbol),
		Initializers: make(map[string]Inite),
	}

	g.Input = ctx.Array(dtype, input)
	_ = g.compose(Var("_input"))

	//Out := MakeLoss(BlockGrad(sym))
	Out := BlockGrad(sym)
	Out.SetName("_output")
	last := g.compose(Out)
	out := last

	if loss != nil {
		symloss := loss.Loss(sym)
		Loss := MakeLoss(symloss)
		Loss.SetName("_loss")
		_,_ = g.compose(symloss)
		others := util.ValsOf(g.outputs).([]*Symbol)
		outs := append([]*Symbol{Out,Loss},others...)
		out = g.compose(Group(outs...))
		if len(others) > 0 {
			outs := append([]*Symbol{Out},others...)
			last = g.compose(Group(outs...))
		}
	} else if len(g.outputs) > 0 {
		others := util.ValsOf(g.outputs).([]*Symbol)
		outs := append([]*Symbol{Out},others...)
		last = g.compose(Group(outs...))
		out = last
	}

	g.symLast = last
	g.symOut = out
	g.symRelease() // other symbols are not necessary more

	g.bind()

	runtime.SetFinalizer(g, func(g *Graph) { g.Release() })
	return g
}

func (g *Graph) subcompose(s *Symbol) []capi.SymbolHandle {
	var a []capi.SymbolHandle

	for _, v := range s.args {
		h := g.compose(v)
		if h != nil {
			a = append(a, h)
		}
	}
	return a
}

func (g *Graph) compose(s *Symbol) capi.SymbolHandle {

	if a, ok := g.alias[s]; ok {
		return g.symbols[a]
	}
	if h, ok := g.symbols[s]; ok {
		return h
	}

	switch s.op {
	case OpInput_:
		return g.vars["_input"]
	case OpScalar_:
		return nil
	case OpRef_:
		if s, ok := g.refs[s.name]; ok {
			return s
		}
		panic(fmt.Sprintf("symbol %s does not exist", s.name))
	case OpVar_, OpNogVar_:
		n := s.name
		if v, ok := g.vars[n]; ok {
			return v
		}
		h := capi.CreateVariable(n)
		g.vars[n] = h
		g.refs[n] = h
		if s.init != nil {
			g.Initializers[n] = s.init
		}
		if s.op != OpNogVar_ && n[0] != '_' {
			g.Autograd[n] = true
		}
		if s.dim.Len > 0 {
			g.Shapes[n] = s.dim.Like(g.Input.Dim())
		}
		return h
	case OpOutput_:
		n := "*"+s.name
		if _,ok := g.outputs[n]; !ok {
			g.outputs[n] = BlockGrad(s.args[0]).SetName(n)
		}
		return g.compose(s.args[0])
	case OpBound_:
		h := g.compose(s.args[0])
		for _, v := range s.args[1:] {
			_ = g.compose(v)
		}
		return h
	case OpDepend_:
		for _, v := range s.args[1:] {
			_ = g.compose(v)
		}
		return g.compose(s.args[0])
	case capi.OpZeros, capi.OpOnes, capi.OpRandomUniform, capi.OpReshape:
		s1 := *s
		s1.attr = make(map[capi.MxnetKey]string)
		for key, value := range s.attr {
			s1.attr[key] = value
		}
		s1.attr[capi.KeyShape] = s.dim.Like(g.Input.Dim()).String()
		a := &s1
		g.alias[s] = a
		s = a
	}

	var op capi.SymbolHandle

	a := g.subcompose(s)

	if s.op == OpGroup_ {
		op = capi.GroupSymbols(a)
		g.symbols[s] = op
	} else {

		op = capi.NewSymbol(s.op, s.attr)
		g.symbols[s] = op
		name := s.name
		if len(name) < 3 {
			name = fmt.Sprintf("%s@%s%02d", s.op.Value(), "sym", NextSymbolId())
		}
		capi.ComposeSymbol(op, name, a...)

		if s.name != "" {
			g.refs[s.name] = op
		}

		if s.output {
			n := "*"+name
			if _,ok := g.outputs[n]; !ok {
				g.outputs[n] = BlockGrad(s).SetName(n)
			}
		}
	}

	return op
}

func (g *Graph) InitParam(name string) {
	param := g.Params[name]
	if i, ok := g.Initializers[name]; ok && i != nil {
		i.Inite(param)
	} else {
		if name[0] == '_' {
			param.Zeros()
		} else if strings.Index(name, "_bias") >= 0 {
			param.Zeros()
		} else {
			param.Xavier(false, 2, 3)
		}
	}
}

func (g *Graph) Initialize(inite func(*NDArray,string)) {
	keys := fu.SortedDictKeys(g.Params)
	for _, name := range keys {
		if inite != nil {
			param := g.Params[name]
			inite(param, name)
		} else {
			g.InitParam(name)
		}
	}
	g.Initialized = true
}

func (g *Graph) Forward(train bool) {
	if !g.Initialized {
		g.Initialize(nil)
	}
	capi.Forward(g.Exec, train)
}

func (g *Graph) Backward() {
	capi.Backward(g.Exec)
}
