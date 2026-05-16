package tabular

import (
	"encoding/csv"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// SheetSelector 用于选择 Excel sheet。
// Name 优先级高于 Index；Index 为 0-based，下标越界会返回错误。
type SheetSelector struct {
	Name  string
	Index int
}

type fieldMeta struct {
	header string
	index  []int
	typ    reflect.Type
}

// MapOptions 用于 map 行数据读写配置。
type MapOptions struct {
	// FieldOrder 指定字段输出顺序；为空时自动从数据中推导并按字典序排序。
	FieldOrder []string
	// TitleMap 字段名 -> 标题名。若字段未配置标题，则默认使用字段名。
	TitleMap map[string]string
}

// ReadExcel 读取 Excel 并按标题映射到结构体切片。
func ReadExcel[T any](path string, selector SheetSelector) ([]T, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	sheetName, err := resolveSheetName(f, selector)
	if err != nil {
		return nil, err
	}
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, err
	}
	return decodeRows[T](rows)
}

// WriteExcel 将结构体切片写入 Excel，第一行为标题。
// sheetName 为空时使用默认值 "Sheet1"。
func WriteExcel[T any](path, sheetName string, rows []T) error {
	metas, err := buildFieldMetas[T]()
	if err != nil {
		return err
	}
	if sheetName == "" {
		sheetName = "Sheet1"
	}

	f := excelize.NewFile()
	defaultName := f.GetSheetName(f.GetActiveSheetIndex())
	if defaultName == "" {
		defaultName = "Sheet1"
	}
	_ = f.SetSheetName(defaultName, sheetName)

	for i, m := range metas {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		if err := f.SetCellValue(sheetName, cell, m.header); err != nil {
			return err
		}
	}
	for r, row := range rows {
		rv := structValue(reflect.ValueOf(row))
		for c, m := range metas {
			cell, _ := excelize.CoordinatesToCellName(c+1, r+2)
			v := rv.FieldByIndex(m.index)
			if err := f.SetCellValue(sheetName, cell, formatCellValue(v)); err != nil {
				return err
			}
		}
	}
	return f.SaveAs(path)
}

// ReadCSV 读取 CSV 并按标题映射到结构体切片。
func ReadCSV[T any](path string) ([]T, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	rows := make([][]string, 0, len(records))
	for _, r := range records {
		rows = append(rows, r)
	}
	return decodeRows[T](rows)
}

// WriteCSV 将结构体切片写入 CSV，第一行为标题。
func WriteCSV[T any](path string, rows []T) error {
	metas, err := buildFieldMetas[T]()
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	w := csv.NewWriter(f)
	headers := make([]string, 0, len(metas))
	for _, m := range metas {
		headers = append(headers, m.header)
	}
	if err := w.Write(headers); err != nil {
		return err
	}

	for _, row := range rows {
		rv := structValue(reflect.ValueOf(row))
		record := make([]string, 0, len(metas))
		for _, m := range metas {
			record = append(record, formatStringValue(rv.FieldByIndex(m.index)))
		}
		if err := w.Write(record); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

// WriteCSVMaps 将 []map[string]any 写入 CSV，第一行为标题。
func WriteCSVMaps(path string, rows []map[string]any, opts MapOptions) error {
	plan := buildMapWritePlan(rows, opts)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	w := csv.NewWriter(f)
	if err := w.Write(plan.headers); err != nil {
		return err
	}
	for _, row := range rows {
		record := make([]string, 0, len(plan.fields))
		for _, field := range plan.fields {
			record = append(record, mapValueToString(row[field]))
		}
		if err := w.Write(record); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

// ReadCSVMaps 读取 CSV 并映射为 []map[string]string（key 为字段名）。
func ReadCSVMaps(path string, opts MapOptions) ([]map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return []map[string]string{}, nil
	}
	return decodeMapRows(records, opts), nil
}

// WriteExcelMaps 将 []map[string]any 写入 Excel，第一行为标题。
func WriteExcelMaps(path, sheetName string, rows []map[string]any, opts MapOptions) error {
	plan := buildMapWritePlan(rows, opts)
	if sheetName == "" {
		sheetName = "Sheet1"
	}
	f := excelize.NewFile()
	defaultName := f.GetSheetName(f.GetActiveSheetIndex())
	if defaultName == "" {
		defaultName = "Sheet1"
	}
	_ = f.SetSheetName(defaultName, sheetName)

	for i, h := range plan.headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		if err := f.SetCellValue(sheetName, cell, h); err != nil {
			return err
		}
	}
	for r, row := range rows {
		for c, field := range plan.fields {
			cell, _ := excelize.CoordinatesToCellName(c+1, r+2)
			if err := f.SetCellValue(sheetName, cell, mapValueToString(row[field])); err != nil {
				return err
			}
		}
	}
	return f.SaveAs(path)
}

// ReadExcelMaps 读取 Excel 并映射为 []map[string]string（key 为字段名）。
func ReadExcelMaps(path string, selector SheetSelector, opts MapOptions) ([]map[string]string, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	sheetName, err := resolveSheetName(f, selector)
	if err != nil {
		return nil, err
	}
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return []map[string]string{}, nil
	}
	return decodeMapRows(rows, opts), nil
}

func resolveSheetName(f *excelize.File, selector SheetSelector) (string, error) {
	if strings.TrimSpace(selector.Name) != "" {
		name := strings.TrimSpace(selector.Name)
		idx, err := f.GetSheetIndex(name)
		if err != nil || idx < 0 {
			return "", fmt.Errorf("sheet %q not found", name)
		}
		return name, nil
	}

	names := f.GetSheetList()
	if len(names) == 0 {
		return "", fmt.Errorf("excel has no sheet")
	}
	if selector.Index < 0 {
		return names[0], nil
	}
	if selector.Index >= len(names) {
		return "", fmt.Errorf("sheet index out of range: %d", selector.Index)
	}
	return names[selector.Index], nil
}

func decodeRows[T any](rows [][]string) ([]T, error) {
	metas, err := buildFieldMetas[T]()
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return []T{}, nil
	}

	headerMap := make(map[string]int, len(rows[0]))
	for i, h := range rows[0] {
		headerMap[normalizeHeader(h)] = i
	}

	result := make([]T, 0, max(0, len(rows)-1))
	for i := 1; i < len(rows); i++ {
		var item T
		rv := structValue(reflect.ValueOf(&item).Elem())
		for _, m := range metas {
			idx, ok := headerMap[normalizeHeader(m.header)]
			if !ok || idx >= len(rows[i]) {
				continue
			}
			raw := strings.TrimSpace(rows[i][idx])
			if raw == "" {
				continue
			}
			if err := assignFromString(rv.FieldByIndex(m.index), raw); err != nil {
				return nil, fmt.Errorf("row %d col %q: %w", i+1, m.header, err)
			}
		}
		result = append(result, item)
	}
	return result, nil
}

func buildFieldMetas[T any]() ([]fieldMeta, error) {
	var zero T
	t := reflect.TypeOf(zero)
	if t == nil {
		return nil, fmt.Errorf("unsupported type")
	}
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("target type must be struct")
	}

	metas := make([]fieldMeta, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue
		}
		name, skip := fieldHeaderName(f)
		if skip || name == "" {
			continue
		}
		metas = append(metas, fieldMeta{
			header: name,
			index:  f.Index,
			typ:    f.Type,
		})
	}
	if len(metas) == 0 {
		return nil, fmt.Errorf("struct has no exported fields")
	}
	return metas, nil
}

func fieldHeaderName(f reflect.StructField) (string, bool) {
	if name, ok := parseTagName(f.Tag.Get("title")); ok {
		return name, false
	}
	if name, ok := parseTagName(f.Tag.Get("header")); ok {
		return name, false
	}
	if name, ok := parseTagName(f.Tag.Get("excel")); ok {
		return name, false
	}
	if name, ok := parseTagName(f.Tag.Get("csv")); ok {
		return name, false
	}
	if name, ok := parseTagName(f.Tag.Get("json")); ok {
		return name, false
	}
	return f.Name, false
}

func parseTagName(tag string) (string, bool) {
	if tag == "" {
		return "", false
	}
	name := strings.Split(tag, ",")[0]
	if name == "-" {
		return "", true
	}
	if name == "" {
		return "", false
	}
	return name, true
}

func structValue(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	return v
}

func normalizeHeader(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func assignFromString(dst reflect.Value, raw string) error {
	for dst.Kind() == reflect.Pointer {
		if dst.IsNil() {
			dst.Set(reflect.New(dst.Type().Elem()))
		}
		dst = dst.Elem()
	}
	if dst.Type() == reflect.TypeOf(time.Time{}) {
		t, err := parseTime(raw)
		if err != nil {
			return err
		}
		dst.Set(reflect.ValueOf(t))
		return nil
	}
	switch dst.Kind() {
	case reflect.String:
		dst.SetString(raw)
	case reflect.Bool:
		v, err := strconv.ParseBool(raw)
		if err != nil {
			return err
		}
		dst.SetBool(v)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(raw, 10, dst.Type().Bits())
		if err != nil {
			return err
		}
		dst.SetInt(v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		v, err := strconv.ParseUint(raw, 10, dst.Type().Bits())
		if err != nil {
			return err
		}
		dst.SetUint(v)
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(raw, dst.Type().Bits())
		if err != nil {
			return err
		}
		dst.SetFloat(v)
	default:
		return fmt.Errorf("unsupported field type: %s", dst.Type().String())
	}
	return nil
}

func parseTime(raw string) (time.Time, error) {
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, l := range layouts {
		if t, err := time.Parse(l, raw); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid time format: %q", raw)
}

func formatCellValue(v reflect.Value) any {
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}
	if !v.IsValid() {
		return ""
	}
	if t, ok := v.Interface().(time.Time); ok {
		return t.Format(time.RFC3339)
	}
	return v.Interface()
}

func formatStringValue(v reflect.Value) string {
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}
	if !v.IsValid() {
		return ""
	}
	val := formatCellValue(v)
	if val == nil {
		return ""
	}
	switch x := val.(type) {
	case string:
		return x
	case bool:
		return strconv.FormatBool(x)
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", x)
	case uint, uint8, uint16, uint32, uint64, uintptr:
		return fmt.Sprintf("%d", x)
	case float32, float64:
		return fmt.Sprintf("%v", x)
	default:
		return fmt.Sprint(x)
	}
}

type mapWritePlan struct {
	fields  []string
	headers []string
}

func buildMapWritePlan(rows []map[string]any, opts MapOptions) mapWritePlan {
	fields := make([]string, 0)
	if len(opts.FieldOrder) > 0 {
		fields = append(fields, opts.FieldOrder...)
	} else {
		fields = inferMapFields(rows)
	}
	headers := make([]string, 0, len(fields))
	for _, f := range fields {
		headers = append(headers, mapFieldHeader(f, opts.TitleMap))
	}
	return mapWritePlan{
		fields:  fields,
		headers: headers,
	}
}

func inferMapFields(rows []map[string]any) []string {
	seen := make(map[string]struct{})
	for _, row := range rows {
		for k := range row {
			seen[k] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func mapFieldHeader(field string, titleMap map[string]string) string {
	if titleMap != nil {
		if t := strings.TrimSpace(titleMap[field]); t != "" {
			return t
		}
	}
	return field
}

func decodeMapRows(rows [][]string, opts MapOptions) []map[string]string {
	headers := rows[0]
	fields := make([]string, 0, len(headers))
	for _, h := range headers {
		fields = append(fields, resolveMapFieldByHeader(h, opts.TitleMap))
	}
	out := make([]map[string]string, 0, len(rows)-1)
	for i := 1; i < len(rows); i++ {
		item := make(map[string]string, len(fields))
		for c, f := range fields {
			if c < len(rows[i]) {
				item[f] = rows[i][c]
			} else {
				item[f] = ""
			}
		}
		out = append(out, item)
	}
	return out
}

func resolveMapFieldByHeader(header string, titleMap map[string]string) string {
	h := normalizeHeader(header)
	for field, title := range titleMap {
		if normalizeHeader(title) == h {
			return field
		}
	}
	return strings.TrimSpace(header)
}

func mapValueToString(v any) string {
	if v == nil {
		return ""
	}
	switch x := v.(type) {
	case string:
		return x
	case bool:
		return strconv.FormatBool(x)
	case time.Time:
		return x.Format(time.RFC3339)
	default:
		return fmt.Sprint(x)
	}
}
