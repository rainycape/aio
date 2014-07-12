package aio

import (
	"fmt"
	"os"
)

func Fd(file interface{}) (int, error) {
	if f, ok := file.(*os.File); ok {
		return int(f.Fd()), nil
	}
	return 0, fmt.Errorf("can't obtain fd from %T", file)
}
