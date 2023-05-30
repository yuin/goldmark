package main

import (
	"archive/zip"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

type TestCase struct {
	Example  int    `json:"example"`
	Markdown string `json:"markdown"`
}

func main() {
	corpus_out := os.Args[1]
	if !strings.HasSuffix(corpus_out, ".zip") {
		log.Fatalln("Expected command line:", os.Args[0], "<corpus_output>.zip")
	}

	zip_file, err := os.Create(corpus_out)

	zip_writer := zip.NewWriter(zip_file)

	if err != nil {
		log.Fatalln("Failed creating file:", err)
	}

	json_corpus := "_test/spec.json"
	bs, err := ioutil.ReadFile(json_corpus)
	if err != nil {
		log.Fatalln("Could not open file:", json_corpus)
		panic(err)
	}
	var testCases []TestCase
	if err := json.Unmarshal(bs, &testCases); err != nil {
		panic(err)
	}

	for _, c := range testCases {
		file_in_zip := "example-" + strconv.Itoa(c.Example)
		f, err := zip_writer.Create(file_in_zip)
		if err != nil {
			log.Fatal(err)
		}
		_, err = f.Write([]byte(c.Markdown))
		if err != nil {
			log.Fatalf("Failed to write file: %s into zip file", file_in_zip)
		}
	}

	err = zip_writer.Close()
	if err != nil {
		log.Fatal("Failed to close zip writer", err)
	}

	zip_file.Close()
}