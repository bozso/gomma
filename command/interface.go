package command

/*
Into represents any type that can be converted into an environment.Env
variable.
*/
type Into interface {
    IntoEnv() (Env, error)
}
