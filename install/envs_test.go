package install

import (
	"fmt"
	"runtime"
	"testing"
)

func Test_ProcessEnv(t *testing.T) {
	s := `1{{.CORES}}`
	newstr := processEnv(s)
	ncpus := runtime.NumCPU()
	if newstr != fmt.Sprintf("1%v", ncpus) {
		t.Errorf("test failed.%s - %s", s, newstr)
	}
}
