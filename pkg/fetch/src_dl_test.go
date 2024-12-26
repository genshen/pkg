package fetch

import "testing"

func TestIntAbs(t *testing.T) {
	got := archivePackageFilepath("a/b/c", "foo/bar", "tar.gz")
	if got != "a/b/c/foo_bar.tar.gz" {
		t.Errorf("unexpetced path %s", got)
	}
}
