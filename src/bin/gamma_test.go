package main;

import (
    "fmt";
    test "testing";
    gm "../gamma";
);

const NUM = 10000000;
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


const (
    rng = 4300;
    azi = 3200;
    img_fmt = "FLOAT";
);


func TestGetParameter(t *test.T) {
    params := 
    fmt.Sprintf("range_samples: %d\nazimuth_lines: %d\nimage_format: %v", 
                rng, azi, img_fmt);
    
    pars := gm.FromString(params, ":");
    
    got := pars.Int("range_samples");
    
    if got != rng {
        t.Errorf("Expected %v for range_samples got %v", rng, got);
    }
    
    got = pars.Int("azimuth_lines");
    
    if got != azi {
        t.Errorf("Expected %v for azimuth_lines got %v", azi, got);
    }
    
    
    gots := pars.Par("image_format");
    
    if gots != img_fmt {
        t.Errorf("Expected %v for image_format got %v", img_fmt, gots);
    }
}

func TestS1Zip(t *test.T) {
    const (
        str = "/mnt/Dszekcso/ASC/S1A_IW_SLC__1SDV_20160702T163342_20160702T163409_011972_012763_24E2.zip";
        zip = "S1A_IW_SLC__1SDV_20160702T163342_20160702T163409_011972_012763_24E2.zip";
        mission = "S1A";
        dateStr = "20160702T163342_20160702T163409";
        mode = "IW";
        productType = "SLC";
        resolution = "_";
        level = "1";
        productClass = "S";
        pol = "DV";
        absoluteOrbit = "011972";
        DTID = "012763";
        UID = "24E2";
    );
    
    
    s1zip := gm.NewS1Zip(str);
    
    if s1zip.path != str {
        t.Errorf("Expected S1Zip.path to be '%s' got '%s'", str, s1zip.path);
    }
    
    
    if s1zip.zipBase != zip {
        t.Errorf("Expected S1Zip.zip to be '%s' got '%s'", zip, s1zip.zipBase);
    }
    
    if s1zip.mission != "S1A" {
    
    }
    
    if s1zip.dateStr != 
    
        path, zipBase, mission, dateStr, mode, productType, resolution string;
        level, productClass, pol, absoluteOrbit, DTID, UID string;
        date date;
    
}

