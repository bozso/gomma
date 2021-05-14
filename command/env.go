package command

var emptySlice = []string{}

// EmptyEnv is an environment with no values.
var EmptyEnv = NewEnvironment(emptySlice)

/*
Env is a struct representing a set of environment values.
*/
type Env struct {
	env []string
}

// NewEnvironment creates an Env out of a string slice.
func NewEnvironment(env []string) (e Env) {
	return Env{env}
}

/*
From uses the implemented IntoEnv function to convert i into an Env variable.
*/
func From(i Into) (e Env, err error) {
	return i.IntoEnv()
}

/*
Get returns the string slice representing the set of environment variables.
*/
func (e Env) Get() (s []string) {
	return e.env
}
