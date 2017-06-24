package indeps

import (
	"golang.org/x/tools/go/loader"
)

// AnalyzePackagePath analyzes the package with the given import path, if
// such a package exists.
func AnalyzePackagePath(path string) (*Graph, error) {
	var conf loader.Config
	conf.Import(path)
	prog, err := conf.Load()
	if err != nil {
		return nil, err
	}

	pkg := prog.Package(path)

	return analyze(pkg)
}
