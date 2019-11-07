package store

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

type ProgressReader struct {
	files   []string
	current int
	reader  io.ReadCloser

	p   *mpb.Progress
	bar *mpb.Bar
}

func NewProgressReader(files []string) ProgressReader {

	p := mpb.New(
		mpb.WithWidth(60),
		mpb.WithRefreshRate(100*time.Millisecond),
	)

	return ProgressReader{
		files: files,
		p:     p,
	}
}

func (pr *ProgressReader) Next(name string) (io.ReadCloser, error) {
	if pr.current > len(pr.files) {
		return nil, nil
	}

	f, err := os.Open(pr.files[pr.current])
	if err != nil {
		return nil, err
	}
	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}

	bar := pr.p.AddBar(stat.Size(), mpb.BarStyle("[=>-|"),
		mpb.PrependDecorators(
			decor.OnComplete(
				decor.Name(fmt.Sprintf("  %-20v", name)),
				fmt.Sprintf("✅ %-20v", name),
			),
		),
		mpb.AppendDecorators(
			decor.EwmaETA(decor.ET_STYLE_GO, 90),
			decor.EwmaSpeed(decor.UnitKiB, "  % .2f", 60),
			decor.CountersKibiByte("  % .2f / % .2f"),
		),
	)

	if pr.reader != nil {
		pr.reader.Close()
	}

	pr.reader = bar.ProxyReader(f)
	pr.current++
	return pr.reader, nil
}

func readFile(file string) (io.Reader, int64, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return nil, 0, err
	}

	var content bytes.Buffer
	buf := bufio.NewWriter(&content)
	if _, err := io.Copy(buf, f); err != nil {
		return nil, 0, err
	}
	return &content, stat.Size(), nil
}
