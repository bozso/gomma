package sentinel1

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"regexp"

	"github.com/bozso/gotoolbox/errors"
	"github.com/bozso/gotoolbox/path"

	"github.com/bozso/gomma/common"
)

type Extractor struct {
	pol  common.Pol
	dst  path.Dir
	path path.ValidFile
	templates
	*zip.ReadCloser
	err error
}

func (s1 Zip) newExtractor(dst path.Dir) (ex Extractor) {
	ex.path = s1.Path
	ex.templates = s1.Templates
	ex.pol = s1.pol
	ex.dst = dst
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

func (ex *Extractor) Extract(mode tplType, iw int) (vf path.ValidFile) {
	if ex.err != nil {
		return
	}

	tpl := ex.templates[mode].Render(iw, ex.pol)
	vf = ex.extract(tpl)
	return
}

func (ex *Extractor) extract(template string) (vf path.ValidFile) {
	if ex.err != nil {
		return
	}

	//log.Fatalf("%s %s", root, template)
	dst := ex.dst

	// go through files in the zipfile
	for _, zipfile := range ex.ReadCloser.File {
		name := zipfile.Name

		matched, err := regexp.MatchString(name, template)
		if err != nil {
			err = errors.WrapFmt(err,
				"failed to check whether zipped file '%s' matches templates",
				name)
			ex.err = err
			return
		}

		if !matched {
			continue
		}

		outFile := dst.Join(name).ToFile()

		//fmt.Printf("Matched: %s\n", dst)
		//fmt.Printf("\n\nCurrent: %s\nTemplate: %s\nMatched: %v\n",
		//    name, template, matched)

		exist, err := outFile.Exist()
		if err != nil {
			ex.err = err
			return
		}

		if exist {
			vf, ex.err = outFile.ToValid()
			return
		}

		if err := extractFile(zipfile, outFile); err != nil {
			ex.err = ExtractError{name, err}
			return
		}

		return
	}
	return
}

func extractFile(src *zip.File, dst path.File) (err error) {
	srcName := src.Name

	in, err := src.Open()
	if err != nil {
		return
	}
	defer in.Close()

	dir := dst.Dir()
	if _, err = dir.Mkdir(); err != nil {
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
	err  error
}

func (e ExtractError) Error() string {
	return fmt.Sprintf("failed to extract file '%s'", e.path)
}

func (e ExtractError) Unwrap() error {
	return e.err
}
