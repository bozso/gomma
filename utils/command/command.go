package command

import (
    "fmt"
    "os/exec"
    "strings"
)

type Command struct {
    cmd string
    debug bool
}

func New(cmd string) Command {
    return Command{
        cmd: cmd,
        debug: false,
    }
}

func (c Command) Debug() Command {
    return Command{
        cmd: c.cmd,
        debug: true,
    }
}

func (c Command) Call(args ...interface{}) (s string, err error) {
    arg := make([]string, len(args))

    for ii, elem := range args {
        arg[ii] = fmt.Sprint(elem)
    }
    
    return c.CallWithArgs(arg...)
}

func (c Command) CallWithArgs(args ...string) (s string, err error) {
    // fmt.Printf("%s %s\n", cmd, str.Join(arg, " "))
    // os.Exit(0)

    out, err := exec.Command(c.cmd, args...).CombinedOutput()
    s = string(out)
    
    if err != nil {
        err = Fail{cmd:c.cmd, out:s, args:args, err:err}
        return
    }

    return
}


const errorMessage = "Command '%s %s' failed!\nOutput of command is: %v"

type Fail struct{
    cmd, out string
    args []string
    err error
}

func (f Fail) Error() string {
    return fmt.Sprintf(errorMessage,
        f.cmd, strings.Join(f.args, " "), f.out)
}

func (f *Fail) Wrap(err error) error {
    f.err = err
    return f
}

func (f Fail) Unwrap() error {
    return f.err
}
