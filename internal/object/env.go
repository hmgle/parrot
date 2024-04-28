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
	currEnv := e
	for currEnv != nil {
		if obj, ok := currEnv.Store[name]; ok {
			return obj, true
		}
		currEnv = currEnv.Outer
	}
	return nil, false
}

func (e *Env) Set(name string, obj Object) Object {
	e.Store[name] = obj
	return obj
}

func (e *Env) Upsert(name string, obj Object) Object {
	currEnv := e
	for currEnv != nil {
		if _, ok := currEnv.Store[name]; ok {
			currEnv.Store[name] = obj
			return obj
		}
		currEnv = currEnv.Outer
	}
	e.Store[name] = obj
	return obj
}
