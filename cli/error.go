package cli

import (
	"fmt"
)

type InvalidInStreamArgument struct {
	Argument string
}

func (i InvalidInStreamArgument) Error() (s string) {
	return fmt.Sprintf(
		"expected either a valid filepath, no argument or '-' got '%s'",
		i.Argument)
}
