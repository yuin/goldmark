package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type caseFolding struct {
	Class byte
	From  rune
	To    []rune
}

func unicodeCaseFoldingMapSubCommand(args []string) {
	cmdName := "unicode-case-folding-map"
	cmd := flag.NewFlagSet(cmdName, flag.ExitOnError)
	cmd.Usage = func() {
		_, _ = fmt.Fprintf(cmd.Output(), `Usage of %s:

  Generate input JSON data for emb-structs subcommand from unicode.org

`, cmdName)
		cmd.PrintDefaults()
	}

	outputPath := cmd.String("o", "", "output file path(required)")
	unicodeVersion := cmd.String("u", "15.0.0", "unicode version")

	if err := cmd.Parse(args); err != nil ||
		len(*outputPath) == 0 {
		usage(cmd.Usage, err)
	}

	url := "http://www.unicode.org/Public/" + *unicodeVersion + "/ucd/CaseFolding.txt"

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Failed to get CaseFolding.txt: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to get CaseFolding.txt: %v\n", err)
		os.Exit(1)
	}

	buf := bytes.NewBuffer(bs)
	scanner := bufio.NewScanner(buf)

	embstructmap := make(map[string]any)
	embstructmap["prefix"] = "unicodeCaseFolding"
	embstructmap["types"] = map[string]any{
		"From": "rune",
		"To":   "[]rune",
	}
	var data []any

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || len(strings.TrimSpace(line)) == 0 {
			continue
		}
		line = strings.Split(line, "#")[0]
		parts := strings.Split(line, ";")
		for i, p := range parts {
			parts[i] = strings.TrimSpace(p)
		}
		cf := caseFolding{}
		v, _ := strconv.ParseInt(parts[0], 16, 32)
		cf.From = rune(int32(v))
		cf.Class = parts[1][0]
		for _, v := range strings.Split(parts[2], " ") {
			c, _ := strconv.ParseInt(v, 16, 32)
			cf.To = append(cf.To, rune(int32(c)))
		}
		if cf.Class != 'C' && cf.Class != 'F' {
			continue
		}
		var tos []string
		for _, v := range cf.To {
			tos = append(tos, fmt.Sprintf("%d", v))
		}
		data = append(data, map[string]any{
			"From": fmt.Sprintf("0x%x", cf.From),
			"To":   tos,
		})
	}
	embstructmap["data"] = data
	jsonData, err := json.MarshalIndent(embstructmap, "", "  ")
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(*outputPath, jsonData, 0644)
	if err != nil {
		panic(err)
	}

}
