package sample

import (
	"context"
	"fmt"
	"quote/common/log"
	"testing"
)

func TestContextLog(t *testing.T) {
	ctx := context.TODO()
	newLogger := log.WithFields("trace_id", "1234567890")
	//将logger设置到context里面，后续调用链可以使用context一路降log传递下去
	//特别适合携带trace id
	newContext := log.ContextWithLogger(ctx, newLogger)
	fmt.Println(newContext.Value("CONTEXT_LOG_KEY"))
	log.InfoContext(newContext, "日志携带trace id")
	f1(newContext)
}

func f1(ctx context.Context) {
	log.InfoContext(ctx, "第一层")
	f2(ctx)
}

func f2(ctx context.Context) {
	log.InfoContext(ctx, "第二层")
}
