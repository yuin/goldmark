package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime/pprof"
	"strconv"
	"time"

	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer/html"
)

func main() {
	n := 50
	file := "_data.md"
	if len(os.Args) > 1 {
		n, _ = strconv.Atoi(os.Args[1])
	}
	if len(os.Args) > 2 {
		file = os.Args[2]
	}
	if len(os.Args) > 3 {
		f, err := os.Create(os.Args[3])
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal(err)
		}
		defer pprof.StopCPUProfile()
	}
	source, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	p := parser.New()
	r := html.New(html.WithXHTML(), html.WithUnsafe())
	var out bytes.Buffer

	sum := time.Duration(0)
	for i := 0; i < n; i++ {
		start := time.Now()
		out.Reset()
		doc := p.Parse(source)
		if err := r.Render(&out, source, doc); err != nil {
			panic(err)
		}
		sum += time.Since(start)
	}
	fmt.Printf("------- goldmark -------\n")
	fmt.Printf("file: %s\n", file)
	fmt.Printf("iteration: %d\n", n)
	fmt.Printf("average: %.10f sec\n", float64((int64(sum)/int64(n)))/1000000000.0)
}
