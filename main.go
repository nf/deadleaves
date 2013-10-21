/*
Copyright 2013 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Command deadleaves finds and prints the import paths of unused Go packages.
// A package is considered unused if it is not a command ("package main") and
// is not transitively imported by a command.
package main

import (
	"flag"
	"fmt"
	"go/build"
	"os"
	"path/filepath"
)

var stdFlag = flag.Bool("std", false, "report unused standard packages")

func main() {
	ctx := build.Default

	pkgs := make(map[string]*build.Package)
	for _, root := range ctx.SrcDirs() {
		err := filepath.Walk(root, func(path string, fi os.FileInfo, err error) error {
			if !fi.IsDir() {
				return nil
			}
			pkg, err := ctx.ImportDir(path, 0)
			if err != nil {
				return nil
			}
			pkgs[pkg.ImportPath] = pkg
			return nil
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "error walking %q: %v\n", root, err)
		}
	}

	used := make(map[string]bool)
	var recordDeps func(*build.Package)
	recordDeps = func(pkg *build.Package) {
		if used[pkg.ImportPath] {
			return
		}
		used[pkg.ImportPath] = true
		imports := append([]string{}, pkg.Imports...)
		imports = append(imports, pkg.TestImports...)
		for _, p := range imports {
			dep, ok := pkgs[p]
			if !ok {
				if p != "C" {
					fmt.Fprintf(os.Stderr, "package %q not found (imported by %q)\n", p, pkg.ImportPath)
				}
				continue
			}
			recordDeps(dep)
		}
	}
	for _, pkg := range pkgs {
		if pkg.Name == "main" {
			recordDeps(pkg)
		}
	}

	for path, pkg := range pkgs {
		if !used[path] {
			if pkg.Goroot && !*stdFlag {
				continue
			}
			fmt.Println(path)
		}
	}
}
