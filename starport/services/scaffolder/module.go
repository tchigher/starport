package scaffolder

import (
	"context"
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobuffalo/genny"
	"github.com/tendermint/starport/starport/pkg/cosmosver"
	"github.com/tendermint/starport/starport/pkg/gomodulepath"
	"github.com/tendermint/starport/starport/templates/module"
)

const (
	wasmImport = "github.com/CosmWasm/wasmd"
	apppkg     = "app"
	moduleDir  = "x"
)

// CreateModule creates a new empty module in the scaffolded app
func (s *Scaffolder) CreateModule(moduleName string) error {
	version, err := s.version()
	if err != nil {
		return err
	}
	// Check if the module already exist
	ok, err := moduleExists(s.path, moduleName)
	if err != nil {
		return err
	}
	if ok {
		return errors.New(fmt.Sprintf("The module %v already exists.", moduleName))
	}
	path, err := gomodulepath.ParseFile(s.path)
	if err != nil {
		return err
	}

	var (
		g    *genny.Generator
		opts = &module.CreateOptions{
			ModuleName: moduleName,
			ModulePath: path.RawPath,
		}
	)
	if version == cosmosver.Launchpad {
		g, err = module.NewCreateLaunchpad(opts)
	} else {
		g, err = module.NewCreateStargate(opts)
	}
	if err != nil {
		return err
	}
	run := genny.WetRunner(context.Background())
	run.With(g)
	return run.Run()
}

// ImportModule imports specified module with name to the scaffolded app.
func (s *Scaffolder) ImportModule(name string) error {
	version, err := s.version()
	if err != nil {
		return err
	}
	if version == cosmosver.Stargate {
		return errors.New("importing modules currently is not supported on Stargate")
	}
	ok, err := isWasmImported(s.path)
	if err != nil {
		return err
	}
	if ok {
		return errors.New("CosmWasm is already imported.")
	}
	path, err := gomodulepath.ParseFile(s.path)
	if err != nil {
		return err
	}
	g, err := module.NewImport(&module.ImportOptions{
		Feature: name,
		AppName: path.Package,
	})
	if err != nil {
		return err
	}
	run := genny.WetRunner(context.Background())
	run.With(g)
	return run.Run()
}

func moduleExists(appPath string, moduleName string) (bool, error) {
	abspath, err := filepath.Abs(filepath.Join(appPath, moduleDir, moduleName))
	if err != nil {
		return false, err
	}

	_, err = os.Stat(abspath)
	if err == nil {
		// The module already exists
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	// Error reading the directory
	return false, err
}

func isWasmImported(appPath string) (bool, error) {
	abspath, err := filepath.Abs(filepath.Join(appPath, apppkg))
	if err != nil {
		return false, err
	}
	fset := token.NewFileSet()
	all, err := parser.ParseDir(fset, abspath, func(os.FileInfo) bool { return true }, parser.ImportsOnly)
	if err != nil {
		return false, err
	}
	for _, pkg := range all {
		for _, f := range pkg.Files {
			for _, imp := range f.Imports {
				if strings.Contains(imp.Path.Value, wasmImport) {
					return true, nil
				}
			}
		}
	}
	return false, nil
}
