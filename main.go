package main // import "github.com/finkf/pxmlcat"

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/finkf/gocrd/mets"
	"github.com/finkf/gocrd/page"
)

var (
	words    bool
	fileInfo bool
	metsFile string
	out      io.Writer
	ai       int
	bi       int
)

func init() {
	flag.BoolVar(&words, "words", false,
		"split page xml on words")
	flag.IntVar(&ai, "ai", 0,
		"set index of first file group's TextEquiv")
	flag.IntVar(&bi, "bi", 0,
		"set index of second file group's TextEquiv")
	flag.BoolVar(&fileInfo, "file-info", false,
		"add file info header to each block")
	flag.StringVar(&metsFile, "mets", "mets.xml",
		"set path to mets file")
}

func main() {
	flag.Parse()
	if err := run(); err != nil {
		fmt.Printf("[error] %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(flag.Args()) != 2 {
		return fmt.Errorf("invalid usage: file groups are missing")
	}
	out = os.Stdout
	return catFileGroups(flag.Args()[0], flag.Args()[1])
}

func catFileGroups(fg1, fg2 string) error {
	mets, err := mets.Open(metsFile)
	if err != nil {
		return err
	}
	f1s, err := fileGroup(mets, fg1)
	if err != nil {
		return err
	}
	f2s, err := fileGroup(mets, fg2)
	if err != nil {
		return err
	}
	if len(f1s) != len(f2s) {
		return fmt.Errorf("different file group sizes")
	}
	for i := 0; i < len(f1s); i++ {
		if err := catPages(f1s[i], f2s[i]); err != nil {
			return err
		}
	}
	return nil
}

func fileGroup(mets mets.Mets, fg string) ([]mets.File, error) {
	fs, ok := mets.FindFileGrp(fg)
	if !ok {
		return nil, fmt.Errorf("invalid file group: %s", fg)
	}
	return fs, nil
}

func catPages(f1, f2 mets.File) error {
	p1, err := page.Open(localFilePath(f1))
	if err != nil {
		return err
	}
	p2, err := page.Open(localFilePath(f2))
	if err != nil {
		return err
	}
	r1 := p1.Regions()
	r2 := p2.Regions()
	if len(r1) != len(r2) {
		return fmt.Errorf("different region sizes: %s", localFilePath(f1))
	}
	for i := 0; i < len(r1); i++ {
		if err := catRegions(f1.FLocat.URL, r1[i], r2[i]); err != nil {
			return err
		}
	}
	return nil
}

func catRegions(fn string, r1, r2 page.Region) error {
	l1 := r1.Lines()
	l2 := r2.Lines()
	if len(l1) != len(l2) {
		return fmt.Errorf("different line sizes")
	}
	for i := 0; i < len(l1); i++ {
		if err := catLines(fn, l1[i], l2[i]); err != nil {
			return err
		}
	}
	return nil
}

func catLines(fn string, l1, l2 page.Line) error {
	if !words {
		return cat(fn, l1.ID, l1, l2)
	}
	w1 := l1.Words()
	w2 := l2.Words()
	if len(w1) != len(w2) {
		return fmt.Errorf("different word sizes")
	}
	for i := 0; i < len(w1); i++ {
		if err := cat(fn, w1[i].ID, w1[i], w2[i]); err != nil {
			return err
		}
	}
	return nil
}

type unicoder interface {
	TextEquivUnicodeAt(int) (string, bool)
}

func cat(fn, id string, a, b unicoder) error {
	if fileInfo {
		if _, err := fmt.Fprintf(out, "%s:%s\n", fn, id); err != nil {
			return err
		}
	}
	astr, _ := a.TextEquivUnicodeAt(ai)
	bstr, _ := b.TextEquivUnicodeAt(bi)
	_, err := fmt.Fprintf(out, "%s\n%s\n", astr, bstr)
	return err
}

func localFilePath(f mets.File) string {
	if strings.HasPrefix(f.FLocat.URL, "file://") {
		return f.FLocat.URL[7:]
	}
	return f.FLocat.URL
}
