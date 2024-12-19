``` go
package main

import (
	"context"
	"log/slog"
	"os"
	"time"
)

func main() {
	myLogger()
}

func myLogger() {
	lgr, err := initLogger(
		loggerConfig{
			Level:    logLevelDebug,
			Encoding: logTypeJson,
		},
	)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	ctx = lgr.CtxWithParams(ctx, KVInt("key1", 111))
	lgr.Info(ctx, "msg1", KVInt("key2", 222), KVInt("key1", 11111))
	ctx = CtxWithValue(ctx, KVInt("key2", 221))
	lgr.Info(ctx, "msg2", KVBool("key2", false), KVString("key1", "val1"))
	ctx = lgr.CtxWithParams(ctx, KVInt("key3", 333))
	lgr.Info(ctx, "msg3")

	errType1 := ErrType("errtype1")

	err = NewErrFromMsg("err1").WithCtx(ctx).WithType(errType1).WithParams(
		KVTime("t1", time.Now().Add(-time.Minute*10000)),
	)
	err = NewErr(err).WithMsgWrap("err2").WithCtx(
		lgr.CtxWithParams(
			context.Background(),
			KVDuration("dur1", time.Minute),
		),
	)
	lgr.Error(ctx, err)

	for err = errors.Unwrap(err); err != nil; err = errors.Unwrap(err) {
		fmt.Println(err)
	}
}
```
