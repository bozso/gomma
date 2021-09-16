package cli

import (
	"errors"
	"io"
	"io/fs"

	sfs "git.sr.ht/~istvan_bozso/shutil/fs"
)

var os = sfs.OS()

func IsFile(fsys fs.FS, s string) (b bool, err error) {
	stat, err := fs.Stat(fsys, s)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}

	if stat.IsDir() {
		return false, nil
	}

	return
}

func Reader(s string) (r io.Reader, err error) {
	return
}
