package main;

import (
    "fmt";
    gm "../gamma";
);


// custom string method
type Al struct {
    a, b int
}


func (self Al) String() string {
    return fmt.Sprintf("%d %d", self.a, self.b);
}

var gamma_cmds = map[string]gm.CmdFun{
    "eog": gm.MakeCmd("eog"),
};

func main() {
    gamma_cmds["eog"]("a", 1, 2.0, "a");
}