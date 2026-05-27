package logx

import (
	"path/filepath"
	"runtime"
	"strings"
)

// getCallerBeyondLogx 返回第一条不在 logx 包内的调用栈（实际打日志的业务代码位置）。
func getCallerBeyondLogx() (file string, line int) {
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
		if isLogxFrame(frame) {
			if !more {
				return "?", 0
			}
			continue
		}
		return shortenCallerPath(frame.File), frame.Line
	}
}

func isLogxFrame(f runtime.Frame) bool {
	if f.Function != "" && strings.HasPrefix(f.Function, modulePrefix) {
		return true
	}
	slash := filepath.ToSlash(f.File)
	return strings.Contains(slash, "/infra-go/logx/")
}

// shortenCallerPath 尽量缩短显示路径；业务代码常截成 internal/...；标准库截成 GOROOT/src 后的 net/http/...，避免把 net/http/internal 误收成项目 internal。
func shortenCallerPath(file string) string {
	if file == "" || file == "?" {
		return file
	}
	slashPath := filepath.ToSlash(filepath.Clean(file))
	if gr := runtime.GOROOT(); gr != "" {
		srcPrefix := filepath.ToSlash(filepath.Clean(gr)) + "/src/"
		if strings.HasPrefix(slashPath, srcPrefix) {
			return strings.TrimPrefix(slashPath, srcPrefix)
		}
	}
	if i := strings.Index(slashPath, "/internal/"); i >= 0 {
		return slashPath[i+1:]
	}
	if i := strings.Index(slashPath, "/infra-go/logx/"); i >= 0 {
		return slashPath[i+1:]
	}
	return file
}
