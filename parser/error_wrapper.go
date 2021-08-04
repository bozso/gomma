package parser

type ErrorWrapper interface {
	WrapSplitError(line string) error
	WrapParseError(string, Mode, error) error
}
