package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var destDir = flag.String("destdir", "dest", "Destination directory")
var inDir = flag.String("indir", "testdata", "directory for Input zip files")
var files []string

func main() {
	flag.Parse()
	// Open a zip archive for reading.
	matches, err := filepath.Glob(filepath.Join(*inDir, "*.zip"))
	if err != nil {
		panic(err)
	}
	for _, inFile := range matches {
		r, err := zip.OpenReader(inFile)
		if err != nil {
			log.Fatal(err)
		}
		defer r.Close()
		if err := os.MkdirAll(*destDir, 0755); err != nil {
			panic("Cannot create dest dir")
		}
		os.Chdir(*destDir)
		// Iterate through the files in the archive,
		// printing some of their contents.
		for _, f := range r.File {
			fName := string(f.Name)
			files = append(files, f.Name)
			if strings.HasSuffix(string(f.Name), "/") == true {
				fmt.Printf("%s\n", fName)
				os.MkdirAll(fName, os.ModeDir|0755)
				continue
			}

			rc, err := f.Open()
			if err != nil {
				panic(err)
			}

			out, err := os.OpenFile(fName, os.O_RDWR|os.O_CREATE, 0644)

			if err != nil {
				panic(err)
			}
			_, err = io.Copy(out, rc)
			if err != nil {
				panic(err)
			}
			rc.Close()
			out.Close()
		}
	}
	bigzip(*destDir, "test.zip")
}

func bigzip(dir, output string) {
	// Create a buffer to write our archive to.
	fp, err := os.Create(output)
	if err != nil {
		panic(err)
	}
	// Create a new zip archive.
	w := zip.NewWriter(fp)

	// Add some files to the archive.
	for _, file := range files {
		fmt.Printf("handlng %s\n", file)
		f, err := w.Create(file)
		if err != nil {
			log.Fatal(err)
		}
		if strings.HasSuffix(file, "/") == true {
			continue
		}

		in, err := os.Open(file)
		if err != nil {
			fmt.Printf("WHILE OPENNING %s", file)
			panic(err)
		}
		_, err = io.Copy(f, in)
		if err != nil {
			log.Fatal(err)
		}
		in.Close()
	}

	// Make sure to check the error on Close.
	err = w.Close()
	if err != nil {
		log.Fatal(err)
	}
}
