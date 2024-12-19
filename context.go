package main

import "context"

func CtxWithValue(ctx context.Context, params ...KVParam) context.Context {
	oldParams := getContextParams(ctx)

	return context.WithValue(ctx, contextParamsKey, append(oldParams, params...))
}

func getContextParams(ctx context.Context) []KVParam {
	fields, ok := ctx.Value(contextParamsKey).([]KVParam)
	if !ok {
		fields = make([]KVParam, 0)
	}

	return fields
}
