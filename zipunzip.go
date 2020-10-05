package main

import (
	"archive/zip"
	"expvar"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var BenchStatus = expvar.NewMap("bench_status").Init()
var destDir = flag.String("destdir", "dest", "Destination directory")
var finalDir = flag.String("finaldir", "dest/out/", "Final zip directory.")
var inDir = flag.String("indir", "testdata", "directory for Input zip files")
var files []string

const ZipTime = "Compression Time"
const UnzipTime = "Decompression Time"

func init() {
	BenchStatus.Set(ZipTime, new(expvar.Int))
	BenchStatus.Set(UnzipTime, new(expvar.Int))
}

func Unzip(indir, dest string) {
	matches, err := filepath.Glob(filepath.Join(indir, "*.zip"))
	if err != nil {
		panic(err)
	}
	for _, inFile := range matches {

		r, err := zip.OpenReader(inFile)
		if err != nil {
			panic(err)
		}
		for _, f := range r.File {
			fName := string(f.Name)
			files = append(files, f.Name)
			if strings.HasSuffix(string(f.Name), "/") == true {
				os.MkdirAll(filepath.Join(dest, fName), os.ModeDir|0755)
				continue
			}

			rc, err := f.Open()
			if err != nil {
				panic(err)
			}

			out, err := os.OpenFile(filepath.Join(dest, fName), os.O_RDWR|os.O_CREATE, 0644)

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
		r.Close()
	}
}

func Rezip(dir, output string) {
	// Create a buffer to write our archive to.
	fp, err := os.Create(output)
	if err != nil {
		panic(err)
	}
	// Create a new zip archive.
	w := zip.NewWriter(fp)

	// Add some files to the archive.
	for _, file := range files {
		f, err := w.Create(file)
		if err != nil {
			log.Fatal(err)
		}
		if strings.HasSuffix(file, "/") == true {
			continue
		}

		in, err := os.Open(filepath.Join(dir, file))
		if err != nil {
			fmt.Printf("WHILE OPENNING %s", file)
			panic(err)
		}
		_, err = io.Copy(f, in)
		if err != nil {
			log.Fatal(err)
		}
		err = in.Close()
		if err != nil {
			panic(err)
		}
	}
	// Make sure to check the error on Close.
	err = w.Close()
	if err != nil {
		log.Fatal(err)
	}

	err = fp.Close()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Rezip complete.")
}

func main() {
	flag.Parse()

	if err := os.MkdirAll(*destDir, 0755); err != nil {
		panic("Cannot create dest dir")
	}
	if err := os.MkdirAll(*finalDir, 0755); err != nil {
		panic("Cannot create final dir")
	}
	start := time.Now().Unix()
	fmt.Printf("Started: input=%s, tempdir=%s\n", *inDir, *destDir)
	Unzip(*inDir, *destDir)
	elapsed := time.Now().Unix() - start
	BenchStatus.Get(UnzipTime).(*expvar.Int).Set(int64(elapsed))
	fmt.Printf("Decompression done.\n")
	start = time.Now().Unix()
	Rezip(*destDir, filepath.Join(*finalDir, "test.zip"))
	BenchStatus.Get(ZipTime).(*expvar.Int).Set(int64(elapsed))
	fmt.Printf("Recompression complete.\n")
}
