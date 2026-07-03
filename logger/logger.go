package logger

import "fmt"

func Log(msgs ...any) {
	fmt.Println(msgs...)
}

func Error(msgs ...any) {
	Log(append([]any{" ❌ | Error:"}, msgs...)...)
}

func Warning(msgs ...any) {
	Log(append([]any{" ⚠️  | Warning:"}, msgs...)...)
}

func Debug(msgs ...any) {
	Log(append([]any{" 🔍 | Debug:"}, msgs...)...)
}

func Trace(msgs ...any) {
	Log(append([]any{" 📝 | Trace:"}, msgs...)...)
}
