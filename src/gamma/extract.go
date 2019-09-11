package gamma

import (
	"io"
	"os"
    "log"
    //"fmt"
    zip "archive/zip"
	fp "path/filepath"
	rex "regexp"
)

type (
	extractInfo struct {
		pol, root string
		Extracted     []string
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

func (self *extractInfo) extract(file *zip.ReadCloser, template string) (string, error) {
	handle := Handler("extractInfo.extract")
    
    
	// go through files in the zipfile
	for _, zipfile := range file.File {
		name := zipfile.Name
		dst := fp.Join(self.root, name)
        
		matched, err := matches(name, template)
        
        
		if err != nil {
			return "", handle(err,
				"Failed to check whether zipped file '%s' matches templates!",
				name)
		}
        
        //fmt.Printf("\n\nCurrent: %s\nTemplate: %s\nMatched: %v\n",
        //    name, template, matched)
        
		if !matched {
            continue
		}
        
        
        // TODO: clean up program logic
        _, err = os.Stat(dst)
        
        if err != nil && os.IsNotExist(err) {
            err = extractFile(zipfile, dst)

            if err != nil {
                return "", handle(err, "Failed to extract file : '%s'!", name)
            }
            
            return dst, nil
        } else {
            return dst, nil
        }
        
        if err != nil {
            return "", handle(err, "Stat failed on file : '%s'!", name)
        }
	}
    return "", nil
}
