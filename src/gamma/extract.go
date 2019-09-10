package gamma

import (
	"io"
	"os"
	rex "regexp"
    zip "archive/zip"
	fp "path/filepath"
)

type (
    extractInfo struct {
        pol, iw, root string
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

	out, err := os.Create(dst)
	if err != nil {
		return handle(err, "Could not create file '%s'!", dst)
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return handle(err, "Could not copy contents of '%s' into '%s'!",
			srcName, dst)
	}

	return nil
}

func matches(candidate string, templates []string) (bool, error) {
    handle := Handler("matches")
    
    for _, tpl := range templates {
        matched, err := rex.MatchString(tpl, candidate)
        if err != nil {
            return false, handle(err, "rex.MatchString failed!")
        }
        
        if matched {
            return true, nil
        }
    }

    return false, nil
}


func extract(path, root string, templates []string) ([]string, error) {
	handle := Handler("extract")

	file, err := zip.OpenReader(path)

	if err != nil {
		return nil, handle(err, "Could not open zipfile: '%s'!", path)
	}

	defer file.Close()

	ret := make([]string, BufSize)

	// go through files in the zipfile
	for _, zipfile := range file.File {
		srcName := zipfile.Name
		dst := fp.Join(root, srcName)

        name := zipfile.Name
        matched, err := matches(name, templates)
        
        if err != nil {
            return nil, handle(err,
                "Failed to check wether zipped file '%s' matches templates!",
                name)
        }
        
        if matched {
            ret = append(ret, dst)
            _, err := os.Stat(dst);
            
            if err != nil {
                return nil, handle(err, "Stat failed on file : '%s'!", file)
            }
            
            if os.IsNotExist(err) {
                err := extractFile(zipfile, dst)

                if err != nil {
                    return nil, handle(err,
                        "Failed to extract file : '%s' from zip '%v'!",
                        srcName, file)
                }
            }
        }
	}
	
    return ret, nil
}

func (self *extractInfo) filterFiles(templates []string) ([]string, error) {
    handle := Handler("extractInfo.filterFiles")
    ret := make([]string, BufSize)
    
    for _, file := range self.Extracted {
        matched, err := matches(file, templates)
        
        if err != nil {
            return nil, handle(err,
                "Failed to check wether extracted file '%s' matches templates!",
                file)
        }
        
        if matched {
            ret = append(ret, file)
        }
    }
    return ret, nil
}
