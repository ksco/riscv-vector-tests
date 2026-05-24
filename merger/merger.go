package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func fatalIf(err error) {
	if err == nil {
		return
	}
	fmt.Printf("\033[0;1;31mfatal:\033[0m %s\n", err.Error())
	os.Exit(1)
}

var stage1OutputDirF = flag.String("stage1output", "", "stage1 output directory.")
var stage2OutputDirF = flag.String("stage2output", "", "stage2 output directory.")
var stage2PatchDirF = flag.String("stage2patch", "", "stage2 patches directory.")

func main() {
	flag.Parse()

	if stage1OutputDirF == nil || *stage1OutputDirF == "" {
		fatalIf(errors.New("-stage1output is required"))
	}

	if stage2OutputDirF == nil || *stage2OutputDirF == "" {
		fatalIf(errors.New("-stage2output is required"))
	}

	if stage2PatchDirF == nil || *stage2PatchDirF == "" {
		fatalIf(errors.New("-stage2patch is required"))
	}

	files, err := os.ReadDir(*stage1OutputDirF)
	fatalIf(err)

	r := regexp.MustCompile(`(?m)^\.word 0x[0-9a-f]+\n`)
	for _, file := range files {
		asmFilepath := filepath.Join(*stage1OutputDirF, file.Name())
		patchFilepath := filepath.Join(*stage2PatchDirF,
			strings.TrimSuffix(file.Name(), ".S")+".patch")

		patchContent, err := os.ReadFile(patchFilepath)
		patchContent = bytes.TrimPrefix(patchContent, []byte("bbl loader"))
		fatalIf(err)
		patches := bytes.Split(patchContent, []byte("---"))

		content, err := os.ReadFile(asmFilepath)
		fatalIf(err)
		content = bytes.Replace(content, []byte("  TEST_CASE(2, x0, 0x0)"), []byte(""), 1)
		contents := r.Split(string(content), -1)

		if len(contents) != len(patches) {
			fatalIf(fmt.Errorf("wrong patch generated for file %s, expected %d, got %d", file.Name(), len(contents), len(patches)))
		}

		builder := strings.Builder{}
		for idx, c := range contents {
			builder.WriteString(c)
			builder.Write(patches[idx])
		}

		writeTo(*stage2OutputDirF, file.Name(), []byte(builder.String()))
	}
}

func writeTo(path string, name string, contents []byte) {
	err := os.MkdirAll(path, 0777)
	fatalIf(err)
	err = os.WriteFile(filepath.Join(path, name), contents, 0644)
	fatalIf(err)
}
