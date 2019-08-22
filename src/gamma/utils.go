package gamma;

import (
    "fmt";
    "log";
    "os/exec";
)


type CmdFun func(args ...interface{}) string;

const cmdErr = 
`Program "%v" exited with Error: %v
Output of command: %v`;


func MakeCmd(cmd string) CmdFun {
    return func (args ...interface{}) string {
        arg := make([]string, len(args))
        
        for ii, elem := range args {
            arg[ii] = fmt.Sprint(elem);
        }
        
        out, err := exec.Command(cmd, arg...).CombinedOutput();
        
        result := string(out);
        
        if err != nil {
            log.Fatalf(cmdErr, cmd, err, result);
        }
        
        return result;
    };
}


func Check(err error, format string, args ...interface{}) {
    if err != nil {
        str := fmt.Sprintf(format, args...);
        log.Fatalf("Error: %s\nError: %s", str, err);
    }
}
