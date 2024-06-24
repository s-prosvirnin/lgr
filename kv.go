package main

import (
	"fmt"
	"time"
)

type KVParamType int

const (
	KVParamTypeString KVParamType = iota
	KVParamTypeInt
	KVParamTypeFloat
	KVParamTypeBoolean
	KVParamTypeTime
	KVParamTypeDuration
)

type KVParam struct {
	key    string
	pType  KVParamType
	string string
	int    int64
	float  float64
}

func (p KVParam) Key() string {
	return p.key
}

func (p KVParam) Type() KVParamType {
	return p.pType
}

func (p KVParam) String() string {
	return p.string
}

func (p KVParam) Bool() bool {
	return p.Int() != 0
}

func (p KVParam) Int() int {
	return int(p.int)
}

func (p KVParam) Float() float64 {
	return p.float
}

func (p KVParam) Time() time.Time {
	return time.Unix(0, p.int)
}

func (p KVParam) Duration() time.Duration {
	return time.Nanosecond * time.Duration(p.int)
}

func KVString(key, value string) KVParam {
	return KVParam{key: key, pType: KVParamTypeString, string: value}
}

func KVSprintf(key, format string, args ...interface{}) KVParam {
	return KVParam{key: key, pType: KVParamTypeString, string: fmt.Sprintf(format, args...)}
}

func KVInt(key string, value int) KVParam {
	return KVParam{key: key, pType: KVParamTypeInt, int: int64(value)}
}

func KVFloat64(key string, value float64) KVParam {
	return KVParam{key: key, pType: KVParamTypeFloat, float: value}
}

func KVBool(key string, value bool) KVParam {
	param := KVParam{key: key, pType: KVParamTypeBoolean, int: 0}
	if value {
		param.int = 1
	}

	return param
}

func KVTime(key string, value time.Time) KVParam {
	return KVParam{key: key, pType: KVParamTypeTime, int: value.UnixNano()}
}

func KVDuration(key string, value time.Duration) KVParam {
	return KVParam{key: key, pType: KVParamTypeDuration, int: value.Nanoseconds()}
}
