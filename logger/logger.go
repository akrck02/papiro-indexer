package logger

import (
	"fmt"
)

func Log(msgs ...any) {
	fmt.Println(msgs...)
}

func Error(msgs ...any) {
	Log(append([]any{" ❌ | Error:"}, msgs...)...)
}

func Warning(msgs ...any) {
	Log(append([]any{" ⚠️  | Warning:"}, msgs...)...)
}
