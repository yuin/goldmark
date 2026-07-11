package goldmark_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/parser"
)

const thisPackage = "github.com/yuin/goldmark/v2"

func TestDoc(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	p := parser.New()
	source, err := os.ReadFile("README.md")
	if err != nil {
		panic(err)
	}
	doc := p.Parse(source)
	var codeBlocks []*ast.CodeBlock
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if cb, ok := n.(*ast.CodeBlock); ok && entering {
			codeBlocks = append(codeBlocks, cb)
		}
		return ast.WalkContinue, nil
	})
	for _, cb := range codeBlocks {
		l, ok := cb.Language(source)
		if !ok || l.Str(source) != "go" {
			continue
		}
		if strings.Contains(cb.Info.Str(source), "no-run") {
			continue
		}
		code := string(cb.Value.Str(source))

		pat := regexp.MustCompile(`(?s)import\s*\((.*?)\)`)
		importSection := pat.FindStringSubmatch(code)
		body := pat.ReplaceAllString(code, "")

		var thirdPartyImports []string
		if len(importSection) > 1 {
			for line := range strings.SplitSeq(importSection[1], "\n") {
				line = strings.TrimSpace(line)
				if line != "" && strings.Contains(line, ".") && !strings.Contains(line, thisPackage) {
					thirdPartyImports = append(thirdPartyImports, line[1:len(line)-1]) // remove quotes
				}
			}
		}

		d, err := os.MkdirTemp("", "goldmark_test")
		if err != nil {
			panic(err)
		}
		defer os.RemoveAll(d)
		gomod, err := os.Create(filepath.Join(d, "go.mod"))
		if err != nil {
			panic(err)
		}
		majarVersion := string([]byte{thisPackage[len(thisPackage)-1]})
		if _, err := gomod.WriteString("module test\n\ngo " + runtime.Version()[2:] + "\nrequire(\n" + thisPackage + " v" + majarVersion + ".0.0\n)\n"); err != nil {
			panic(err)
		}
		if _, err := gomod.WriteString("replace " + thisPackage + " v" + majarVersion + ".0.0 => " + cwd + "\n"); err != nil {
			panic(err)
		}
		if err := gomod.Close(); err != nil {
			panic(err)
		}

		w, err := os.Create(filepath.Join(d, "main.go"))
		if err != nil {
			panic(err)
		}
		if _, err := w.WriteString("package main\n"); err != nil {
			panic(err)
		}
		if len(importSection) > 1 {
			if _, err := w.WriteString("import (\n" + importSection[1] + "\n)\n"); err != nil {
				panic(err)
			}
		}
		if _, err := w.WriteString("func main() {\n"); err != nil {
			panic(err)
		}
		if _, err := w.WriteString(string(body) + "}\n"); err != nil {
			panic(err)
		}

		if err := w.Close(); err != nil {
			panic(err)
		}

		for _, imp := range thirdPartyImports {
			cmd := exec.Command("go", "get", imp+"@latest")
			cmd.Dir = d
			if err := cmd.Run(); err != nil {
				panic(err)
			}
		}

		cmd := exec.Command("go", "mod", "tidy")
		cmd.Dir = d
		if err := cmd.Run(); err != nil {
			panic(err)
		}

		cmd = exec.Command("go", "run", "main.go")
		cmd.Dir = d
		output, err := cmd.CombinedOutput()
		if err != nil {
			content, _ := os.ReadFile(w.Name())
			t.Errorf("Code block failed to run: \n  Line: %v\n  Code:\n%s  Output:\n%s",
				posToLine(source, cb.Value.Segments()[0].Start), addIndent(string(content), 4), addIndent(string(output), 4))
		}
	}

}

func posToLine(source []byte, pos int) int {
	line := 1
	for i := 0; i < pos && i < len(source); i++ {
		if source[i] == '\n' {
			line++
		}
	}
	return line
}

func addIndent(s string, w int) string {
	indent := strings.Repeat(" ", w)
	return strings.TrimRight(indent+regexp.MustCompile(`(?m)^`).ReplaceAllString(s, indent)[len(indent):], " ") + "\n"
}
