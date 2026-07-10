package logx

import "context"

type fieldsContextKey struct{}

// WithFields 将字段附加到 context，日志写入时会带入 Record.Fields。
func WithFields(ctx context.Context, fields map[string]string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if len(fields) == 0 {
		return ctx
	}
	merged := map[string]string{}
	if prev := FieldsFrom(ctx); len(prev) > 0 {
		for k, v := range prev {
			merged[k] = v
		}
	}
	for k, v := range fields {
		merged[k] = v
	}
	return context.WithValue(ctx, fieldsContextKey{}, merged)
}

// FieldsFrom 从 context 提取日志字段。
func FieldsFrom(ctx context.Context) map[string]string {
	if ctx == nil {
		return nil
	}
	v, _ := ctx.Value(fieldsContextKey{}).(map[string]string)
	if len(v) == 0 {
		return nil
	}
	out := make(map[string]string, len(v))
	for k, val := range v {
		out[k] = val
	}
	return out
}
