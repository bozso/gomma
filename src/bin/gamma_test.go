package main;

import (
    test "testing";
);

const NUM = 1000000;
const buf = 1000;


func BenchmarkAppend(b *test.B) {
    b.ReportAllocs();
    
    ints := make([]int, NUM);
    out := make([]int, buf);
    
    for _, val := range ints {
        if val % 2 == 0 {
            out = append(out, val);
        }
    }
}

func BenchmarkPrealloc(b *test.B) {
    b.ReportAllocs();
    
    ints := make([]int, NUM);
    idx := make([]bool, len(ints))
    num := 0;
    
    for ii, val := range ints {
        if val % 2 == 0 {
            idx[ii] = true;
            num++;
        }
    }
    
    out := make([]int, len(ints));
    
    for ii, val := range ints {
        jj := 0
        if idx[ii] {
            out[jj] = val;
            jj++;
        }
    }
}
