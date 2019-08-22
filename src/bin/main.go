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
    fmt.Println(gm.ParseDate(gm.DateLong, "20121203T123112"));
}