# infra-go

可复用的 Go 基础库（独立于业务项目）。

## logx

`import "github.com/bucketheadv/infra-go/logx"`

完整用法见 **[docs/logx.md](docs/logx.md)**。仓库根目录 **[logx.example.yaml](logx.example.yaml)** 为配置示例。

- `logx.Load(path)` / `logx.Get(name)` / `logx.Has(name)` / `logx.MustGet(name)`
- `logx.WithFields(ctx, map[string]string{"traceId": "..."})`
- `Fatalf` 写完 FATAL 后会 `os.Exit(1)`

## timex

`import "github.com/bucketheadv/infra-go/timex"`

- `timex.ParseAny` / `timex.FormatAny` / `timex.NowIn`
- `timex.StartOfDay` / `StartOfWeek` / `StartOfYear` / `DaysBetween`
- `timex.FormatRelative` / `RegisterRelativeLocale`（默认回退语言：英文）
- `timex.FormatDuration` / `RegisterDurationLocale`

## timezone

`import "github.com/bucketheadv/infra-go/timezone"`

- `timezone.GetTimeZone("UTC+08:00")`
- `timezone.Lookup("Asia/Shanghai")`（固定偏移 + IANA）
- `timezone.WithZone` / `WithZoneRetainFields`

## typex

`import "github.com/bucketheadv/infra-go/typex"`

- `typex.StringTo[int]("123")` / `typex.ToString(123)`
- `typex.Ptr` / `Deref` / `Coalesce` / `CoalescePtr` / `Must` / `If`
- `typex.Pair` / `Triple`

## stringx

`import "github.com/bucketheadv/infra-go/stringx"`

- `IsEmpty` / `IsBlank` / `DefaultIfBlank` / `SplitTrim` / `JoinNonEmpty`
- `SnakeCase` / `CamelCase` / `LowerCamelCase` / `KebabCase`
- `PadLeft` / `PadRight` / `Ellipsis` / `Truncate`

## collectionx

`import "github.com/bucketheadv/infra-go/collectionx"`

- `Map` / `Filter` / `FilterMap` / `Unique` / `DistinctBy` / `GroupBy` / `ArrayToMap`
- `Intersect` / `Union` / `Difference` / `Shuffle` / `Shuffled`
- `Keys` / `Values` / `MergeMaps` / `Pick` / `Omit`

## numx

`import "github.com/bucketheadv/infra-go/numx"`

- `numx.Clamp` / `numx.InRange`（数值区间，区别于 `timex.InRange`）
- `numx.Round` / `Percent` / `ApproximatelyEqual`

## retryx

`import "github.com/bucketheadv/infra-go/retryx"`

- `retryx.Do(ctx, cfg, fn)`
- `retryx.Fixed` / `Exponential` / `ExponentialWithJitter`

## versionx

`import "github.com/bucketheadv/infra-go/versionx"`

支持 `1.2`、`1.2.30`、`1.2.3.40`、`1.2.30-beta`。预发布版本小于正式版（SemVer 风格）：

- `versionx.Compare("1.2.30-beta", "1.2.30") // -1`
- `versionx.Less` / `Greater` / `Equal`
- `v.Major()` / `Minor()` / `Patch()` / `Build()` / `Suffix()`

## trigger

`import "github.com/bucketheadv/infra-go/trigger"`

- `trigger.NextTriggerTimes` / `NextTriggerTime`
- `trigger.ValidateSpec`

## tabular

`import "github.com/bucketheadv/infra-go/tabular"`

Excel/CSV 与结构体 / map 互转；写入采用临时文件 + rename。

- `tabular.ReadCSV[Row](path)` / `ReadExcel[Row](path, selector)`
- `tabular.ReadCSV[Row](path, tabular.DecodeOptions{Strict: true})`
- `tabular.WriteCSV` / `WriteExcel` / Stream / Maps 系列

## jsonx

`import "github.com/bucketheadv/infra-go/jsonx"`

增强 JSON 反序列化：数字字符串自动转数值，支持 `time.Time` 与嵌入 struct：

- `jsonx.Unmarshal([]byte(\`{"id":"123","at":"2026-05-16T10:00:00Z"}\`), &obj)`

## 本地替换

```text
require github.com/bucketheadv/infra-go v0.1.0
replace github.com/bucketheadv/infra-go => /绝对路径/infra-go
```
