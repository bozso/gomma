package gamma

import (
    "io"
    "os"
    "log"
    "fmt"
    "archive/zip"
    "path/filepath"
    "regexp"
)

type Extractor struct {
    pol, dst string
    templates templates
    *zip.ReadCloser
}

func (s1 *S1Zip) newExtractor(dst string) (ex Extractor, err error) {
    path := s1.Path
    
    ex.templates       = s1.Templates
    ex.pol             = s1.pol
    ex.dst             = dst
    
    if ex.ReadCloser, err = zip.OpenReader(path); err != nil {
        err = Handle(err, "failed to open zipfile '%s'", path)
        return
    }
    
    return ex, nil
}

//func (ex Extractor) Wrap() error {
    //return fmt.Errorf("failure during extraction: %w", ex.err)
//}

func (ex *Extractor) extract(mode tplType, iw int) (s string, err error) {
    var tpl string
    
    if fmtNeeded[mode] {
        tpl = fmt.Sprintf(ex.templates[mode], iw, ex.pol)
    } else {
        tpl = ex.templates[mode]
    }
    
    s, err = extract(ex.ReadCloser, tpl, ex.dst)
    
    return
}


func extractFile(src *zip.File, dst string) error {
    srcName := src.Name

    in, err := src.Open()
    if err != nil {
        return Handle(err, "failed to open file '%s'", srcName)
    }
    defer in.Close()
    
    dir := filepath.Dir(dst)
    if err = os.MkdirAll(dir, os.ModePerm); err != nil {
        return DirCreateErr.Wrap(err, dir)
    }
    
    var out *os.File
    if out, err = os.Create(dst); err != nil {
        return FileCreateErr.Wrap(err, dst)
    }
    defer out.Close()
    
    log.Printf("Extracting '%s' into '%s'", srcName, dst)
    
    if _, err = io.Copy(out, in); err != nil {
        return Handle(err, "failed to copy contents of '%s' into '%s'",
            srcName, dst)
    }

    return nil
}

func extract(file *zip.ReadCloser, template, dst string) (s string, err error) {
    //log.Fatalf("%s %s", root, template)
    
    var matched, exist bool
    
    // go through files in the zipfile
    for _, zipfile := range file.File {
        name := zipfile.Name
        
        if matched, err = regexp.MatchString(name, template); err != nil {
            err = Handle(err,
                "failed to check whether zipped file '%s' matches templates",
                name)
            return
        }
        
        if !matched {
            continue
        }
        
        s = filepath.Join(dst, name)
        
        //fmt.Printf("Matched: %s\n", dst)
        //fmt.Printf("\n\nCurrent: %s\nTemplate: %s\nMatched: %v\n",
        //    name, template, matched)
        
        if exist, err = Exist(s); err != nil {
            err = Handle(err, "stat failed on file '%s'", name)
            return
        }
        
        if !exist {
            if err = extractFile(zipfile, s); err != nil {
                err = Handle(err, "failed to extract file '%s'", name)
                return
            }
        }
        return s, nil
    }
    return s, nil
}

type ExtractError struct {
    file, path string
    err error
}

func (e ExtractError) Error() string {
    return fmt.Sprintf("failed to extract %s file from '%s'",
        e.file, e.path)
}

func (e ExtractError) Unwrap() error {
    return e.err
}
