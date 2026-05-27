package logx

// modulePrefix 与 go.mod 的 module + "/logx" 一致，用于 runtime 栈中识别本包（调用方解析、GORM 桥接跳过）。
const modulePrefix = "github.com/bucketheadv/infra-go/logx"
