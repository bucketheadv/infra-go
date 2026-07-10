package logx

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type yamlRoot struct {
	// CallerFileMaxLen 调用方文件路径截断最大长度。
	CallerFileMaxLen int `yaml:"callerFileMaxLen"`
	// Appenders 按名称定义的输出器。
	Appenders map[string]yamlAppender `yaml:"appenders"`
	// Root 根 logger 配置。
	Root yamlLoggerDef `yaml:"root"`
	// Loggers 命名 logger 配置。
	Loggers map[string]yamlLoggerDef `yaml:"loggers"`
}

type yamlAppender struct {
	// Type 输出器类型（console / rollingFile）。
	Type string `yaml:"type"`
	// Colored 是否启用 ANSI 着色。
	Colored bool `yaml:"colored"`
	// Layout 布局类型（text / json）。
	Layout string `yaml:"layout"`
	// Pattern 文本布局 pattern。
	Pattern string `yaml:"pattern"`
	// Path 滚动文件路径（rollingFile 必填）。
	Path string `yaml:"path"`
	// MaxLinesPerFile 单文件最大行数。
	MaxLinesPerFile int `yaml:"maxLinesPerFile"`
	// RetentionDays 历史文件保留天数。
	RetentionDays int `yaml:"retentionDays"`
	// LevelColors 级别名到颜色的映射。
	LevelColors map[string]string `yaml:"levelColors"`
	// FieldColors 字段名到颜色的映射。
	FieldColors map[string]string `yaml:"fieldColors"`
}

type yamlLoggerDef struct {
	// Level 最低输出级别。
	Level string `yaml:"level"`
	// Appenders 绑定的输出器名称列表。
	Appenders []string `yaml:"appenders"`
}

func loadYAMLConfig(path string) (*yamlRoot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg yamlRoot
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func buildAppenders(
	cfg map[string]yamlAppender,
	callerMax int,
) (map[string]Appender, error) {
	out := make(map[string]Appender)
	fail := func(err error) (map[string]Appender, error) {
		closeAppenderPool(out)
		return nil, err
	}
	for name, a := range cfg {
		switch a.Type {
		case "console":
			lc, err := parseLevelColors(a.LevelColors)
			if err != nil {
				return fail(fmt.Errorf("appender %q: %w", name, err))
			}
			fc := normalizeFieldColorKeys(a.FieldColors)
			out[name] = newConsoleAppender(a.Layout, a.Pattern, a.Colored, lc, fc, callerMax)
		case "rollingFile":
			if a.Path == "" {
				return fail(fmt.Errorf("appender %q: rollingFile 需要 path", name))
			}
			maxL := a.MaxLinesPerFile
			if maxL == 0 {
				maxL = 100_000
			}
			lc, err := parseLevelColors(a.LevelColors)
			if err != nil {
				return fail(fmt.Errorf("appender %q: %w", name, err))
			}
			fc := normalizeFieldColorKeys(a.FieldColors)
			rf, err := newRollingFileAppender(a.Path, maxL, a.RetentionDays, a.Layout, a.Pattern, a.Colored, lc, fc, callerMax)
			if err != nil {
				return fail(fmt.Errorf("appender %q: %w", name, err))
			}
			out[name] = rf
		default:
			return fail(fmt.Errorf("appender %q: 未知 type %q", name, a.Type))
		}
	}
	return out, nil
}

func resolveAppenderRefs(names []string, pool map[string]Appender) ([]Appender, error) {
	if len(names) == 0 {
		return nil, nil
	}
	var list []Appender
	for _, ref := range names {
		a, ok := pool[ref]
		if !ok {
			return nil, fmt.Errorf("未找到 appender %q", ref)
		}
		list = append(list, a)
	}
	return list, nil
}
