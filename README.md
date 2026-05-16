# infra-go

可复用的 Go 基础库（独立于业务项目）。

## applog

`import "github.com/bucketheadv/infra-go/applog"`

完整用法见 **[docs/applog.md](docs/applog.md)**。仓库根目录 **[applog.example.yaml](applog.example.yaml)** 为配置示例。

在业务仓库的 `go.mod` 中发布前可使用本地替换：

```text
require github.com/bucketheadv/infra-go v0.1.0
replace github.com/bucketheadv/infra-go => /绝对路径/infra-go
```

发布到 GitHub 后删除 `replace` 行并 `go get` 指定 tag 即可。

## timezone

`import "github.com/bucketheadv/infra-go/timezone"`

用于 UTC 偏移时区查询与时间换算，例如：

- `timezone.GetTimeZone("UTC+08:00")`
- `timezone.WithZone(t, zone)`
- `timezone.WithZoneRetainFields(t, zone)`

## basic

`import "github.com/bucketheadv/infra-go/basic"`

用于基础类型转换与通用元组，例如：

- `basic.StringTo[int]("123")`
- `basic.ArrayElemTo[bool]([]string{"true","false"})`
- `basic.Pair[int,string]`
- `basic.Triple[int,string,bool]`

## collection

`import "github.com/bucketheadv/infra-go/collection"`

用于常见集合处理，例如：

- `collection.Partition(arr, size)`
- `collection.GroupBy(arr, keyFn)`
- `collection.ArrayToMap(arr, coverExists, keyFn)`
- `collection.SortedMapTraversal(m, reverse, fn)`

## version

`import "github.com/bucketheadv/infra-go/version"`

用于版本号解析和比较（支持 `1.2`、`1.2.30`、`1.2.3.40`、`1.2.30-beta`），例如：

- `version.Compare("1.2", "1.3.0") // -1`
- `version.Compare("1.2", "1.1.99") // 1`
- `version.Compare("1.2.3.40", "1.2.3.5") // 1`
- `version.Compare("1.2.30-beta", "1.2.30") // 1`
