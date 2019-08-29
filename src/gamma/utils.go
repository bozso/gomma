package gamma;

import (
    "fmt";
    "log";
    "os/exec";
    "os";
    io "io/ioutil";
)


type CmdFun func(args ...interface{}) string;

const cmdErr = 
`Program "%v" exited with Error: %v
Output of command: %v`;


func MakeCmd(cmd string) CmdFun {
    return func (args ...interface{}) string {
        arg := make([]string, len(args))
        
        for ii, elem := range args {
            if elem != nil {
                arg[ii] = fmt.Sprint(elem);
            } else {
                arg[ii] = "-";
            }
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

type Tmp struct {
    files []string;
}


var tmp = Tmp{};


func TmpFile() string {
    file, err := io.TempFile("", "*");
    Check(err, "Could not create temporary file!");
    defer file.Close()
    
    name := file.Name();
    
    tmp.files = append(tmp.files, name);
    
    return name;
}


func RemoveTmp() {
    for _, file := range tmp.files {
        os.Remove(file);
    }
}

