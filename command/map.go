/*Packege environment implements setting up and handling environment
variables.

Any varible implementing the Formatter interface can be set as value to Map.

Example usage:
    envMap := environment.Map{
        "OS": environemnt.String("linux"),
        "GOLANG": environment.True,
        "CORES": environment.Int(6),
    }
    
    env := envMap.IntoEnv()
    // converted to environment.Env the result is
    // env = {}string{"OS=linux", "GOLANG", "CORES=6"}

    // alternatively the Map above could be created like so:
    envMap := make(environment.Map)

    envMap.Set("OS", "linux")
    envMap.Use("GOLANG")
    envMap.SetInt("CORES", 6)
*/
package command

import (
    "fmt"
)

/*
Formatter is the value type for Map.
*/
type Formatter interface {
    // FormatEnv will return with the string representation of an environment
    // variable given a specific key.
    FormatEnv(key string) string
}

type trueVal struct {}

func (t trueVal) FormatEnv(key string) (s string) {
    return key
}

/*
True represents that the environemnt variable is set.
*/
var True Formatter = trueVal{}

/*
String is an alias to string to implement the Formatter interface.
*/
type String string

// FormatEnv implements the Formatter interface
func (s String) FormatEnv(key string) (out string) {
    return fmt.Sprintf("%s=%s", key, s)
}

/*
Int is an alias to int to implement the Formatter interface.
*/
type Int int

// FormatEnv implements the Formatter interface
func (i Int) FormatEnv(key string) (out string) {
    return fmt.Sprintf("%s=%d", key, i)
}

/*
Map represent key value pairs of environment variables
*/
type Map map[string]Formatter

// Set sets key to String(val).
func (m Map) Set(key, val string) {
    m[key] = String(val)
}

// SetInt sets key to Int(val).
func (m Map) SetInt(key string, val int) {
    m[key] = Int(val)
}

// Use sets key to True.
func (m Map) Use(key string) {
    m[key] = True
}

// Remove removes a key from the map.
func (m Map) Remove(key string) {
    delete(m, key)
}

// IntoEnv converts m into an Env
func (m Map) IntoEnv() (env Env, err error) {
    s := make([]string, len(m))

    ii := 0
    for key, val := range m {
        s[ii] = val.FormatEnv(key)
    }

    return NewEnvironment(s), nil
}
