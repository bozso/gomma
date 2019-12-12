package gamma;

import (
    "fmt"
    "testing"
    "os"
    "io/ioutil"
)

const NUM = 10000000
const NUMErr = 10000
const buf = 1000

var mainErr = NewModuleErr("test")


func testError1() error {
    var ferr = mainErr("testError1")
    
    if err := testError2(); err != nil {
        return ferr.Wrap(err, "failed to load some file")
    }
    
    return nil
}


func testError2() error {
    var ferr = mainErr("testError2")
    
    file, err := os.Open("asd")
    if err != nil {
        return ferr.Wrap(err, "failed to open file")
    }
    defer file.Close()
    
    return nil
}

func testError1Vanilla() error {
    if err := testError2(); err != nil {
        return err
    }
    
    return nil
}


func testError2Vanilla() error {
    
    file, err := os.Open("asd")
    if err != nil {
        return err
    }
    defer file.Close()
    
    return nil
}

func BenchmarkVanillaError(b *testing.B) {
    b.ReportAllocs()
    
    for ii := 0; ii < NUMErr; ii++ {
        if err := testError1Vanilla(); err != nil {
            fmt.Fprintf(ioutil.Discard, "Error occurred: %s\n", err)
        }
    }
}

func BenchmarkCustomError(b *testing.B) {
    b.ReportAllocs()
    
    for ii := 0; ii < NUMErr; ii++ {
        if err := testError1(); err != nil {
            fmt.Fprintf(ioutil.Discard, "Error occurred: %s\n", err)
        }
    }
}

func BenchmarkAppend(b *testing.B) {
    b.ReportAllocs();
    
    ints := make([]int, NUM);
    out := make([]int, buf);
    
    for _, val := range ints {
        if val % 2 == 0 {
            out = append(out, val);
        }
    }
}

func BenchmarkPrealloc(b *testing.B) {
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


func TestGetParameter(t *testing.T) {
    params := fmt.Sprintf(
        "range_samples: %d\nazimuth_lines: %d\nimage_format: %v",
        rng, azi, img_fmt);
    
    pars := FromString(params, ":");
    
    got, err := pars.Int("range_samples", 0);
    
    if err != nil {
        t.Errorf("Failed to parse %s!", params)
    }
    
    if got != rng {
        t.Errorf("Expected %v for range_samples got %v", rng, got);
    }
    
    got, err = pars.Int("azimuth_lines", 0);
    
    if err != nil {
        t.Errorf("Failed to parse %s!", params)
    }
    
    if got != azi {
        t.Errorf("Expected %v for azimuth_lines got %v", azi, got);
    }
    
    
    gots, err := pars.Param("image_format");
    if err != nil {
        t.Errorf("Failed to parse %s!", params)
    }

    if gots != img_fmt {
        t.Errorf("Expected %v for image_format got %v", img_fmt, gots);
    }
}


func TestPoints(t *testing.T) {
    rect := Rect{Max: Point{X:1.0, Y:2.0},
                 Min: Point{X:0.0, Y:-1.0}}
    
    point1, point2 :=  Point{X:0.5, Y: -0.8}, Point{X:2.0, Y: -0.8}
    
    if !point1.InRect(&rect) {
        t.Errorf("point1 (%v) should be in rectangle (%v)", point1, rect)
    }
    
    if point2.InRect(&rect) {
        t.Errorf("point2 (%v) should not be in rectangle (%v)", point2, rect)
    }
}


/*
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
*/
