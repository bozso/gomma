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
    CmdFun func(args ...interface{}) (string, error);
    handlerFun func(err error, format string, args ...interface{}) error;
    
    Params struct {
        par, sep string;
        contents []string;
    };

    Tmp struct {
        files []string;
    };
    
);


const cmdErr = 
`Command '%v' failed!
Output of command is: %v
%w`;


func Fatal(err error, format string, args ...interface{}) {
    if err != nil {
        str := fmt.Sprintf(format, args...);
        log.Fatalf("Error: %s\nError: %s", str, err);
    }
}


func Handler(name string) handlerFun {
    return func(err error, format string, args ...interface{}) error {
        str := fmt.Sprintf(format, args...);
        return fmt.Errorf("In %s: %s\nError: %w", name, str, err);
    }
}


func MakeCmd(cmd string) CmdFun {
    return func (args ...interface{}) (string, error) {
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
            return "", fmt.Errorf(cmdErr, cmd, result, err);
        }
        
        return result, nil;
    };
}


func ReadFile(path string) ([]byte, error) {
    handle := Handler("ReadFile");
    
    f, err := os.Open(path);
    if err != nil {
        return []byte{}, handle(err, "Could not open file: '%v'!", path)
    }
    
    defer f.Close();
    
    contents, err := io.ReadAll(f);
    if err != nil {
        return []byte{}, handle(err, "Could not read file: '%v'!", path);
    }
    
    return contents, nil;
}


func FromFile(path, sep string) (Params, error) {
    data, err := ReadFile(path);
    
    if err != nil {
        return Params{}, 
               fmt.Errorf("In FromFile: Failed to read file: '%s'!\nError: %w",
                           path, err);
    }
    
    return Params{par:path, sep:sep,
                  contents:str.Split(string(data[:]), "\n")}, nil;
}


func FromString(contents, sep string) Params {
    return Params{sep:sep, contents:str.Split(contents, "\n")};
}


func (self *Params) Par(name string) (string, error) {
    for _, line := range self.contents {
        if str.Contains(line, name) {
            return str.Trim(str.Split(line, self.sep)[1], " "), nil;
        }
    }
    
    return "", fmt.Errorf("In Par: Could not find parameter '%s' in %v", 
                           name, self.par);
}


func toInt(par string, idx int) (int, error) {
    ret, err := conv.Atoi(str.Split(par, " ")[idx]);
    
    if err != nil {
        return 0, 
               fmt.Errorf("In toInt: Could not convert string %s to float64!\nError: %w",
                           par, err);
    }
    
    return ret, nil;
}


func toFloat(par string, idx int) (float64, error) {
    ret, err := conv.ParseFloat(str.Split(par, " ")[idx], 64);
    
    if err != nil {
        return 0.0, 
               fmt.Errorf("Could not convert string %s to float64!\nError: %w",
                           par, err);
    }
    
    return ret, nil;
}

func (self *Params) Int(name string) (int, error) {
    data, err := self.Par(name);
    
    if err != nil {
        return 0, err;
    }
    
    return toInt(data, 0);
}

func (self *Params) Float(name string) (float64, error) {
    data, err := self.Par(name);
    
    if err != nil {
        return 0.0, err;
    }
    
    return toFloat(data, 0);
}


var tmp = Tmp{};


func TmpFile() (string, error) {
    file, err := io.TempFile("", "*");
    
    if err != nil {
        return "", 
               fmt.Errorf("In TmpFile: Failed to create a temporary file!\nError: %w", 
                           err);
    }
    
    defer file.Close()
    
    name := file.Name();
    
    tmp.files = append(tmp.files, name);
    
    return name, nil;
}


func RemoveTmp() error {
    for _, file := range tmp.files {
        if err := os.Remove(file); err != nil {
            continue;
            //return fmt.Errorf("Failed to remove temporary file '%s': %w", 
            //                   file, err);
        }
    }
    return nil;
}
