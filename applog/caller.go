package applog

import (
	"path/filepath"
	"runtime"
	"strings"
)

// getCallerBeyondApplog 返回第一条不在 applog 包内的调用栈（实际打日志的业务代码位置）。
func getCallerBeyondApplog() (file string, line int) {
	var pcs [32]uintptr
	n := runtime.Callers(2, pcs[:])
	if n == 0 {
		return "?", 0
	}
	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()
		if frame.PC == 0 {
			return "?", 0
		}
		if isApplogFrame(frame) {
			if !more {
				return "?", 0
			}
			continue
		}
		return shortenCallerPath(frame.File), frame.Line
	}
}

func isApplogFrame(f runtime.Frame) bool {
	if f.Function != "" && strings.HasPrefix(f.Function, modulePrefix) {
		return true
	}
	slash := filepath.ToSlash(f.File)
	return strings.Contains(slash, "/infra-go/applog/")
}

// shortenCallerPath 尽量缩成从 internal/ 起的相对路径，便于阅读且与仓库布局一致。
func shortenCallerPath(file string) string {
	if file == "" || file == "?" {
		return file
	}
	s := filepath.ToSlash(file)
	if i := strings.Index(s, "/internal/"); i >= 0 {
		return s[i+1:]
	}
	return file
}
