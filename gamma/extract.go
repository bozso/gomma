package gamma

import (
	"io"
	"os"
    "log"
    "fmt"
    zip "archive/zip"
	fp "path/filepath"
	rex "regexp"
)

type (
	ExtractOpt struct {
		pol, root string
	}
    
    S1Extractor struct {
        ExtractOpt
		templates templates
        zip 	  *zip.ReadCloser
    }
)

func extractFile(src *zip.File, dst string) error {
	handle := Handler("extractFile")
    
	srcName := src.Name

	in, err := src.Open()
	if err != nil {
		return handle(err, "Could not open file '%s'!", srcName)
	}
	defer in.Close()
    
    dir := fp.Dir(dst)
    err = os.MkdirAll(dir, os.ModePerm)
    
    if err != nil {
        return handle(err, "Failed to create directory: %s!", dir)
    }
    
	out, err := os.Create(dst)
	if err != nil {
		return handle(err, "Could not create file '%s'!", dst)
	}
	defer out.Close()
    
    log.Printf("Extracting '%s' into '%s'", srcName, dst)
	_, err = io.Copy(out, in)
	if err != nil {
		return handle(err, "Could not copy contents of '%s' into '%s'!",
			srcName, dst)
	}

	return nil
}

func matches(candidate string, template string) (bool, error) {
	handle := Handler("matches")

    matched, err := rex.MatchString(template, candidate)
    if err != nil {
        return false, handle(err, "MatchString failed!")
    }

	return matched, nil
}

func extract(file *zip.ReadCloser, template, root string) (ret string, err error) {
	handle := Handler("extract")
    
    //log.Fatalf("%s %s", root, template)
    
    var matched, exist bool
	// go through files in the zipfile
	for _, zipfile := range file.File {
		name := zipfile.Name
        
		matched, err = matches(name, template)
        
		if err != nil {
			err = handle(err,
				"Failed to check whether zipped file '%s' matches templates!",
				name)
            return
		}
		
        ret = fp.Join(root, name)
        
		if !matched {
            continue
		}
        
        //fmt.Printf("Matched: %s\n", dst)
        //fmt.Printf("\n\nCurrent: %s\nTemplate: %s\nMatched: %v\n",
        //    name, template, matched)
        
        exist, err = Exist(ret)
        
        if err != nil {
            err = handle(err, "Stat failed on file : '%s'!", name)
            return
        }
        
        if !exist {
            err = extractFile(zipfile, ret)

            if err != nil {
                err = handle(err, "Failed to extract file : '%s'!", name)
                return
            }
        }
        return ret, nil
	}
    return "", nil
}

func (self *S1Zip) newExtractor(ext *ExtractOpt) (ret S1Extractor, err error) {
	handle := Handler("S1Zip.newExtractor")
    path := self.Path
    
    ret.templates = self.Templates
    ret.pol       = ext.pol
    ret.root      = ext.root
    ret.zip, err  = zip.OpenReader(path)

    if err != nil {
        err = handle(err, "Could not open zipfile: '%s'!", path)
		return
    }
    
    return ret, nil
}

func (self *S1Extractor) extract(mode tplType, iw int) (string, error) {
    handle := Handler("S1Extractor.extract")
	
    var tpl string
    
    if fmtNeeded[mode] {
        tpl = fmt.Sprintf(self.templates[mode], iw, self.pol)
    } else {
        tpl = self.templates[mode]
    }
    
    
    ret, err := extract(self.zip, tpl, self.root)
    
    if err != nil {
        return "", handle(err, "Error occurred while extracting!")
    }
    
    return ret, nil
}

func (self *S1Extractor) Close() {
    self.zip.Close()
}
