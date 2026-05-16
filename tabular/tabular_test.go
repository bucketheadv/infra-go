package tabular

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var (
	// TestOutputToDownloads 为 true 时，测试文件输出到 ~/Downloads/infra-go-tabular-tests。
	TestOutputToDownloads = true
)

type personRow struct {
	ID      int       `json:"id"`
	Name    string    `json:"name"`
	Score   float64   `json:"score"`
	Enabled bool      `json:"enabled"`
	At      time.Time `json:"at"`
}

type personRowWithTitle struct {
	ID      int       `title:"编号"`
	Name    string    `title:"姓名"`
	Score   float64   `title:"得分"`
	Enabled bool      `title:"启用"`
	At      time.Time `title:"时间"`
}

func TestWriteReadCSV(t *testing.T) {
	dir := outputDir(t, "TestWriteReadCSV")
	file := filepath.Join(dir, "data.csv")
	t.Logf("TestWriteReadCSV file=%s", file)
	input := sampleRows()
	if err := WriteCSV(file, input); err != nil {
		t.Fatalf("write csv failed: %v", err)
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read csv raw failed: %v", err)
	}
	t.Logf("TestWriteReadCSV raw=%s", string(raw))

	got, err := ReadCSV[personRow](file)
	if err != nil {
		t.Fatalf("read csv failed: %v", err)
	}
	t.Logf("TestWriteReadCSV input=%+v", input)
	t.Logf("TestWriteReadCSV output=%+v", got)
	assertRows(t, got, input)
}

func TestReadCSVHeaderRecognition(t *testing.T) {
	dir := outputDir(t, "TestReadCSVHeaderRecognition")
	file := filepath.Join(dir, "reorder.csv")
	t.Logf("TestReadCSVHeaderRecognition file=%s", file)
	content := "name,id,enabled,score,at\nalice,1,true,88.5,2026-05-16T10:00:00Z\n"
	if err := os.WriteFile(file, []byte(content), 0o644); err != nil {
		t.Fatalf("write file failed: %v", err)
	}
	got, err := ReadCSV[personRow](file)
	if err != nil {
		t.Fatalf("read csv failed: %v", err)
	}
	want := []personRow{
		{ID: 1, Name: "alice", Score: 88.5, Enabled: true, At: time.Date(2026, 5, 16, 10, 0, 0, 0, time.UTC)},
	}
	t.Logf("TestReadCSVHeaderRecognition output=%+v", got)
	assertRows(t, got, want)
}

func TestReadCSVFromReader(t *testing.T) {
	content := "id,name,score,enabled,at\n1,alice,88.5,true,2026-05-16T10:00:00Z\n"
	got, err := ReadCSVFromReader[personRow](strings.NewReader(content))
	if err != nil {
		t.Fatalf("read csv from reader failed: %v", err)
	}
	want := []personRow{
		{ID: 1, Name: "alice", Score: 88.5, Enabled: true, At: time.Date(2026, 5, 16, 10, 0, 0, 0, time.UTC)},
	}
	t.Logf("TestReadCSVFromReader output=%+v", got)
	assertRows(t, got, want)
}

func TestWriteReadExcelByNameAndIndex(t *testing.T) {
	dir := outputDir(t, "TestWriteReadExcelByNameAndIndex")
	file := filepath.Join(dir, "data.xlsx")
	t.Logf("TestWriteReadExcelByNameAndIndex file=%s", file)
	input := sampleRows()
	if err := WriteExcel(file, "People", input); err != nil {
		t.Fatalf("write excel failed: %v", err)
	}

	gotByName, err := ReadExcel[personRow](file, SheetSelector{Name: "People"})
	if err != nil {
		t.Fatalf("read excel by name failed: %v", err)
	}
	t.Logf("TestWriteReadExcelByNameAndIndex byName=%+v", gotByName)
	assertRows(t, gotByName, input)

	gotByIndex, err := ReadExcel[personRow](file, SheetSelector{Index: 0})
	if err != nil {
		t.Fatalf("read excel by index failed: %v", err)
	}
	t.Logf("TestWriteReadExcelByNameAndIndex byIndex=%+v", gotByIndex)
	assertRows(t, gotByIndex, input)
}

func TestReadExcelFromReader(t *testing.T) {
	dir := outputDir(t, "TestReadExcelFromReader")
	file := filepath.Join(dir, "reader.xlsx")
	input := sampleRows()
	if err := WriteExcel(file, "People", input); err != nil {
		t.Fatalf("write excel failed: %v", err)
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read excel raw failed: %v", err)
	}
	got, err := ReadExcelFromReader[personRow](bytes.NewReader(raw), SheetSelector{Name: "People"})
	if err != nil {
		t.Fatalf("read excel from reader failed: %v", err)
	}
	t.Logf("TestReadExcelFromReader output=%+v", got)
	assertRows(t, got, input)
}

func TestWriteReadCSVStreamStruct(t *testing.T) {
	dir := outputDir(t, "TestWriteReadCSVStreamStruct")
	file := filepath.Join(dir, "struct-stream.csv")
	t.Logf("TestWriteReadCSVStreamStruct file=%s", file)

	input := sampleRows()
	err := WriteCSVStream(file, func(write func(personRow) error) error {
		for _, r := range input {
			if err := write(r); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("write csv struct stream failed: %v", err)
	}

	got := make([]personRow, 0, len(input))
	err = ReadCSVStream[personRow](file, func(rowNum int, item personRow) error {
		t.Logf("csv struct stream row=%d item=%+v", rowNum, item)
		got = append(got, item)
		return nil
	})
	if err != nil {
		t.Fatalf("read csv struct stream failed: %v", err)
	}
	assertRows(t, got, input)
}

func TestReadCSVStreamFromReader(t *testing.T) {
	content := "id,name,score,enabled,at\n1,alice,98.5,true,2026-05-16T10:00:00Z\n2,bob,77,false,2026-05-17T11:30:00Z\n"
	got := make([]personRow, 0, 2)
	err := ReadCSVStreamFromReader[personRow](strings.NewReader(content), func(rowNum int, item personRow) error {
		t.Logf("TestReadCSVStreamFromReader row=%d item=%+v", rowNum, item)
		got = append(got, item)
		return nil
	})
	if err != nil {
		t.Fatalf("read csv stream from reader failed: %v", err)
	}
	assertRows(t, got, sampleRows())
}

func TestWriteReadCSVWithTitleTag(t *testing.T) {
	dir := outputDir(t, "TestWriteReadCSVWithTitleTag")
	file := filepath.Join(dir, "title.csv")
	t.Logf("TestWriteReadCSVWithTitleTag file=%s", file)

	input := []personRowWithTitle{
		{ID: 1, Name: "alice", Score: 98.5, Enabled: true, At: time.Date(2026, 5, 16, 10, 0, 0, 0, time.UTC)},
	}
	if err := WriteCSV(file, input); err != nil {
		t.Fatalf("write csv failed: %v", err)
	}

	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read csv raw failed: %v", err)
	}
	t.Logf("TestWriteReadCSVWithTitleTag raw=%s", string(raw))
	if !strings.Contains(string(raw), "编号,姓名,得分,启用,时间") {
		t.Fatalf("title header not found in csv: %s", string(raw))
	}

	got, err := ReadCSV[personRowWithTitle](file)
	if err != nil {
		t.Fatalf("read csv failed: %v", err)
	}
	if len(got) != 1 || got[0].Name != "alice" || got[0].ID != 1 {
		t.Fatalf("unexpected parsed result: %+v", got)
	}
}

func TestWriteReadCSVMaps(t *testing.T) {
	dir := outputDir(t, "TestWriteReadCSVMaps")
	file := filepath.Join(dir, "maps.csv")
	t.Logf("TestWriteReadCSVMaps file=%s", file)

	rows := []map[string]any{
		{"id": 1, "name": "alice", "enabled": true},
		{"id": 2, "name": "bob", "enabled": false},
	}
	opts := sampleMapOptions()
	if err := WriteCSVMaps(file, rows, opts); err != nil {
		t.Fatalf("write csv maps failed: %v", err)
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read csv raw failed: %v", err)
	}
	t.Logf("TestWriteReadCSVMaps raw=%s", string(raw))
	if !strings.Contains(string(raw), "编号,姓名,启用") {
		t.Fatalf("custom title header not found: %s", string(raw))
	}

	got, err := ReadCSVMaps(file, opts)
	if err != nil {
		t.Fatalf("read csv maps failed: %v", err)
	}
	t.Logf("TestWriteReadCSVMaps output=%+v", got)
	if len(got) != 2 || got[0]["id"] != "1" || got[0]["name"] != "alice" || got[0]["enabled"] != "true" {
		t.Fatalf("unexpected parsed maps: %+v", got)
	}
}

func TestReadCSVMapsFromReader(t *testing.T) {
	content := "编号,姓名,启用\n1,alice,true\n"
	opts := MapOptions{
		TitleMap: map[string]string{
			"id":      "编号",
			"name":    "姓名",
			"enabled": "启用",
		},
	}
	got, err := ReadCSVMapsFromReader(strings.NewReader(content), opts)
	if err != nil {
		t.Fatalf("read csv maps from reader failed: %v", err)
	}
	t.Logf("TestReadCSVMapsFromReader output=%+v", got)
	if len(got) != 1 || got[0]["id"] != "1" || got[0]["name"] != "alice" {
		t.Fatalf("unexpected parsed maps: %+v", got)
	}
}

func TestWriteReadCSVMapsDefaultHeader(t *testing.T) {
	dir := outputDir(t, "TestWriteReadCSVMapsDefaultHeader")
	file := filepath.Join(dir, "maps-default.csv")
	t.Logf("TestWriteReadCSVMapsDefaultHeader file=%s", file)

	rows := []map[string]any{
		{"id": 1, "name": "alice"},
	}
	opts := MapOptions{FieldOrder: []string{"id", "name"}}
	if err := WriteCSVMaps(file, rows, opts); err != nil {
		t.Fatalf("write csv maps failed: %v", err)
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read csv raw failed: %v", err)
	}
	t.Logf("TestWriteReadCSVMapsDefaultHeader raw=%s", string(raw))
	if !strings.Contains(string(raw), "id,name") {
		t.Fatalf("default field header not found: %s", string(raw))
	}
	got, err := ReadCSVMaps(file, opts)
	if err != nil {
		t.Fatalf("read csv maps failed: %v", err)
	}
	if len(got) != 1 || got[0]["id"] != "1" || got[0]["name"] != "alice" {
		t.Fatalf("unexpected parsed maps: %+v", got)
	}
}

func TestWriteReadExcelMaps(t *testing.T) {
	dir := outputDir(t, "TestWriteReadExcelMaps")
	file := filepath.Join(dir, "maps.xlsx")
	t.Logf("TestWriteReadExcelMaps file=%s", file)

	rows := []map[string]any{
		{"id": 1, "name": "alice", "enabled": true},
	}
	opts := sampleMapOptions()
	if err := WriteExcelMaps(file, "People", rows, opts); err != nil {
		t.Fatalf("write excel maps failed: %v", err)
	}
	gotByName, err := ReadExcelMaps(file, SheetSelector{Name: "People"}, opts)
	if err != nil {
		t.Fatalf("read excel maps by name failed: %v", err)
	}
	t.Logf("TestWriteReadExcelMaps byName=%+v", gotByName)
	if len(gotByName) != 1 || gotByName[0]["id"] != "1" || gotByName[0]["name"] != "alice" {
		t.Fatalf("unexpected parsed maps by name: %+v", gotByName)
	}

	gotByIndex, err := ReadExcelMaps(file, SheetSelector{Index: 0}, opts)
	if err != nil {
		t.Fatalf("read excel maps by index failed: %v", err)
	}
	if len(gotByIndex) != 1 || gotByIndex[0]["enabled"] != "true" {
		t.Fatalf("unexpected parsed maps by index: %+v", gotByIndex)
	}
}

func TestReadExcelMapsFromReader(t *testing.T) {
	dir := outputDir(t, "TestReadExcelMapsFromReader")
	file := filepath.Join(dir, "maps-reader.xlsx")
	rows := []map[string]any{
		{"id": 1, "name": "alice", "enabled": true},
	}
	opts := sampleMapOptions()
	if err := WriteExcelMaps(file, "People", rows, opts); err != nil {
		t.Fatalf("write excel maps failed: %v", err)
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read excel raw failed: %v", err)
	}
	got, err := ReadExcelMapsFromReader(bytes.NewReader(raw), SheetSelector{Name: "People"}, opts)
	if err != nil {
		t.Fatalf("read excel maps from reader failed: %v", err)
	}
	t.Logf("TestReadExcelMapsFromReader output=%+v", got)
	if len(got) != 1 || got[0]["id"] != "1" || got[0]["enabled"] != "true" {
		t.Fatalf("unexpected parsed maps: %+v", got)
	}
}

func TestWriteReadExcelStreamStruct(t *testing.T) {
	dir := outputDir(t, "TestWriteReadExcelStreamStruct")
	file := filepath.Join(dir, "struct-stream.xlsx")
	t.Logf("TestWriteReadExcelStreamStruct file=%s", file)

	input := sampleRows()
	err := WriteExcelStream(file, "People", func(write func(personRow) error) error {
		for _, r := range input {
			if err := write(r); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("write excel struct stream failed: %v", err)
	}

	got := make([]personRow, 0, len(input))
	err = ReadExcelStream[personRow](file, SheetSelector{Name: "People"}, func(rowNum int, item personRow) error {
		t.Logf("excel struct stream row=%d item=%+v", rowNum, item)
		got = append(got, item)
		return nil
	})
	if err != nil {
		t.Fatalf("read excel struct stream failed: %v", err)
	}
	assertRows(t, got, input)
}

func TestReadExcelStreamFromReader(t *testing.T) {
	dir := outputDir(t, "TestReadExcelStreamFromReader")
	file := filepath.Join(dir, "struct-stream-reader.xlsx")
	input := sampleRows()
	err := WriteExcelStream(file, "People", func(write func(personRow) error) error {
		for _, r := range input {
			if err := write(r); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("write excel struct stream failed: %v", err)
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read excel raw failed: %v", err)
	}
	got := make([]personRow, 0, len(input))
	err = ReadExcelStreamFromReader[personRow](bytes.NewReader(raw), SheetSelector{Name: "People"}, func(rowNum int, item personRow) error {
		t.Logf("excel stream reader row=%d item=%+v", rowNum, item)
		got = append(got, item)
		return nil
	})
	if err != nil {
		t.Fatalf("read excel stream from reader failed: %v", err)
	}
	assertRows(t, got, input)
}

func TestWriteReadCSVMapsStream(t *testing.T) {
	dir := outputDir(t, "TestWriteReadCSVMapsStream")
	file := filepath.Join(dir, "maps-stream.csv")
	t.Logf("TestWriteReadCSVMapsStream file=%s", file)

	opts := sampleMapOptions()

	err := WriteCSVMapsStream(file, opts, writeSampleMapRows)
	if err != nil {
		t.Fatalf("write csv maps stream failed: %v", err)
	}

	got := make([]map[string]string, 0)
	err = ReadCSVMapsStream(file, opts, func(rowNum int, item map[string]string) error {
		t.Logf("csv stream row=%d item=%+v", rowNum, item)
		got = append(got, item)
		return nil
	})
	if err != nil {
		t.Fatalf("read csv maps stream failed: %v", err)
	}
	if len(got) != 2 || got[0]["id"] != "1" || got[1]["name"] != "bob" {
		t.Fatalf("unexpected csv stream maps: %+v", got)
	}
}

func TestReadCSVMapsStreamFromReader(t *testing.T) {
	content := "编号,姓名,启用\n1,alice,true\n2,bob,false\n"
	opts := sampleMapTitleOptions()
	got := make([]map[string]string, 0, 2)
	err := ReadCSVMapsStreamFromReader(strings.NewReader(content), opts, func(rowNum int, item map[string]string) error {
		t.Logf("csv maps stream reader row=%d item=%+v", rowNum, item)
		got = append(got, item)
		return nil
	})
	if err != nil {
		t.Fatalf("read csv maps stream from reader failed: %v", err)
	}
	if len(got) != 2 || got[1]["name"] != "bob" {
		t.Fatalf("unexpected csv maps stream: %+v", got)
	}
}

func TestWriteReadExcelMapsStream(t *testing.T) {
	dir := outputDir(t, "TestWriteReadExcelMapsStream")
	file := filepath.Join(dir, "maps-stream.xlsx")
	t.Logf("TestWriteReadExcelMapsStream file=%s", file)

	opts := sampleMapOptions()

	err := WriteExcelMapsStream(file, "People", opts, writeSampleMapRows)
	if err != nil {
		t.Fatalf("write excel maps stream failed: %v", err)
	}

	got := make([]map[string]string, 0)
	err = ReadExcelMapsStream(file, SheetSelector{Name: "People"}, opts, func(rowNum int, item map[string]string) error {
		t.Logf("excel stream row=%d item=%+v", rowNum, item)
		got = append(got, item)
		return nil
	})
	if err != nil {
		t.Fatalf("read excel maps stream failed: %v", err)
	}
	if len(got) != 2 || got[0]["id"] != "1" || got[1]["name"] != "bob" {
		t.Fatalf("unexpected excel stream maps: %+v", got)
	}
}

func TestReadExcelMapsStreamFromReader(t *testing.T) {
	dir := outputDir(t, "TestReadExcelMapsStreamFromReader")
	file := filepath.Join(dir, "maps-stream-reader.xlsx")
	opts := sampleMapOptions()
	err := WriteExcelMapsStream(file, "People", opts, writeSampleMapRows)
	if err != nil {
		t.Fatalf("write excel maps stream failed: %v", err)
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read excel raw failed: %v", err)
	}
	got := make([]map[string]string, 0, 2)
	err = ReadExcelMapsStreamFromReader(bytes.NewReader(raw), SheetSelector{Name: "People"}, opts, func(rowNum int, item map[string]string) error {
		t.Logf("excel maps stream reader row=%d item=%+v", rowNum, item)
		got = append(got, item)
		return nil
	})
	if err != nil {
		t.Fatalf("read excel maps stream from reader failed: %v", err)
	}
	if len(got) != 2 || got[1]["name"] != "bob" {
		t.Fatalf("unexpected excel maps stream: %+v", got)
	}
}

func assertRows(t *testing.T, got, want []personRow) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("size mismatch: got=%d want=%d", len(got), len(want))
	}
	for i := range want {
		if got[i].ID != want[i].ID ||
			got[i].Name != want[i].Name ||
			got[i].Score != want[i].Score ||
			got[i].Enabled != want[i].Enabled ||
			!got[i].At.Equal(want[i].At) {
			t.Fatalf("row mismatch at %d: got=%+v want=%+v", i, got[i], want[i])
		}
	}
}

func outputDir(t *testing.T, caseName string) string {
	t.Helper()
	// 全局开关为 true 时输出到 ~/Downloads，便于手工打开校验文件。
	if TestOutputToDownloads {
		home, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("resolve home dir failed: %v", err)
		}
		dir := filepath.Join(home, "Downloads", "infra-go-tabular-tests", caseName)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("create downloads output dir failed: %v", err)
		}
		return dir
	}
	return t.TempDir()
}

func sampleRows() []personRow {
	return []personRow{
		{ID: 1, Name: "alice", Score: 98.5, Enabled: true, At: time.Date(2026, 5, 16, 10, 0, 0, 0, time.UTC)},
		{ID: 2, Name: "bob", Score: 77.0, Enabled: false, At: time.Date(2026, 5, 17, 11, 30, 0, 0, time.UTC)},
	}
}

func sampleMapOptions() MapOptions {
	return MapOptions{
		FieldOrder: []string{"id", "name", "enabled"},
		TitleMap: map[string]string{
			"id":      "编号",
			"name":    "姓名",
			"enabled": "启用",
		},
	}
}

func sampleMapTitleOptions() MapOptions {
	return MapOptions{
		TitleMap: map[string]string{
			"id":      "编号",
			"name":    "姓名",
			"enabled": "启用",
		},
	}
}

func writeSampleMapRows(write func(map[string]any) error) error {
	if err := write(map[string]any{"id": 1, "name": "alice", "enabled": true}); err != nil {
		return err
	}
	return write(map[string]any{"id": 2, "name": "bob", "enabled": false})
}
