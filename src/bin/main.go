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
    
    gm.MakeDefaultConfig("gamma.json");
    
    _, err := gm.FromFile("asd", ":");
    
    if err != nil {
        gm.Fatal(err, "A");
    }
    
    defer gm.RemoveTmp();
    fmt.Println(gm.First());
}