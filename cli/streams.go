package cli

import (
	"errors"
	"io"
	"io/fs"
	"os"

	sfs "git.sr.ht/~istvan_bozso/shutil/fs"
)

var osFS = sfs.OS()

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

func IsDefaultStreamValue(s string) (b bool) {
	return len(s) == 0 || s == "-"
}

func Reader(s string) (r io.ReadCloser, err error) {
	if IsDefaultStreamValue(s) {
		r = os.Stdin

		return
	}

	isFile, err := IsFile(osFS, s)
	if err != nil {
		return
	}

	if !isFile {
		err = &InvalidInStreamArgument{
			Argument: s,
		}

		return
	}

	r, err = osFS.Open(s)
	return
}

func Writer(s string) (w io.WriteCloser, err error) {
	if IsDefaultStreamValue(s) {
		w = os.Stdout

		return
	}

	w, err = sfs.Create(osFS, s)
	return
}
