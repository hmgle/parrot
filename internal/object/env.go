package object

type Env struct {
	Store map[string]Object
	Outer *Env
}

func NewEnv() *Env {
	return &Env{
		Store: make(map[string]Object),
		Outer: nil,
	}
}

func NewEnvWrap(e *Env) *Env {
	return &Env{
		Store: make(map[string]Object),
		Outer: e,
	}
}

func (e *Env) Get(name string) (Object, bool) {
	obj, ok := e.Store[name]
	if !ok && e.Outer != nil {
		return e.Outer.Get(name)
	}
	return obj, ok
}

func (e *Env) Set(name string, obj Object) Object {
	e.Store[name] = obj
	return obj
}
