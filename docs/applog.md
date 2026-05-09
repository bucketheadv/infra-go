# applog 使用说明

`applog` 提供基于 YAML 的应用日志：多 logger、类 Spring 的 `pattern` 与 `%clr` 着色、按行滚动与保留天数、以及 GORM SQL 桥接。

## 1. 安装

在业务项目的 `go.mod` 中：

```text
require github.com/bucketheadv/infra-go v0.1.0
```

若库尚未发布或本地调试，可增加 `replace`（路径改为你的本机目录）：

```text
replace github.com/bucketheadv/infra-go => /path/to/infra-go
```

代码中导入：

```go
import "github.com/bucketheadv/infra-go/applog"
```

执行 `go mod tidy`。

> 若你 fork 后修改了模块路径，需与 `infra-go/go.mod` 的 `module` 及库内 `applog/const.go` 中的前缀保持一致。

## 2. 初始化

在进程早期（读取配置之后）加载 YAML，失败时 `MustLoad` 会 panic：

```go
applog.MustLoad("config/applog.yaml")
```

或使用 `Load` 自行处理错误：

```go
if err := applog.Load("config/applog.yaml"); err != nil {
    log.Fatal(err)
}
```

未 `Load` 前会使用内置的简单控制台输出；`Load` 成功后按 YAML 替换全局注册表。

### 2.1 配置文件路径有什么要求？怎么「导入」？

- **没有特殊格式要求**：传入 `Load` / `MustLoad` 的是普通**文件路径字符串**，库内部用 `os.ReadFile` 读取；路径合法且进程有读权限即可。
- **相对路径**：相对于进程启动时的 **当前工作目录（cwd）**（你在哪个目录执行 `./app`，或 IDE「工作目录」填的是什么，就相对于那里）。例如 `config/applog.yaml` 表示 `cwd/config/applog.yaml`。
- **绝对路径**：始终推荐在部署/容器里使用，例如 `/etc/myapp/applog.yaml`，避免随 cwd 变化找不到文件。
- **不是 Go 的 `import`**：YAML 不会在编译期打进包名；「导入配置」就是在启动早期调用一次 `applog.MustLoad(path)`（或 `Load`）。若要用 `//go:embed` 把 YAML 嵌进二进制，需自行把内容落到可读路径，或封装一层在启动时 `os.WriteFile` 临时文件再 `Load`（本库当前仅提供按路径加载）。
- **YAML 里的 `rollingFile.path`**：同样是操作系统路径；写相对路径时也是相对于 **cwd**，与配置文件本身在哪无关。需要日志落在可执行文件旁时，应用代码里应用 `filepath.Join` 拼出绝对路径写进配置，或在业务层生成 YAML / 多套配置。

推荐写法示例（需 `import ("os"; "path/filepath")`，配置目录由你统一约定）：

```go
cfgDir := os.Getenv("APP_CONFIG_DIR") // 或 flag、spring 配置等
if cfgDir == "" {
    cfgDir = "config"
}
applog.MustLoad(filepath.Join(cfgDir, "applog.yaml"))
```

## 3. 配置文件

可参考仓库根目录的 **`applog.example.yaml`**，复制为业务项目中的路径（例如 `config/applog.yaml`）。

主要结构：

| 段 | 含义 |
|----|------|
| `callerFileMaxLen` | 日志里 `file:line` 最大长度，超出左侧省略为 `...` |
| `appenders` | 输出目的地： `console`、`rollingFile` |
| `root` | 根 logger：默认级别与 appender 列表 |
| `loggers` | 命名 logger，可覆盖级别与 appender；未写 `appenders` 时可继承示例中的写法 |

Appender 类型：

- **`console`**：`layout` 可选 `text` | `pattern` | `json`；`pattern` 与 Spring 风格占位符、`%clr(子模式){颜色}` 见下文。
- **`rollingFile`**：`path` 当前文件；`maxLinesPerFile` 单文件最大行数；`retentionDays` 轮转文件保留天数（0 表示不清理）。

## 4. 占位符与颜色（`layout: pattern`）

常用占位符（`%%` 为字面量 `%`）：

| 占位符 | 说明 |
|--------|------|
| `%d` / `%date` | 时间；可 `%d{2006-01-02 15:04:05.000}`（Go 参考时间） |
| `%level` / `%p` | 级别大写 |
| `%fileLine` / `%F` | 源码位置 |
| `%logger` / `%c` | logger 名称 |
| `%pid` | 进程 PID |
| `%msg` / `%m` | 消息正文 |
| `%n` | 换行 |

- **`fieldColors`**：为未包在 `%clr` 内的字段单独着色（键可用 `date`、`level`、`fileLine`、`logger`、`msg`、`pid` 等别名）。
- **`levelColors`**：当未配置 `fieldColors.level` 时，按日志级别为 `%level` 着色。
- **`%clr(子模式){颜色}`**：子模式展开后整体套色；颜色可为名称、`faint`/`bold`、SGR 数字等。

## 5. 代码里打日志

按**命名 logger**（与 YAML 里 `loggers` 的 key 对应）：

```go
applog.Infof(ctx, applog.NameApp, "hello %s", name)
applog.Errorf(ctx, applog.NameApp, "err: %v", err)
```

预置名称常量：`NameRoot`、`NameApp`、`NameAccess`、`NameGorm`（可按需在业务里自行定义字符串）。

也可先取实例再调方法：

```go
applog.Get("app").Infof(ctx, "msg")
```

**指定调用位置**（例如由其它框架传入 file/line）：

```go
applog.LogFrom(ctx, "app", applog.LevelInfo, file, line, msg)
```

进程退出前建议关闭滚动文件等：

```go
applog.Close()
```

## 6. GORM

将 `*gorm.DB` 的 logger 换成 `applog`（需在 YAML 中配置 `loggers.gorm` 或使用 root 的 appender）：

```go
import (
    glog "gorm.io/gorm/logger"
    // ...
)

db.Config.Logger = applog.NewGormLogger(applog.GormLoggerConfig{
    LoggerName:                applog.NameGorm,
    SlowThreshold:             200 * time.Millisecond,
    IgnoreRecordNotFoundError: true,
}).LogMode(glog.Info)
```

- GORM 的 `LogMode`（如 `Info` / `Warn` / `Error`）控制**是否以及何种 SQL**会回调到本库。
- YAML 里 **`loggers.gorm.level`** 控制 applog **是否输出**对应级别；两者需一起考虑（例如要看普通成功 SQL，GORM 一般为 `Info`，applog 里 `gorm` 也应允许 `info`）。

SQL 会合并为单行：`[耗时ms] [rows:n] SQL...`，`file:line` 为业务里发起查询的栈位置（已跳过 GORM 与本库）。

## 7. 注意事项

- 配置文件与 `rollingFile.path` 的解析规则见 **2.1 节**。
- 与 **go-spring** 等框架并存时，框架自带的 `log.yaml` 与 **`applog` 的 YAML** 是两套配置，互不影响。
- 发布本库后删除业务项目 `go.mod` 中的 `replace`，改用 `go get github.com/bucketheadv/infra-go@vx.y.z`。
