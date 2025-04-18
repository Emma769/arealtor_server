package validator

type ValidationErrors map[string]string

type Validator struct {
	errs ValidationErrors
}

type ValidationMsg struct {
	Info, Prop string
}

func New() *Validator {
	return &Validator{
		errs: ValidationErrors{},
	}
}

func (v Validator) Valid() bool {
	return len(v.errs) == 0
}

func (v *Validator) add(msg ValidationMsg) {
	if _, ok := v.errs[msg.Prop]; !ok {
		v.errs[msg.Prop] = msg.Info
	}
}

func (v Validator) Err() ValidationErrors {
	return v.errs
}

type ValidationFn[T any] func(T) (bool, ValidationMsg)

func Check[T any](v *Validator, t T, fns ...ValidationFn[T]) {
	for i := range fns {
		ok, msg := fns[i](t)
		if !ok {
			v.add(msg)
		}
	}
}
