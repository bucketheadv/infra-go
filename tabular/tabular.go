package tabular

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bucketheadv/infra-go/timex"
	"github.com/xuri/excelize/v2"
)

// SheetSelector 用于选择 Excel sheet。
// Name 优先级高于 Index；Index 为 0-based，下标越界会返回错误。
type SheetSelector struct {
	// Name sheet 名称；非空时优先按名称选择。
	Name string
	// Index sheet 下标（0-based）；Name 为空时使用。
	Index int
}

type fieldMeta struct {
	// header 列标题。
	header string
	// index 结构体字段索引路径（支持嵌套）。
	index []int
	// typ 字段类型。
	typ reflect.Type
}

// MapOptions 用于 map 行数据读写配置。
type MapOptions struct {
	// FieldOrder 指定字段输出顺序；为空时：
	//   - 批量写：从全部行推导字段并集；
	//   - 流式写：仅根据第一行推导并锁定列（后续新字段会报错）。流式场景建议显式设置。
	FieldOrder []string
	// TitleMap 字段名 -> 标题名。若字段未配置标题，则默认使用字段名。
	TitleMap map[string]string
	// Strict 为 true 时（主要用于读路径）：
	//   - 读入要求非空 TitleMap，否则报错；
	//   - TitleMap 出现重复标题会报错；
	//   - 读入时重复列标题或映射到同一字段会报错；
	//   - 无法映射到 TitleMap 的标题会报错。
	Strict bool
}

// DecodeOptions 用于结构体行解码配置。
type DecodeOptions struct {
	// Strict 为 true 时，结构体字段对应标题缺失会返回错误。
	// 无论是否 Strict，表头出现重复列一律报错（避免静默错列）。
	Strict bool
}

// ReadExcel 读取 Excel 并按标题映射到结构体切片。
func ReadExcel[T any](path string, selector SheetSelector, opts ...DecodeOptions) ([]T, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	return ReadExcelFromReader[T](f, selector, opts...)
}

// ReadExcelFromReader 从 io.Reader 读取 Excel 并按标题映射到结构体切片。
func ReadExcelFromReader[T any](reader io.Reader, selector SheetSelector, opts ...DecodeOptions) ([]T, error) {
	rows, err := readExcelRowsFromReader(reader, selector)
	if err != nil {
		return nil, err
	}
	return decodeRows[T](rows, opts...)
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
	defer func() { _ = f.Close() }()
	defaultName := f.GetSheetName(f.GetActiveSheetIndex())
	if defaultName == "" {
		defaultName = "Sheet1"
	}
	if err := f.SetSheetName(defaultName, sheetName); err != nil {
		return err
	}

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
	return saveExcelAtomic(f, path)
}

// ReadCSV 读取 CSV 并按标题映射到结构体切片。
func ReadCSV[T any](path string, opts ...DecodeOptions) ([]T, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	return ReadCSVFromReader[T](f, opts...)
}

// ReadCSVFromReader 从 io.Reader 读取 CSV 并按标题映射到结构体切片。
func ReadCSVFromReader[T any](reader io.Reader, opts ...DecodeOptions) ([]T, error) {
	records, err := readCSVRecords(reader)
	if err != nil {
		return nil, err
	}
	return decodeRows[T](records, opts...)
}

// ReadCSVStream 流式读取 CSV，并按行回调结构体数据（适合大文件）。
// rowNum 为 CSV 行号（从 1 开始，含表头行；回调从第 2 行开始触发）。
func ReadCSVStream[T any](path string, handler func(rowNum int, item T) error, opts ...DecodeOptions) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	return ReadCSVStreamFromReader[T](f, handler, opts...)
}

// ReadCSVStreamFromReader 从 io.Reader 流式读取 CSV 并按行回调结构体数据。
func ReadCSVStreamFromReader[T any](reader io.Reader, handler func(rowNum int, item T) error, opts ...DecodeOptions) error {
	metas, err := buildFieldMetas[T]()
	if err != nil {
		return err
	}
	strict := firstDecodeOpt(opts).Strict
	headerMap := map[string]int(nil)
	return streamCSVRecords(
		reader,
		func(header []string) error {
			var err error
			headerMap, err = makeHeaderIndexMap(header)
			return err
		},
		func(rowNum int, record []string) error {
			item, err := decodeStructRecord[T](record, metas, headerMap, rowNum, strict)
			if err != nil {
				return err
			}
			return handler(rowNum, item)
		},
	)
}

// WriteCSV 将结构体切片写入 CSV，第一行为标题。
func WriteCSV[T any](path string, rows []T) error {
	return WriteCSVStream(path, func(write func(T) error) error {
		for _, row := range rows {
			if err := write(row); err != nil {
				return err
			}
		}
		return nil
	})
}

// WriteCSVStream 流式写入 CSV（适合大文件）。
// producer 通过 write 回调逐条产出结构体数据。
func WriteCSVStream[T any](path string, producer func(write func(T) error) error) error {
	metas, err := buildFieldMetas[T]()
	if err != nil {
		return err
	}
	return writeFileAtomic(path, func(tmpPath string) error {
		f, err := os.Create(tmpPath)
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

		writeRow := func(row T) error {
			rv := structValue(reflect.ValueOf(row))
			record := make([]string, 0, len(metas))
			for _, m := range metas {
				record = append(record, formatStringValue(rv.FieldByIndex(m.index)))
			}
			return w.Write(record)
		}

		if err := producer(writeRow); err != nil {
			return err
		}
		w.Flush()
		return w.Error()
	})
}

// WriteCSVMaps 将 []map[string]any 写入 CSV，第一行为标题。
func WriteCSVMaps(path string, rows []map[string]any, opts MapOptions) error {
	plan, err := buildMapWritePlan(rows, opts)
	if err != nil {
		return err
	}
	return writeFileAtomic(path, func(tmpPath string) error {
		f, err := os.Create(tmpPath)
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()

		w := csv.NewWriter(f)
		if err := w.Write(plan.headers); err != nil {
			return err
		}
		for _, row := range rows {
			if err := rejectUnexpectedMapKeys(plan.fields, row); err != nil {
				return err
			}
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
	})
}

// ReadCSVMaps 读取 CSV 并映射为 []map[string]string（key 为字段名）。
func ReadCSVMaps(path string, opts MapOptions) ([]map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	return ReadCSVMapsFromReader(f, opts)
}

// ReadCSVMapsFromReader 从 io.Reader 读取 CSV 并映射为 []map[string]string。
func ReadCSVMapsFromReader(reader io.Reader, opts MapOptions) ([]map[string]string, error) {
	records, err := readCSVRecords(reader)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return []map[string]string{}, nil
	}
	return decodeMapRows(records, opts)
}

// ReadCSVMapsStream 流式读取 CSV，并按行回调 map 数据（适合大文件）。
// rowNum 为 CSV 行号（从 1 开始，含表头行；回调从第 2 行开始触发）。
func ReadCSVMapsStream(path string, opts MapOptions, handler func(rowNum int, item map[string]string) error) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	return ReadCSVMapsStreamFromReader(f, opts, handler)
}

// ReadCSVMapsStreamFromReader 从 io.Reader 流式读取 CSV 并按行回调 map 数据。
func ReadCSVMapsStreamFromReader(reader io.Reader, opts MapOptions, handler func(rowNum int, item map[string]string) error) error {
	fields := []string(nil)
	return streamCSVRecords(
		reader,
		func(header []string) error {
			var err error
			fields, err = mapFieldsFromHeaders(header, opts)
			return err
		},
		func(rowNum int, record []string) error {
			return handler(rowNum, mapFromRecord(fields, record))
		},
	)
}

// WriteExcelMaps 将 []map[string]any 写入 Excel，第一行为标题。
func WriteExcelMaps(path, sheetName string, rows []map[string]any, opts MapOptions) error {
	plan, err := buildMapWritePlan(rows, opts)
	if err != nil {
		return err
	}
	if sheetName == "" {
		sheetName = "Sheet1"
	}
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()
	defaultName := f.GetSheetName(f.GetActiveSheetIndex())
	if defaultName == "" {
		defaultName = "Sheet1"
	}
	if err := f.SetSheetName(defaultName, sheetName); err != nil {
		return err
	}

	for i, h := range plan.headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		if err := f.SetCellValue(sheetName, cell, h); err != nil {
			return err
		}
	}
	for r, row := range rows {
		if err := rejectUnexpectedMapKeys(plan.fields, row); err != nil {
			return err
		}
		for c, field := range plan.fields {
			cell, _ := excelize.CoordinatesToCellName(c+1, r+2)
			if err := f.SetCellValue(sheetName, cell, mapValueToString(row[field])); err != nil {
				return err
			}
		}
	}
	return saveExcelAtomic(f, path)
}

// WriteCSVMapsStream 流式写入 CSV（适合大文件）。
// producer 通过 write 回调逐行产出数据。
// 表头在首次写入时锁定：若未指定 FieldOrder，则仅根据第一行推断列；
// 后续行出现未声明字段会返回错误（避免静默丢列）。建议显式设置 FieldOrder。
func WriteCSVMapsStream(path string, opts MapOptions, producer func(write func(map[string]any) error) error) error {
	if err := validateTitleMap(opts); err != nil {
		return err
	}
	return writeFileAtomic(path, func(tmpPath string) error {
		f, err := os.Create(tmpPath)
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()

		w := csv.NewWriter(f)
		headerWritten := false
		fields := append([]string{}, opts.FieldOrder...)

		writeHeader := func() error {
			headers := make([]string, 0, len(fields))
			for _, field := range fields {
				headers = append(headers, mapFieldHeader(field, opts.TitleMap))
			}
			if err := w.Write(headers); err != nil {
				return err
			}
			headerWritten = true
			return nil
		}

		if len(fields) > 0 {
			if err := writeHeader(); err != nil {
				return err
			}
		}

		writeRow := func(row map[string]any) error {
			if !headerWritten {
				fields = inferMapFields([]map[string]any{row})
				if err := writeHeader(); err != nil {
					return err
				}
			} else if err := rejectUnexpectedMapKeys(fields, row); err != nil {
				return err
			}
			record := make([]string, 0, len(fields))
			for _, field := range fields {
				record = append(record, mapValueToString(row[field]))
			}
			return w.Write(record)
		}

		if err := producer(writeRow); err != nil {
			return err
		}
		w.Flush()
		return w.Error()
	})
}

// ReadExcelMapsStream 流式读取 Excel，并按行回调 map 数据（适合大文件）。
// rowNum 为 sheet 行号（从 1 开始，回调从第 2 行开始触发）。
func ReadExcelMapsStream(path string, selector SheetSelector, opts MapOptions, handler func(rowNum int, item map[string]string) error) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	return ReadExcelMapsStreamFromReader(f, selector, opts, handler)
}

// ReadExcelMapsStreamFromReader 从 io.Reader 流式读取 Excel 并按行回调 map 数据。
func ReadExcelMapsStreamFromReader(reader io.Reader, selector SheetSelector, opts MapOptions, handler func(rowNum int, item map[string]string) error) error {
	fields := []string(nil)
	return streamExcelRows(reader, selector, func(rowNum int, cols []string) error {
		if rowNum == 1 {
			var err error
			fields, err = mapFieldsFromHeaders(cols, opts)
			return err
		}
		return handler(rowNum, mapFromRecord(fields, cols))
	})
}

// ReadExcelStream 流式读取 Excel，并按行回调结构体数据（适合大文件）。
// rowNum 为 sheet 行号（从 1 开始，回调从第 2 行开始触发）。
func ReadExcelStream[T any](path string, selector SheetSelector, handler func(rowNum int, item T) error, opts ...DecodeOptions) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	return ReadExcelStreamFromReader[T](f, selector, handler, opts...)
}

// ReadExcelStreamFromReader 从 io.Reader 流式读取 Excel 并按行回调结构体数据。
func ReadExcelStreamFromReader[T any](reader io.Reader, selector SheetSelector, handler func(rowNum int, item T) error, opts ...DecodeOptions) error {
	metas, err := buildFieldMetas[T]()
	if err != nil {
		return err
	}
	strict := firstDecodeOpt(opts).Strict
	headerMap := map[string]int(nil)
	return streamExcelRows(reader, selector, func(rowNum int, cols []string) error {
		if rowNum == 1 {
			var err error
			headerMap, err = makeHeaderIndexMap(cols)
			return err
		}
		item, err := decodeStructRecord[T](cols, metas, headerMap, rowNum, strict)
		if err != nil {
			return err
		}
		return handler(rowNum, item)
	})
}

// WriteExcelMapsStream 流式写入 Excel（适合大文件）。
// producer 通过 write 回调逐行产出数据。
// 表头在首次写入时锁定：若未指定 FieldOrder，则仅根据第一行推断列；
// 后续行出现未声明字段会返回错误（避免静默丢列）。建议显式设置 FieldOrder。
func WriteExcelMapsStream(path, sheetName string, opts MapOptions, producer func(write func(map[string]any) error) error) error {
	if err := validateTitleMap(opts); err != nil {
		return err
	}
	if sheetName == "" {
		sheetName = "Sheet1"
	}
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()
	defaultName := f.GetSheetName(f.GetActiveSheetIndex())
	if defaultName == "" {
		defaultName = "Sheet1"
	}
	if err := f.SetSheetName(defaultName, sheetName); err != nil {
		return err
	}

	sw, err := f.NewStreamWriter(sheetName)
	if err != nil {
		return err
	}

	headerWritten := false
	fields := append([]string{}, opts.FieldOrder...)
	rowIndex := 1

	writeHeader := func() error {
		headers := make([]any, 0, len(fields))
		for _, field := range fields {
			headers = append(headers, mapFieldHeader(field, opts.TitleMap))
		}
		cell, _ := excelize.CoordinatesToCellName(1, rowIndex)
		if err := sw.SetRow(cell, headers); err != nil {
			return err
		}
		headerWritten = true
		rowIndex++
		return nil
	}

	if len(fields) > 0 {
		if err := writeHeader(); err != nil {
			return err
		}
	}

	writeRow := func(row map[string]any) error {
		if !headerWritten {
			fields = inferMapFields([]map[string]any{row})
			if err := writeHeader(); err != nil {
				return err
			}
		} else if err := rejectUnexpectedMapKeys(fields, row); err != nil {
			return err
		}
		values := make([]any, 0, len(fields))
		for _, field := range fields {
			values = append(values, mapValueToString(row[field]))
		}
		cell, _ := excelize.CoordinatesToCellName(1, rowIndex)
		if err := sw.SetRow(cell, values); err != nil {
			return err
		}
		rowIndex++
		return nil
	}

	if err := producer(writeRow); err != nil {
		return err
	}
	if err := sw.Flush(); err != nil {
		return err
	}
	return saveExcelAtomic(f, path)
}

// WriteExcelStream 流式写入 Excel（适合大文件）。
// producer 通过 write 回调逐条产出结构体数据。
func WriteExcelStream[T any](path, sheetName string, producer func(write func(T) error) error) error {
	metas, err := buildFieldMetas[T]()
	if err != nil {
		return err
	}
	if sheetName == "" {
		sheetName = "Sheet1"
	}
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()
	defaultName := f.GetSheetName(f.GetActiveSheetIndex())
	if defaultName == "" {
		defaultName = "Sheet1"
	}
	if err := f.SetSheetName(defaultName, sheetName); err != nil {
		return err
	}

	sw, err := f.NewStreamWriter(sheetName)
	if err != nil {
		return err
	}

	headers := make([]any, 0, len(metas))
	for _, m := range metas {
		headers = append(headers, m.header)
	}
	startCell, _ := excelize.CoordinatesToCellName(1, 1)
	if err := sw.SetRow(startCell, headers); err != nil {
		return err
	}
	rowIndex := 2

	writeRow := func(row T) error {
		rv := structValue(reflect.ValueOf(row))
		values := make([]any, 0, len(metas))
		for _, m := range metas {
			values = append(values, formatCellValue(rv.FieldByIndex(m.index)))
		}
		cell, _ := excelize.CoordinatesToCellName(1, rowIndex)
		if err := sw.SetRow(cell, values); err != nil {
			return err
		}
		rowIndex++
		return nil
	}

	if err := producer(writeRow); err != nil {
		return err
	}
	if err := sw.Flush(); err != nil {
		return err
	}
	return saveExcelAtomic(f, path)
}

// ReadExcelMaps 读取 Excel 并映射为 []map[string]string（key 为字段名）。
func ReadExcelMaps(path string, selector SheetSelector, opts MapOptions) ([]map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	return ReadExcelMapsFromReader(f, selector, opts)
}

// ReadExcelMapsFromReader 从 io.Reader 读取 Excel 并映射为 []map[string]string。
func ReadExcelMapsFromReader(reader io.Reader, selector SheetSelector, opts MapOptions) ([]map[string]string, error) {
	rows, err := readExcelRowsFromReader(reader, selector)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return []map[string]string{}, nil
	}
	return decodeMapRows(rows, opts)
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

func readCSVRecords(reader io.Reader) ([][]string, error) {
	csvReader := csv.NewReader(reader)
	return csvReader.ReadAll()
}

func streamCSVRecords(
	reader io.Reader,
	onHeader func(header []string) error,
	onRecord func(rowNum int, record []string) error,
) error {
	csvReader := csv.NewReader(reader)
	headerRow, err := csvReader.Read()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}
	if err := onHeader(headerRow); err != nil {
		return err
	}

	rowNum := 1
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		rowNum++
		if err := onRecord(rowNum, record); err != nil {
			return err
		}
	}
}

func readExcelRowsFromReader(reader io.Reader, selector SheetSelector) ([][]string, error) {
	f, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	sheetName, err := resolveSheetName(f, selector)
	if err != nil {
		return nil, err
	}
	return f.GetRows(sheetName)
}

func streamExcelRows(reader io.Reader, selector SheetSelector, onRow func(rowNum int, cols []string) error) error {
	f, err := excelize.OpenReader(reader)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	sheetName, err := resolveSheetName(f, selector)
	if err != nil {
		return err
	}
	rows, err := f.Rows(sheetName)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()

	rowNum := 0
	for rows.Next() {
		rowNum++
		cols, err := rows.Columns()
		if err != nil {
			return err
		}
		if err := onRow(rowNum, cols); err != nil {
			return err
		}
	}
	return rows.Error()
}

func makeHeaderIndexMap(headers []string) (map[string]int, error) {
	headerMap := make(map[string]int, len(headers))
	for i, h := range headers {
		key := normalizeHeader(h)
		if prev, ok := headerMap[key]; ok {
			return nil, fmt.Errorf("duplicate header %q at columns %d and %d", h, prev+1, i+1)
		}
		headerMap[key] = i
	}
	return headerMap, nil
}

func decodeRows[T any](rows [][]string, opts ...DecodeOptions) ([]T, error) {
	metas, err := buildFieldMetas[T]()
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return []T{}, nil
	}

	strict := firstDecodeOpt(opts).Strict
	headerMap, err := makeHeaderIndexMap(rows[0])
	if err != nil {
		return nil, err
	}

	result := make([]T, 0, max(0, len(rows)-1))
	for i := 1; i < len(rows); i++ {
		item, err := decodeStructRecord[T](rows[i], metas, headerMap, i+1, strict)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, nil
}

func decodeStructRecord[T any](record []string, metas []fieldMeta, headerMap map[string]int, rowNum int, strict bool) (T, error) {
	var item T
	root := reflect.ValueOf(&item).Elem()
	// 仅当目标类型本身是指针时分配外层；字段级指针保持 nil，直到有非空单元格再分配。
	rv, err := ensureRootStruct(root)
	if err != nil {
		return item, err
	}
	for _, m := range metas {
		idx, ok := headerMap[normalizeHeader(m.header)]
		if !ok || idx >= len(record) {
			if strict {
				return item, fmt.Errorf("row %d: missing column %q", rowNum, m.header)
			}
			continue
		}
		raw := strings.TrimSpace(record[idx])
		if raw == "" {
			continue
		}
		if err := assignFromString(rv.FieldByIndex(m.index), raw); err != nil {
			return item, fmt.Errorf("row %d col %q: %w", rowNum, m.header, err)
		}
	}
	return item, nil
}

// ensureRootStruct 返回可写的结构体 Value；若 root 是 nil 指针则分配一层。
func ensureRootStruct(root reflect.Value) (reflect.Value, error) {
	v := root
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("target type must be struct")
	}
	return v, nil
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
			// 写路径只读字段：nil 指针不分配，避免污染调用方或对不可寻址值 Set panic。
			return reflect.Zero(v.Type().Elem())
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
		timex.DateTimeCommon,
		timex.DateOnly,
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
	// fields 输出字段名顺序。
	fields []string
	// headers 对应列标题。
	headers []string
}

func buildMapWritePlan(rows []map[string]any, opts MapOptions) (mapWritePlan, error) {
	if err := validateTitleMap(opts); err != nil {
		return mapWritePlan{}, err
	}
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
	}, nil
}

func validateTitleMap(opts MapOptions) error {
	if !opts.Strict || len(opts.TitleMap) == 0 {
		return nil
	}
	seen := make(map[string]string, len(opts.TitleMap))
	for field, title := range opts.TitleMap {
		key := normalizeHeader(title)
		if key == "" {
			continue
		}
		if prev, ok := seen[key]; ok && prev != field {
			return fmt.Errorf("duplicate title %q mapped by fields %q and %q", title, prev, field)
		}
		seen[key] = field
	}
	return nil
}

func firstDecodeOpt(opts []DecodeOptions) DecodeOptions {
	if len(opts) == 0 {
		return DecodeOptions{}
	}
	return opts[0]
}

func writeFileAtomic(path string, write func(tmpPath string) error) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tabular-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	_ = tmp.Close()
	defer func() { _ = os.Remove(tmpPath) }()
	if err := write(tmpPath); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

func saveExcelAtomic(f *excelize.File, path string) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tabular-*.xlsx")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	_ = tmp.Close()
	defer func() { _ = os.Remove(tmpPath) }()
	if err := f.SaveAs(tmpPath); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
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

// rejectUnexpectedMapKeys 在流式写表头已锁定后，拒绝后续行中的未知字段，避免静默丢列。
func rejectUnexpectedMapKeys(fields []string, row map[string]any) error {
	if len(row) == 0 {
		return nil
	}
	known := make(map[string]struct{}, len(fields))
	for _, f := range fields {
		known[f] = struct{}{}
	}
	for k := range row {
		if _, ok := known[k]; !ok {
			return fmt.Errorf("map write: unexpected field %q not in output columns; include it in FieldOrder or remove the key", k)
		}
	}
	return nil
}

func mapFieldHeader(field string, titleMap map[string]string) string {
	if titleMap != nil {
		if t := strings.TrimSpace(titleMap[field]); t != "" {
			return t
		}
	}
	return field
}

func decodeMapRows(rows [][]string, opts MapOptions) ([]map[string]string, error) {
	headers := rows[0]
	fields, err := mapFieldsFromHeaders(headers, opts)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]string, 0, len(rows)-1)
	for i := 1; i < len(rows); i++ {
		item := mapFromRecord(fields, rows[i])
		out = append(out, item)
	}
	return out, nil
}

func mapFieldsFromHeaders(headers []string, opts MapOptions) ([]string, error) {
	if opts.Strict && len(opts.TitleMap) == 0 {
		return nil, fmt.Errorf("MapOptions.Strict requires non-empty TitleMap")
	}
	if err := validateTitleMap(opts); err != nil {
		return nil, err
	}
	fields := make([]string, 0, len(headers))
	seen := make(map[string]int, len(headers))
	for i, h := range headers {
		field, err := resolveMapFieldByHeader(h, opts)
		if err != nil {
			return nil, err
		}
		if prev, ok := seen[field]; ok {
			return nil, fmt.Errorf("duplicate field %q from headers at columns %d and %d", field, prev+1, i+1)
		}
		seen[field] = i
		fields = append(fields, field)
	}
	return fields, nil
}

func mapFromRecord(fields []string, record []string) map[string]string {
	item := make(map[string]string, len(fields))
	for c, f := range fields {
		if c < len(record) {
			item[f] = strings.TrimSpace(record[c])
		} else {
			item[f] = ""
		}
	}
	return item
}

func resolveMapFieldByHeader(header string, opts MapOptions) (string, error) {
	h := normalizeHeader(header)
	matched := ""
	for field, title := range opts.TitleMap {
		if normalizeHeader(title) == h {
			if matched != "" && matched != field {
				return "", fmt.Errorf("ambiguous title %q matches fields %q and %q", header, matched, field)
			}
			matched = field
		}
	}
	if matched != "" {
		return matched, nil
	}
	if opts.Strict {
		return "", fmt.Errorf("unmapped header %q", header)
	}
	return strings.TrimSpace(header), nil
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
