package logx

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type yamlRoot struct {
	CallerFileMaxLen int                      `yaml:"callerFileMaxLen"`
	Appenders        map[string]yamlAppender  `yaml:"appenders"`
	Root             yamlLoggerDef            `yaml:"root"`
	Loggers          map[string]yamlLoggerDef `yaml:"loggers"`
}

type yamlAppender struct {
	Type            string            `yaml:"type"`
	Colored         bool              `yaml:"colored"`
	Layout          string            `yaml:"layout"`
	Pattern         string            `yaml:"pattern"`
	Path            string            `yaml:"path"`
	MaxLinesPerFile int               `yaml:"maxLinesPerFile"`
	RetentionDays   int               `yaml:"retentionDays"`
	LevelColors     map[string]string `yaml:"levelColors"`
	FieldColors     map[string]string `yaml:"fieldColors"`
}

type yamlLoggerDef struct {
	Level     string   `yaml:"level"`
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
	for name, a := range cfg {
		switch a.Type {
		case "console":
			lc := parseLevelColors(a.LevelColors)
			fc := normalizeFieldColorKeys(a.FieldColors)
			out[name] = newConsoleAppender(a.Layout, a.Pattern, a.Colored, lc, fc, callerMax)
		case "rollingFile":
			if a.Path == "" {
				return nil, fmt.Errorf("appender %q: rollingFile 需要 path", name)
			}
			maxL := a.MaxLinesPerFile
			if maxL == 0 {
				maxL = 100_000
			}
			lc := parseLevelColors(a.LevelColors)
			fc := normalizeFieldColorKeys(a.FieldColors)
			rf, err := newRollingFileAppender(a.Path, maxL, a.RetentionDays, a.Layout, a.Pattern, a.Colored, lc, fc, callerMax)
			if err != nil {
				return nil, fmt.Errorf("appender %q: %w", name, err)
			}
			out[name] = rf
		default:
			return nil, fmt.Errorf("appender %q: 未知 type %q", name, a.Type)
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
