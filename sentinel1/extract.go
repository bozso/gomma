package sentinel1

import (
    "io"
    "log"
    "fmt"
    "archive/zip"
    "path/filepath"
    "regexp"

    "github.com/bozso/gotoolbox/path"
    "github.com/bozso/gotoolbox/errors"
    
    "github.com/bozso/gomma/common"
)

type Extractor struct {
    pol common.Pol
    dst path.Dir
    path path.ValidFile
    templates
    *zip.ReadCloser
    err error
}

func (s1 Zip) newExtractor(dst path.Dir) (ex Extractor) {
    ex.path      = s1.Path
    ex.templates = s1.Templates
    ex.pol       = s1.pol
    ex.dst       = dst
    ex.ReadCloser, ex.err = zip.OpenReader(ex.path.GetPath())

    return
}

func (ex Extractor) Err() (err error) {
    err = ex.err
    if err != nil {
        err = fmt.Errorf(
            "failure during the extraction from zipfile '%s': %w",
            ex.path, err)
    }
    
    return err
}

func (ex *Extractor) Extract(mode tplType, iw int) (s string) {
    if ex.err != nil {
        return
    }
    
    tpl := ex.templates[mode].Render(iw, ex.pol)
    
    s, ex.err = ex.extract(tpl, ex.dst)
    
    return
}

func (ex Extractor) extract(template string, dst path.Dir) (s string, err error) {
    //log.Fatalf("%s %s", root, template)
    var matched, exist bool
    
    // go through files in the zipfile
    for _, zipfile := range ex.ReadCloser.File {
        name := zipfile.Name
        
        if matched, err = regexp.MatchString(name, template); err != nil {
            err = errors.WrapFmt(err,
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
        
        if exist, err = s.Exist(); err != nil {
            return
        }
        
        if !exist {
            if err = extractFile(zipfile, s); err != nil {
                err = ExtractError{name, err}
                return
            }
        }
        return s, nil
    }
    return s, nil
}

func extractFile(src *zip.File, dst path.Path) (err error) {
    in, err := src.Open()
    if err != nil {
        return
    }
    defer in.Close()
    
    dir := dst.Dir(dst)
    if err = dir.Make(); err != nil {
        return
    }
    
    out, err := dst.Create()
    if err != nil {
        return
    }
    defer out.Close()
    
    log.Printf("Extracting '%s' into '%s'", srcName, dst)
    
    if _, err = io.Copy(out, in); err != nil {
        return errors.WrapFmt(err,
            "failed to copy contents of '%s' into '%s'", srcName, dst)
    }

    return nil
}

type ExtractError struct {
    path string
    err error
}

func (e ExtractError) Error() string {
    return fmt.Sprintf("failed to extract file '%s'", e.path)
}

func (e ExtractError) Unwrap() error {
    return e.err
}
