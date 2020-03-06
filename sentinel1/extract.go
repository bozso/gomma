package sentinel1

import (
    "io"
    "log"
    "fmt"
    "archive/zip"
    "path/filepath"
    "regexp"

    "github.com/bozso/gamma/utils"
    "github.com/bozso/gamma/utils/path"
    "github.com/bozso/gamma/utils/stream"
)

type Extractor struct {
    pol, dst string
    path *string
    templates *templates
    *zip.ReadCloser
    err error
}

func (s1 *Zip) newExtractor(dst string) (ex Extractor) {
    ex.path      = &s1.Path
    ex.templates = &s1.Templates
    ex.pol       = s1.pol
    ex.dst       = dst
    ex.ReadCloser, ex.err = zip.OpenReader(*ex.path)

    return
}

func (ex Extractor) Wrap() error {
    if ex.err == nil {
        return nil
    }
    
    return fmt.Errorf(
        "failure during the extraction from zipfile '%s': %w", *ex.path,
        ex.err)
}

func (ex *Extractor) Extract(mode tplType, iw int) (s string) {
    if ex.err != nil {
        return
    }
    
    var tpl string
    
    if fmtNeeded[mode] {
        tpl = fmt.Sprintf(ex.templates[mode], iw, ex.pol)
    } else {
        tpl = ex.templates[mode]
    }
    
    s, ex.err = ex.extract(tpl, ex.dst)
    
    return
}

func (ex Extractor) extract(template, dst string) (s string, err error) {
    //log.Fatalf("%s %s", root, template)
    var matched, exist bool
    
    // go through files in the zipfile
    for _, zipfile := range ex.ReadCloser.File {
        name := zipfile.Name
        
        if matched, err = regexp.MatchString(name, template); err != nil {
            err = utils.WrapFmt(err,
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
        
        if exist, err = path.Exist(s); err != nil {
            err = utils.WrapFmt(err, "stat failed on file '%s'", name)
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

func extractFile(src *zip.File, dst string) (err error) {
    srcName := src.Name
    
    in, err := stream.Open(srcName)
    
    if err != nil { return }
    defer in.Close()
    
    dir := filepath.Dir(dst)
    if err = path.Mkdir(dir); err != nil {
        return
    }
    
    out, err := stream.Create(dst)
    if err != nil { return }
    defer out.Close()
    
    log.Printf("Extracting '%s' into '%s'", srcName, dst)
    
    if _, err = io.Copy(out, in); err != nil {
        return utils.WrapFmt(err,
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
