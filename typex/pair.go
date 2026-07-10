package typex

// Pair 表示左右两个值组成的二元组。
type Pair[T any, R any] struct {
	// Left 左侧值。
	Left T `json:"left"`
	// Right 右侧值。
	Right R `json:"right"`
}

// Triple 表示左中右三个值组成的三元组。
type Triple[T any, M any, R any] struct {
	// Left 左侧值。
	Left T `json:"left"`
	// Middle 中间值。
	Middle M `json:"middle"`
	// Right 右侧值。
	Right R `json:"right"`
}
