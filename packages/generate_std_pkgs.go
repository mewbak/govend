// +build ignore

// This file generates std_pkgs.go, which contains the standard library packages.

// This file has been modified from its original source:
// https://github.com/golang/tools/blob/master/imports/mkindex.go

package main

import (
	"bytes"
	"fmt"
	"go/build"
	"go/format"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
)

// pkgs is a map of the standard golang packages.
var pkgs = map[string][]stdpkg{
	"appengine": []stdpkg{
		stdpkg{path: "appengine"},
	},
	"cloudsql": []stdpkg{
		stdpkg{path: "appengine/cloudsql"},
	},
	"urlfetch": []stdpkg{
		stdpkg{path: "appengine/urlfetch"},
	},
}

// fset is a set of file tokens.
var fset = token.NewFileSet()

// stdpkg represents a standard package in "std_pkgs.go".
type stdpkg struct {
	path string // full pkg import path, e.g. "net/http"
	dir  string // absolute file path to pkg directory e.g. "/usr/lib/go/src/fmt"
}

func main() {

	// start with the default context.
	ctx := build.Default

	// remove the GOPATH, we only want to search packages in the GOROOT.
	ctx.GOPATH = ""

	// iterate through the list of package source root directories.
	for _, path := range ctx.SrcDirs() {

		// open the file
		f, err := os.Open(path)
		if err != nil {
			log.Print(err)
			continue
		}

		// gather all the child names from the directory in a single slice.
		children, err := f.Readdir(-1)

		// close the file and check for errors.
		f.Close()
		if err != nil {
			log.Print(err)
			continue
		}

		// iterate through each child name.
		for _, child := range children {

			// check the child name is a directory.
			if child.IsDir() {

				// load the package path and name.
				load(path, child.Name())
			}
		}
	}

	// start with a byte buffer.
	var buf bytes.Buffer

	// write preliminary data to the byte buffer.
	buf.WriteString("// AUTO-GENERATED BY generate_std_pkgs.go\n\npackage packages\n")
	fmt.Fprintf(&buf, "var stdpkgs = %#v\n", pkgs)

	buf.WriteString("type stdpkg struct { path, dir string }")

	// transfer buffer bytes to final source.
	src := buf.Bytes()

	// replace main.pkg type name with pkg.
	src = bytes.Replace(src, []byte("main.stdpkg"), []byte("stdpkg"), -1)

	// replace actual GOROOT with "/go".
	src = bytes.Replace(src, []byte(ctx.GOROOT), []byte("/go"), -1)

	// add line wrapping and some better formatting.
	src = bytes.Replace(src, []byte("[]stdpkg{"), []byte("[]stdpkg{\n"), -1)
	src = bytes.Replace(src, []byte(", stdpkg"), []byte(",\nstdpkg"), -1)
	src = bytes.Replace(src, []byte("}}, "), []byte("},\n},\n"), -1)
	src = bytes.Replace(src, []byte("true, "), []byte("true,\n"), -1)
	src = bytes.Replace(src, []byte("}}}"), []byte("},\n},\n}"), -1)

	// format the source bytes.
	var err error
	src, err = format.Source(src)
	if err != nil {
		log.Fatal(err)
	}

	// write source bytes to file.
	if err := ioutil.WriteFile("std_pkgs.go", src, 0644); err != nil {
		log.Fatal(err)
	}
}

// load takes a path root and import path
func load(root, importpath string) {

	// get the package name
	name := path.Base(importpath)
	if name == "testdata" {
		return
	}

	// calculate the package source directory
	dir := filepath.Join(root, importpath)

	// append the package values to the package map.
	pkgs[name] = append(pkgs[name], stdpkg{
		path: importpath,
		dir:  dir,
	})

	// get the package directory
	pkgDir, err := os.Open(dir)
	if err != nil {
		return
	}

	// gather all the child names from the directory in a single slice.
	children, err := pkgDir.Readdir(-1)

	// close the file and check for errors.
	pkgDir.Close()
	if err != nil {
		return
	}

	// iterate through each child name.
	for _, child := range children {

		// get the child's name.
		name := child.Name()

		// check that the childs names not blank
		if name == "" {
			continue
		}

		// handle special package name cases.
		if c := name[0]; c == '.' || ('0' <= c && c <= '9') {
			continue
		}

		// check if the child name is a directory.
		if child.IsDir() {

			// load the package path and name.
			load(root, filepath.Join(importpath, name))
		}
	}
}