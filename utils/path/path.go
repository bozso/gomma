package path

func Exist(s string) (b bool, err error) {
    b = false
    _, err = os.Stat(s)

    if err == nil {
        b = true
        return
    }
    
    if os.IsNotExist(err) {
        err = nil
        return
    }
    
    err = WrapFmt(err, "failed to check wether file '%s' exists", s)
    return
}
func Move(path string, dir string) (s string, err error) {
    dst, err := filepath.Abs(filepath.Join(dir, filepath.Base(path)))
    if err != nil {
        err = WrapFmt(err, "failed to create absolute path")
        return
    }
    
    if err = os.Rename(path, dst); err != nil {
        return
    }
    
    return dst, nil
}
