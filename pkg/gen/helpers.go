package gen

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"
	"time"
)

func findConfig() (*CDKConfig, string, error) {

	configPath := ""
	err := filepath.Walk(".",
		func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.Name() == "cdk.json" {
				fmt.Println("found:", p)
				configPath = p
				return io.EOF
			}
			return nil
		})
	if err != io.EOF && err != nil {
		return nil, "", err
	}

	jsonFile, err := os.Open(configPath)
	if err != nil {
		return nil, "", err
	}
	byteValue, _ := ioutil.ReadAll(jsonFile)
	result := &CDKConfig{}
	err = json.Unmarshal(byteValue, result)

	configDir := path.Dir(configPath)

	return result, configDir, err
}

func fmtDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	return fmt.Sprintf("%02d:%02d", h, m)
}

func isContext(input reflect.Type) bool {
	return input.Implements(reflect.TypeOf((*context.Context)(nil)).Elem())
}

func openFile(file string) (*os.File, error) {

	dir := path.Dir(file)

	if _, err := os.Stat(file); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return nil, err
		}
		return os.Create(file)
	} else if err != nil {
		return nil, err
	}

	return os.OpenFile(file, os.O_WRONLY, 0644)
}

func buildMain(dir, key, content string) (string, error) {
	pkgDir := path.Join(dir, "assets", key)
	mainPath := path.Join(pkgDir, "main.go")
	binPath := path.Join(pkgDir, "main.out")

	// 1. write contect to main.go file
	f, err := openFile(mainPath)
	if err != nil {
		return "", err
	}

	if _, err := f.WriteString(content); err != nil {
		return "", err
	}

	// 2. compile main.go file to main.out
	cmd := exec.Command("go", "build", "-ldflags", "-s -w", "-o", "main.out")
	cmd.Dir = pkgDir
	envs := append(os.Environ(), "GOOS=linux", "GOARCH=amd64", "GOBIN="+pkgDir)
	cmd.Env = envs
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		fmt.Println(cmd)
		fmt.Println(content)
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return "", err
	}

	// 3. zip result. zip name is a sha1 hash of the binary
	zipPath := path.Join(dir, "assets", filehash(binPath)) + ".zip"

	if err := zipFile(zipPath, binPath); err != nil {
		return "", err
	}

	return zipPath, nil
}

func generate(tmpl *template.Template, asset Asset) (string, error) {
	var b strings.Builder
	if err := tmpl.Execute(&b, asset); err != nil {
		return "", err
	}
	return b.String(), nil
}

func getTemplate(t AssetType) (*template.Template, error) {

	var tmpl string

	switch t {
	case CommandType:
		tmpl = resolverTmpl
	case QueryType:
		tmpl = resolverTmpl
	case MutationType:
		tmpl = mutationTmpl
	case FunctionType:
		return nil, errors.New("not implemented")
	default:
		return nil, errors.New("not implemented")
	}
	return template.New(string(t)).Parse(tmpl)
}
