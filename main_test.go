package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"os"
	"testing"
)

var update bool

func init() {
	flag.BoolVar(&update, "update", false, "update gold files")
}

// get gold data file path based on the words variable
func goldFile() string {
	if words {
		return "testdata/words.gold.txt"
	}
	return "testdata/lines.gold.txt"
}

func mustUpdateGold() {
	buf := &bytes.Buffer{}
	out = buf
	if err := catFileGroups("A", "B"); err != nil {
		panic(err)
	}
	gold := goldFile()
	if err := ioutil.WriteFile(gold, buf.Bytes(), os.ModePerm); err != nil {
		panic(err)
	}
	// reset just in case
	out = os.Stdout
}

func TestOutput(t *testing.T) {
	for _, w := range []bool{true, false} {
		metsFile = "testdata/mets.xml"
		fileInfo = true
		words = w
		if update {
			mustUpdateGold()
		}
		buf := &bytes.Buffer{}
		out = buf
		if err := catFileGroups("A", "B"); err != nil {
			t.Fatalf("got error: %v", err)
		}
		in, err := os.Open(goldFile())
		if err != nil {
			panic(err)
		}
		defer in.Close()
		want, err := ioutil.ReadAll(in)
		if err != nil {
			panic(err)
		}
		if got := buf.Bytes(); !bytes.Equal(got, want) {
			t.Fatalf("expected\n%q\ngot\n%q\n", want, got)
		}
	}
}
