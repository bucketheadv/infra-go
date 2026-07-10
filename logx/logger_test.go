package logx

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseLevel(t *testing.T) {
	lv, err := parseLevel("debug")
	if err != nil || lv != LevelDebug {
		t.Fatalf("parseLevel(debug) = %v, %v", lv, err)
	}
	if _, err := parseLevel("nope"); err == nil {
		t.Fatal("expected error for unknown level")
	}
	lv, err = parseLevel("")
	if err != nil || lv != LevelInfo {
		t.Fatalf("parseLevel(empty) = %v, %v", lv, err)
	}
}

func TestGetHasAndMustGet(t *testing.T) {
	if !Has(NameRoot) {
		t.Fatal("root should exist")
	}
	if Has("missing-logger") {
		t.Fatal("missing logger should not exist")
	}
	if Get("missing-logger") != Get(NameRoot) {
		t.Fatal("Get should fallback to root")
	}
	defer func() {
		if recover() == nil {
			t.Fatal("MustGet should panic")
		}
	}()
	_ = MustGet("missing-logger")
}

func TestWithFieldsAndJSONFormat(t *testing.T) {
	ctx := WithFields(context.Background(), map[string]string{"traceId": "abc"})
	fields := FieldsFrom(ctx)
	if fields["traceId"] != "abc" {
		t.Fatalf("FieldsFrom = %#v", fields)
	}
	rec := &Record{
		Time:   time.Date(2026, 5, 16, 10, 0, 0, 0, time.UTC),
		Level:  LevelInfo,
		Logger: "root",
		Msg:    "hello",
		File:   "a.go",
		Line:   1,
		Fields: fields,
	}
	data, err := formatJSONLine(rec, 48)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(data, []byte(`"traceId":"abc"`)) {
		t.Fatalf("json missing field: %s", data)
	}
}

func TestConsoleAppenderWrite(t *testing.T) {
	var buf bytes.Buffer
	c := newConsoleAppender("text", "", false, nil, nil, 48)
	c.out = &buf
	c.Append(&Record{
		Time:   time.Date(2026, 5, 16, 10, 0, 0, 0, time.UTC),
		Level:  LevelInfo,
		Logger: "root",
		Msg:    "x",
		File:   "a.go",
		Line:   1,
	})
	if !strings.Contains(buf.String(), "x") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestLoadUnknownLevel(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	content := `
root:
  level: nope
  appenders: []
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Load(path); err == nil {
		t.Fatal("expected load error for unknown level")
	}
}

func TestLoadFailureClosesAppenders(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "app.log")
	cfg := filepath.Join(dir, "bad.yaml")
	content := `
appenders:
  file:
    type: rollingFile
    path: ` + logPath + `
    maxLinesPerFile: 100
root:
  level: nope
  appenders: [file]
`
	if err := os.WriteFile(cfg, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Load(cfg); err == nil {
		t.Fatal("expected load error")
	}
	// 失败后文件句柄应已关闭，允许删除日志文件（Windows 上若未关闭会失败）。
	_ = os.Remove(logPath)
}

func TestCloseConcurrentWithInfo(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "app.log")
	cfg := filepath.Join(dir, "log.yaml")
	content := `
callerFileMaxLen: 48
appenders:
  file:
    type: rollingFile
    path: ` + logPath + `
    maxLinesPerFile: 1000
    layout: text
root:
  level: info
  appenders: [file]
`
	if err := os.WriteFile(cfg, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Load(cfg); err != nil {
		t.Fatal(err)
	}
	defer Close()

	ctx := context.Background()
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 200; i++ {
			Infof(ctx, NameRoot, "msg-%d", i)
		}
	}()
	for i := 0; i < 20; i++ {
		Close()
		if err := Load(cfg); err != nil {
			t.Fatalf("reload: %v", err)
		}
	}
	<-done
	Close()
	// 不应 panic；关闭后仍可写到 bootstrap console
	Infof(ctx, NameRoot, "after-close")
}
