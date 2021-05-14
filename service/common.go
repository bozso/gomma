package service

import (
	"bufio"
	"io"
	"log"

	"github.com/bozso/gotoolbox/cli/stream"
	"github.com/bozso/gotoolbox/path"

	s1 "github.com/bozso/gomma/sentinel1"
)

type Empty struct{}

type Output struct {
	Out stream.Out `json:"output"`
}

type Input struct {
	In stream.In `json:"input"`
}

func parseS1(zip path.ValidFile, dst path.Dir) (S1 *s1.Zip, IWs s1.IWInfos, err error) {
	if S1, err = s1.NewZip(zip); err != nil {
		return
	}

	log.Printf("Parsing IW Information for S1 zipfile '%s'", S1.Path)

	if IWs, err = S1.Info(dst); err != nil {
		return
	}

	return
}

func loadS1(reader io.Reader) (S1 s1.Zips, err error) {
	file := bufio.NewScanner(reader)

	for file.Scan() {
		if err = file.Err(); err != nil {
			return
		}

		vf, Err := path.New(file.Text()).ToValidFile()
		if Err != nil {
			err = Err
			return
		}

		s1zip, Err := s1.NewZip(vf)
		if Err != nil {
			err = Err
			return
		}

		S1 = append(S1, s1zip)
	}

	return
}
