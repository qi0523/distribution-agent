package reference

import (
	"fmt"
	"testing"
)

func TestTagRegexp(t *testing.T) {
	strReg := NameRegexp.String()
	fmt.Println(strReg)
}
