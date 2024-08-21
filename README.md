``` go
package main

import (
	"context"
	"log/slog"
	"os"
	"time"
)

func main() {
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
	lgr.Info(ctx, "test1", KVInt("key2", 222), KVInt("key1", 11111))
	lgr.Info(ctx, "test2", KVBool("key2", false), KVString("key1", "val1"))
	ctx = lgr.CtxWithParams(ctx, KVInt("key3", 333))
	lgr.Info(ctx, "test3")

	err = NewErrFromMsg("err1").WithCtx(ctx).WithType(ErrType("errtype1")).WithParams(
		KVTime("t1", time.Now().Add(-time.Minute*10000)),
	)
	err = NewErr(err).WithMsgWrap("err2").WithCtx(
		lgr.CtxWithParams(
			context.Background(),
			KVDuration("dur1", time.Minute),
		),
	)
	lgr.Error(ctx, err)
}
```
