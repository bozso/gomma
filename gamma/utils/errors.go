package utils

import "fmt"

const (
    ParseIntErr Werror = "failed to parse '%s' into an integer"
    ParseFloatErr Werror = "failed to parse '%s' into an float"
    CmdErr Werror = "execution of command '%s' failed"
    ExeErr Werror = `Command '%s %s' failed!
    Output of command is: %v`

    FileOpenErr Werror = "failed to open file '%s'"
    FileReadErr Werror = "failed to read file '%s'"

    DirCreateErr Werror = "failed to create directory '%s'"
    FileExistErr Werror = "failed to determine wether '%s' exist"
    FileWriteErr Werror = "failed to write to file '%s'"
    FileCreateErr Werror = "failed to create file '%s'"
    MoveErr Werror = "failed to move '%s' to '%s'"
    EmptyStringErr Werror = "expected %s to be a non empty string"
)

type (
    Werror string
    CWerror string
)

func (e Werror) Wrap(err error, args ...interface{}) error {
    str := fmt.Sprintf(string(e), args...)
    return fmt.Errorf("%s\n%w", str, err)
}

func Wrap(err1, err2 error) error {
    return fmt.Errorf("%w\n%w", err1, err2)
}

func WrapFmt(err error, msg string, args ...interface{}) error {
    s := fmt.Sprintf(msg, args...)
    
    return fmt.Errorf("%s\n%w", s, err)
}

func (e Werror) Make(args ...interface{}) error {
    return fmt.Errorf(string(e), args...)
}

func (e CWerror) Wrap(err error) error {
    return fmt.Errorf("%s: %w", string(e), err)
}

func (e CWerror) Make() error {
    return fmt.Errorf(string(e))
}

type ErrorBase struct {
    err error
}

func (e ErrorBase) Unwrap() error {
    return e.err
}

type FileError struct {
    path, op string
    err error
}

func (e FileError) Error() string {
    return fmt.Sprintf("failed to %s file '%s'", e.op, e.path)
}

func (e FileError) Unwrap() error {
    return e.err
}

func OpenFail(path string, err error) FileError {
    return FileError{path, "open", err}
}

func CreateFail(path string, err error) FileError {
    return FileError{path, "create", err}
}

func ReadFail(path string, err error) FileError {
    return FileError{path, "read from", err}
}

func WriteFail(path string, err error) FileError {
    return FileError{path, "write to", err}
}

type EmptyStringError struct {
    variable string
    err      error
}

func (e EmptyStringError) Error() (s string) {
    s = "expected non empty string"
    
    if v := e.variable; len(v) > 0 {
        s = fmt.Sprintf("%s for '%s'", s, v)
    }
    
    return
}

func (e EmptyStringError) Unwrap() error {
    return e.err
}

type UnrecognizedMode struct {
    name, got string
    err error
}

func (e UnrecognizedMode) Error() string {
    return fmt.Sprintf("unrecognized mode '%s' for %s", e.got, e.name)
}

func (e UnrecognizedMode) Unwrap() error {
    return e.err
}

type ModeError struct {
    name string
    got fmt.Stringer
    err error
}

func (e ModeError) Error() string {
    return fmt.Sprintf("unrecognized mode '%s' for %s", e.got.String(), e.name)
}

func (e ModeError) Unwrap() error {
    return e.err
}

type OutOfBoundError struct {
    idx, length int
    err error
}

func (o OutOfBoundError) Error() string {
    return fmt.Sprintf("idx '%d' is out of bounds of length '%d'",
        o.idx, o.length)
}

func (o OutOfBoundError) Unwrap() error {
    return o.err
}

func IsOutOfBounds(idx, length int) error {
    if idx >= length {
        return OutOfBoundError{idx:idx, length:length}
    }
    return nil
}
