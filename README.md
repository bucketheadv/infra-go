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

用法参考 joda-time 设计（含方法命名），例如：

- `timezone.ForID("Asia/Shanghai")`
- `timezone.ForID("UTC+08")`
- `timezone.GetDefault()` / `timezone.SetDefault(...)`
- `timezone.UTC()`
- `timezone.WithZone(t, zone)` / `timezone.WithZoneRetainFields(t, zone)`
- `timezone.ForOffsetHours(8)`
- 常用时区常量：`timezone.IDAsiaShanghai`、`timezone.IDAmericaNewYork`
