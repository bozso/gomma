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


func main() {
    var a string;
    
    fmt.Println(a);
    fmt.Println(gm.First());
}