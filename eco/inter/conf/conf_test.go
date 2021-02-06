package conf

import (
	"fmt"
	"strings"
	"testing"
)

func TestFile(t *testing.T) {
	f := "conf/oneorange.yaml"
	fmt.Println(GenConfigurationFile(f))
	s := GenConfigurationFile(f)
	pathArr := strings.Split(s, "/")
	fmt.Println(pathArr)
}
