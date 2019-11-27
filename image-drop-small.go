package main

import (
	"flag"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func images(root string) <-chan string {
	ret := make(chan string, 16)
	go func() {
		defer close(ret)
		_ = filepath.Walk(root, func(fn string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if fi.IsDir() {
				return nil
			}
			ext := strings.ToLower(path.Ext(fn))

			if ext == ".jpg" || ext == ".jpeg" || ext == ".png" {
				ret <- fn
			}
			return nil
		})
	}()
	return ret
}

func filter(ic image.Config, fn string) {
	if ic.Height > config.maxHeight && ic.Width > config.maxWidth {
		fmt.Println("skip", fn, ic.Height, ic.Width)
		return
	}
	if ic.Width < config.minWidth && ic.Height < config.minHeight {
		err := os.Remove(fn)
		fmt.Println("drop", fn, ic.Width, ic.Height, err)
		return
	}

	if ic.Width*config.ratio < ic.Height || ic.Height*config.ratio < ic.Width {
		target := path.Join(config.recycle, path.Base(fn))
		err := os.Rename(fn, target)
		fmt.Println("mv", fn, ic.Width, ic.Height, err)
		return
	}
}
func main() {
	err := os.MkdirAll(config.recycle, 0o755)
	if err != nil {
		log.Println(err)
		return
	}
	for fn := range images(config.file) {
		name := path.Base(fn)
		lname := strings.ToLower(name)
		if lname != name {
			nfn := path.Join(path.Dir(fn), lname)
			if err := os.Rename(fn, nfn); err == nil {
				fn = nfn
			}
		}
		if f, err := os.Open(fn); err == nil {
			if img, _, err := image.DecodeConfig(f); err == nil {
				filter(img, fn)
			}
			_ = f.Close()
		}
	}
}

var config struct {
	file      string
	recycle   string
	minWidth  int
	minHeight int
	maxWidth  int
	maxHeight int
	ratio     int
}

func init() {
	flag.StringVar(&config.file, "file", "", "")
	flag.StringVar(&config.recycle, "recycle", "/data.d/spiding/recycle", "")
	flag.IntVar(&config.minWidth, "min-width", 400, "")
	flag.IntVar(&config.minHeight, "min-height", 400, "")
	flag.IntVar(&config.maxWidth, "max-width", 1000, "")
	flag.IntVar(&config.maxHeight, "max-height", 1000, "")
	flag.IntVar(&config.ratio, "ratio", 3, "")

	flag.Parse()
}
