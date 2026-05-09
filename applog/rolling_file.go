package applog

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type rollingFileAppender struct {
	path            string
	maxLinesPerFile int64
	retentionDays   int
	layout          string
	pattern         string
	colored         bool
	levelColors     map[Level]string
	fieldColors     map[string]string
	callerFileMax   int

	mu        sync.Mutex
	file      *os.File
	lineCount int64
	prefix    string
}

func newRollingFileAppender(
	path string,
	maxLines int,
	retentionDays int,
	layout string,
	pattern string,
	colored bool,
	levelColors map[Level]string,
	fieldColors map[string]string,
	callerMax int,
) (*rollingFileAppender, error) {
	if path == "" {
		return nil, os.ErrInvalid
	}
	if maxLines <= 0 {
		maxLines = 100_000
	}
	if layout == "" {
		layout = "text"
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	base := filepath.Base(path)
	r := &rollingFileAppender{
		path:            path,
		maxLinesPerFile: int64(maxLines),
		retentionDays:   retentionDays,
		layout:          layout,
		pattern:         pattern,
		colored:         colored,
		levelColors:     levelColors,
		fieldColors:     fieldColors,
		callerFileMax:   callerMax,
		prefix:          base + ".",
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}
	r.file = f
	r.lineCount = r.countLinesInFile(path)
	r.purgeOldFiles()
	return r, nil
}

func (r *rollingFileAppender) countLinesInFile(path string) int64 {
	data, err := os.ReadFile(path)
	if err != nil || len(data) == 0 {
		return 0
	}
	var n int64
	for _, b := range data {
		if b == '\n' {
			n++
		}
	}
	return n
}

func (r *rollingFileAppender) Append(rec *Record) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var data []byte
	var err error
	switch r.layout {
	case "json":
		data, err = formatJSONLine(rec, r.callerFileMax)
		if err != nil {
			data = formatTextLine(rec, r.callerFileMax, false, nil)
		}
		data = append(data, '\n')
	case "pattern":
		data = formatPatternLine(rec, r.callerFileMax, r.colored, r.fieldColors, r.levelColors, r.pattern)
	default:
		data = formatTextLine(rec, r.callerFileMax, r.colored, r.levelColors)
	}

	if r.lineCount >= r.maxLinesPerFile {
		_ = r.rotateUnlocked()
	}
	if r.file == nil {
		return
	}
	_, _ = r.file.Write(data)
	r.lineCount++
}

func (r *rollingFileAppender) rotateUnlocked() error {
	if r.file != nil {
		_ = r.file.Sync()
		_ = r.file.Close()
		r.file = nil
	}
	rot := fmt.Sprintf("%s.%d", r.path, time.Now().UnixNano())
	if err := os.Rename(r.path, rot); err != nil && !os.IsNotExist(err) {
		f, oerr := os.OpenFile(r.path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
		if oerr != nil {
			return oerr
		}
		r.file = f
		return err
	}
	f, err := os.OpenFile(r.path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	r.file = f
	r.lineCount = 0
	r.purgeOldFilesUnlocked()
	return nil
}

func (r *rollingFileAppender) purgeOldFiles() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.purgeOldFilesUnlocked()
}

func (r *rollingFileAppender) purgeOldFilesUnlocked() {
	if r.retentionDays <= 0 {
		return
	}
	dir := filepath.Dir(r.path)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	cutoff := time.Now().AddDate(0, 0, -r.retentionDays)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, r.prefix) || name == filepath.Base(r.path) {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			_ = os.Remove(filepath.Join(dir, name))
		}
	}
}

func (r *rollingFileAppender) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.file != nil {
		_ = r.file.Sync()
		err := r.file.Close()
		r.file = nil
		return err
	}
	return nil
}
