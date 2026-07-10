package logx

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type rollingFileAppender struct {
	// path 当前日志文件路径。
	path string
	// maxLinesPerFile 单文件最大行数，超出后滚动。
	maxLinesPerFile int64
	// retentionDays 历史文件保留天数。
	retentionDays int
	// layout 布局类型（text / json）。
	layout string
	// pattern 文本布局 pattern。
	pattern string
	// colored 是否启用 ANSI 着色。
	colored bool
	// levelColors 级别名到颜色的映射。
	levelColors map[Level]string
	// fieldColors 字段名到颜色的映射。
	fieldColors map[string]string
	// callerFileMax 调用方文件路径截断最大长度。
	callerFileMax int

	// mu 保护文件句柄与行计数。
	mu sync.Mutex
	// file 当前打开的日志文件。
	file *os.File
	// lineCount 当前文件已写入行数。
	lineCount int64
	// prefix 滚动文件名前缀。
	prefix string
	// rotateCooldown 旋转失败后的退避计数，避免每条日志都重试旋转。
	rotateCooldown int
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
	// 末尾无换行时仍计为一行，避免少计导致多写一行才滚动。
	if data[len(data)-1] != '\n' {
		n++
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
			reportAppendError("json format", err)
			data = formatTextLine(rec, r.callerFileMax, false, nil)
		}
		data = append(data, '\n')
	case "pattern":
		data = ensureTrailingNewline(formatPatternLine(rec, r.callerFileMax, r.colored, r.fieldColors, r.levelColors, r.pattern))
	default:
		data = formatTextLine(rec, r.callerFileMax, r.colored, r.levelColors)
	}

	if r.lineCount >= r.maxLinesPerFile {
		if r.rotateCooldown > 0 {
			r.rotateCooldown--
		} else if err := r.rotateUnlocked(); err != nil {
			reportAppendError("rotate", err)
			r.rotateCooldown = 256
			// 旋转失败时尽量保证当前仍可写：若 file 仍为 nil 则尝试重新打开。
			if r.file == nil {
				if oerr := r.reopenAppendUnlocked(false); oerr != nil {
					reportAppendError("reopen "+r.path, oerr)
					return
				}
			}
		} else {
			r.rotateCooldown = 0
		}
	}
	if r.file == nil {
		reportAppendError("append", fmt.Errorf("file is closed: %s", r.path))
		return
	}
	if _, err := r.file.Write(data); err != nil {
		reportAppendError("write "+r.path, err)
		return
	}
	// 按实际写入的换行数累计，避免 pattern 含多个 %n 时与物理行数偏离。
	r.lineCount += countNewlines(data)
}

func countNewlines(data []byte) int64 {
	var n int64
	for _, b := range data {
		if b == '\n' {
			n++
		}
	}
	if n == 0 {
		return 1
	}
	return n
}

func (r *rollingFileAppender) reopenAppendUnlocked(keepLineCount bool) error {
	f, err := os.OpenFile(r.path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	r.file = f
	if !keepLineCount {
		r.lineCount = r.countLinesInFile(r.path)
	}
	return nil
}

func (r *rollingFileAppender) syncFile() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.file != nil {
		if err := r.file.Sync(); err != nil {
			reportAppendError("sync "+r.path, err)
		}
	}
}

func (r *rollingFileAppender) rotateUnlocked() error {
	old := r.file
	if old != nil {
		_ = old.Sync()
		_ = old.Close()
		r.file = nil
	}

	rot := fmt.Sprintf("%s.%d", r.path, time.Now().UnixNano())
	if err := os.Rename(r.path, rot); err != nil && !os.IsNotExist(err) {
		// 重命名失败：回到原文件继续追加，保留当前行计数（避免全量重读大文件）。
		if oerr := r.reopenAppendUnlocked(true); oerr != nil {
			return fmt.Errorf("rename: %w; reopen: %v", err, oerr)
		}
		return fmt.Errorf("rename: %w", err)
	}

	f, err := os.OpenFile(r.path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		// 原文件已搬走但新文件创建失败：重新打开并重计行数。
		if oerr := r.reopenAppendUnlocked(false); oerr != nil {
			return fmt.Errorf("create: %w; reopen: %v", err, oerr)
		}
		return fmt.Errorf("create: %w", err)
	}
	r.file = f
	r.lineCount = 0
	r.rotateCooldown = 0
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
