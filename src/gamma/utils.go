package gamma;

import (
    "fmt";
    str "strings";
)

type CmdFun func(args ...interface{})

func MakeCmd(cmd string) CmdFun {
    return func (args ...interface{}) {
        arg := make([]string, len(args))
        
        for ii, elem := range args {
            arg[ii] = fmt.Sprint(elem);
        }
        
        fmt.Println(str.Join(arg, " "))
    };
}
