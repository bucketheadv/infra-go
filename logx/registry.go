package logx

import (
	"fmt"
	"sync"
)

// Registry 保存全局日志配置与已命名的 Logger。
type Registry struct {
	mu               sync.RWMutex
	callerFileMaxLen int
	appenders        []Appender
	appenderPool     map[string]Appender
	loggers          map[string]*Logger
	root             *Logger
}

var defaultRegistry = newBootstrapRegistry()

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
	rootLevel := parseLevel(cfg.Root.Level)
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
		lv := parseLevel(def.Level)
		if def.Level == "" {
			lv = rootLevel
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

	prev := defaultRegistry
	prev.mu.Lock()
	defaultRegistry = reg
	prev.mu.Unlock()

	prev.closeAppenders()
	return nil
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
		_ = a.Close()
	}
}

// Close 关闭当前注册表下所有 appender（如滚动日志刷盘）。
func Close() {
	defaultRegistry.mu.Lock()
	reg := defaultRegistry
	defaultRegistry.mu.Unlock()
	reg.closeAppenders()
}

// Get 按名称取 Logger；未配置时回退到 root。
func Get(name string) *Logger {
	defaultRegistry.mu.RLock()
	defer defaultRegistry.mu.RUnlock()
	if l, ok := defaultRegistry.loggers[name]; ok {
		return l
	}
	return defaultRegistry.root
}
