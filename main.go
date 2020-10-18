package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

func hash(p string) (string, int64, error) {
	h := sha256.New()

	inf, err := os.Open(p)
	if err != nil {
		return "", 0, err
	}

	nbytes, err := io.Copy(h, inf)
	if err != nil {
		return "", 0, err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nbytes, nil
}

func main() {
	concurrency := flag.Int("j", runtime.NumCPU(), "number of concurrent operations.")
	flag.Parse()

	dirs := func() []string {
		if args := flag.Args(); len(args) != 0 {
			return args
		}
		return []string{"."}
	}

	fnch := make(chan string, *concurrency)
	go func() {
		for _, dir := range dirs() {
			filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
				if info.IsDir() {
					return nil
				}
				fnch <- path
				return nil
			})
		}
		close(fnch)
	}()

	sem := make(chan struct{}, *concurrency)
	for fn := range fnch {
		fn := fn
		sem <- struct{}{}
		go func() {
			h, nbytes, err := hash(fn)
			if err != nil {
				log.Printf("error: %v", err)
			}
			fmt.Printf("%s %d %s\n", h, nbytes, fn)
			<-sem
		}()
	}

	for i := 0; i < *concurrency; i++ {
		sem <- struct{}{}
	}
}
