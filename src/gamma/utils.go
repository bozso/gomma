package gamma;

import (
    "fmt";
    "log";
    "os/exec";
    "os";
    io "io/ioutil";
    str "strings";
    conv "strconv";
);



type (
    CmdFun func(args ...interface{}) string;
    
    Params struct {
        par, sep string;
        contents []string;
    };

    Tmp struct {
        files []string;
    };
);


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


func ReadFile(path string) []byte {
    f, err := os.Open(path)
    Check(err, "Could not open file: %v", path);
    defer f.Close();
    
    contents, err := io.ReadAll(f);
    Check(err, "Could not read file: %v", path);
    
    return contents;
}


func FromFile(path, sep string) Params {
    return Params{par:path, sep:sep, 
                  contents:str.Split(string(ReadFile(path)[:]), "\n")};
}


func FromString(contents, sep string) Params {
    return Params{sep:sep, contents:str.Split(contents, "\n")};
}


func (self *Params) Par(name string) string {
    for _, line := range self.contents {
        if str.Contains(line, name) {
            return str.Trim(str.Split(line, self.sep)[1], " ");
        }
    }
    log.Fatalf("Could not find parameter '%s' in %v", name, self.par);
    return "";
}


func toInt(par string, idx int) int {
    ret, err := conv.Atoi(str.Split(par, " ")[idx]);
    Check(err, "Could not convert string %s to int!", par);
    return ret;
}


func toFloat(par string, idx int) float64 {
    ret, err := conv.ParseFloat(str.Split(par, " ")[idx], 64);
    Check(err, "Could not convert string %s to float64!", par);
    return ret;
}

func (self *Params) Int(name string) int {
    return toInt(self.Par(name), 0);
}

func (self *Params) Float(name string) float64 {
    return toFloat(self.Par(name), 0);
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
