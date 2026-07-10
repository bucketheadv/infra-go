package logx

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// Registry 保存全局日志配置与已命名的 Logger。
type Registry struct {
	// mu 保护注册表并发读写。
	mu sync.RWMutex
	// inFlight 跟踪进行中的日志写入，便于优雅关闭。
	inFlight sync.WaitGroup
	// callerFileMaxLen 调用方文件路径截断最大长度。
	callerFileMaxLen int
	// appenders root 使用的输出器列表。
	appenders []Appender
	// appenderPool 按名称索引的输出器池。
	appenderPool map[string]Appender
	// loggers 按名称索引的 Logger。
	loggers map[string]*Logger
	// root 根 logger。
	root *Logger
}

var defaultRegistry atomic.Pointer[Registry]

func init() {
	defaultRegistry.Store(newBootstrapRegistry())
}

func newBootstrapRegistry() *Registry {
	r := &Registry{loggers: make(map[string]*Logger)}
	c := newConsoleAppender("text", "", true, nil, nil, 48)
	r.appenders = []Appender{c}
	r.root = &Logger{name: NameRoot, level: LevelInfo, appenders: []Appender{c}, reg: r}
	r.loggers[NameRoot] = r.root
	return r
}

// Load 从 YAML 文件加载配置并替换默认注册表。
func Load(path string) error {
	cfg, err := loadYAMLConfig(path)
	if err != nil {
		return err
	}
	callerMax := cfg.CallerFileMaxLen
	if callerMax <= 0 {
		callerMax = 48
	}
	pool, err := buildAppenders(cfg.Appenders, callerMax)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			closeAppenderPool(pool)
		}
	}()

	rootLevel, err := parseLevel(cfg.Root.Level)
	if err != nil {
		return fmt.Errorf("root: %w", err)
	}
	rootApps, err := resolveAppenderRefs(cfg.Root.Appenders, pool)
	if err != nil {
		return fmt.Errorf("root: %w", err)
	}
	if len(rootApps) == 0 {
		rootApps = []Appender{newConsoleAppender("text", "", true, nil, nil, callerMax)}
	}

	reg := &Registry{
		callerFileMaxLen: callerMax,
		appenderPool:     pool,
		loggers:          make(map[string]*Logger),
	}
	var all []Appender
	for _, a := range pool {
		all = append(all, a)
	}
	reg.appenders = all

	rootLogger := &Logger{
		name:      NameRoot,
		level:     rootLevel,
		appenders: rootApps,
		reg:       reg,
	}
	reg.root = rootLogger
	reg.loggers[NameRoot] = rootLogger

	for lname, def := range cfg.Loggers {
		lv := rootLevel
		if def.Level != "" {
			parsed, err := parseLevel(def.Level)
			if err != nil {
				return fmt.Errorf("logger %q: %w", lname, err)
			}
			lv = parsed
		}
		apps, err := resolveAppenderRefs(def.Appenders, pool)
		if err != nil {
			return fmt.Errorf("logger %q: %w", lname, err)
		}
		if len(apps) == 0 {
			apps = rootApps
		}
		reg.loggers[lname] = &Logger{
			name:      lname,
			level:     lv,
			appenders: apps,
			reg:       reg,
		}
	}

	replaceRegistry(reg)
	committed = true
	return nil
}

func closeAppenderPool(pool map[string]Appender) {
	for _, a := range pool {
		_ = a.Close()
	}
}

// MustLoad 等价于 Load，失败时 panic。
func MustLoad(path string) {
	if err := Load(path); err != nil {
		panic("logx: " + err.Error())
	}
}

func (r *Registry) closeAppenders() {
	if r == nil {
		return
	}
	for _, a := range r.appenders {
		if err := a.Close(); err != nil {
			reportAppendError("close appender", err)
		}
	}
}

// flushRegistry 将当前注册表中可刷盘的 appender 同步到磁盘（Fatalf 前调用）。
func flushRegistry(reg *Registry) {
	if reg == nil {
		return
	}
	reg.mu.RLock()
	apps := append([]Appender(nil), reg.appenders...)
	reg.mu.RUnlock()
	for _, a := range apps {
		if rf, ok := a.(*rollingFileAppender); ok {
			rf.syncFile()
		}
	}
}

// replaceRegistry 原子切换注册表，并等待旧注册表上的 in-flight 写入结束后再关闭 appender。
func replaceRegistry(next *Registry) {
	prev := defaultRegistry.Swap(next)
	prev.inFlight.Wait()
	prev.closeAppenders()
}

// Close 关闭当前注册表下所有 appender，并重置为可继续写控制台的引导注册表。
// 关闭后继续打日志不会静默丢失，而是写入新的默认 console appender。
func Close() {
	replaceRegistry(newBootstrapRegistry())
}

func lookupLogger(reg *Registry, name string) *Logger {
	reg.mu.RLock()
	defer reg.mu.RUnlock()
	if l, ok := reg.loggers[name]; ok {
		return l
	}
	return reg.root
}

// Get 按名称取 Logger；未配置时回退到 root。
// 每次调用都从当前注册表解析，Load 后旧 *Logger 通过 name 仍可写到新配置。
func Get(name string) *Logger {
	return lookupLogger(defaultRegistry.Load(), name)
}

// Has 判断指定名称的 Logger 是否已配置（不含 root 回退）。
func Has(name string) bool {
	reg := defaultRegistry.Load()
	reg.mu.RLock()
	defer reg.mu.RUnlock()
	_, ok := reg.loggers[name]
	return ok
}

// MustGet 按名称取 Logger；未配置时 panic。
func MustGet(name string) *Logger {
	reg := defaultRegistry.Load()
	reg.mu.RLock()
	defer reg.mu.RUnlock()
	l, ok := reg.loggers[name]
	if !ok {
		panic("logx: logger not found: " + name)
	}
	return l
}
