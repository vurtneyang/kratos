package xredis

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"kratos/pkg/cache/redis"
)

type Cmder interface {
	Name() string
	Args() []interface{}

	setReply(interface{}, error)

	setErr(error)
	Err() error
}

type baseCmd struct {
	ctx  context.Context
	args []interface{}
	err  error
}

func (cmd *baseCmd) Name() string {
	if len(cmd.args) == 0 {
		return ""
	}
	// Cmd name must be lower cased.
	return strings.ToLower(cmd.stringArg(0))
}

func (cmd *baseCmd) Args() []interface{} {
	return cmd.args
}

func (cmd *baseCmd) stringArg(pos int) string {
	if pos < 0 || pos >= len(cmd.args) {
		return ""
	}
	s, _ := cmd.args[pos].(string)
	return s
}

func (cmd *baseCmd) setErr(e error) {
	cmd.err = e
}

func (cmd *baseCmd) Err() error {
	return cmd.err
}

//------------------------------------------------------------------------------

type Cmd struct {
	baseCmd

	val interface{}
}

var _ Cmder = (*Cmd)(nil)

func NewCmd(ctx context.Context, args ...interface{}) *Cmd {
	return &Cmd{
		baseCmd: baseCmd{
			ctx:  ctx,
			args: args,
		},
	}
}

func (cmd *Cmd) Val() interface{} {
	return cmd.val
}

func (cmd *Cmd) Result() (interface{}, error) {
	return cmd.val, cmd.err
}

func (cmd *Cmd) StringValue() (string, error) { //Text()
	return redis.String(cmd.val, cmd.err)
}

func (cmd *Cmd) Int() (int, error) {
	return redis.Int(cmd.val, cmd.err)
}

func (cmd *Cmd) Int64() (int64, error) {
	return redis.Int64(cmd.val, cmd.err)
}

func (cmd *Cmd) Uint64() (uint64, error) {
	return redis.Uint64(cmd.val, cmd.err)
}

// func (cmd *Cmd) Float32() (float32, error) {
// }

func (cmd *Cmd) Float64() (float64, error) {
	return redis.Float64(cmd.val, cmd.err)
}

func (cmd *Cmd) Bool() (bool, error) {
	return redis.Bool(cmd.val, cmd.err)
}

func (cmd *Cmd) setReply(val interface{}, err error) {
	cmd.val, cmd.err = val, err
}

//------------------------------------------------------------------------------

type SliceCmd struct {
	baseCmd

	val []interface{}
}

var _ Cmder = (*SliceCmd)(nil)

func NewSliceCmd(ctx context.Context, args ...interface{}) *SliceCmd {
	return &SliceCmd{
		baseCmd: baseCmd{
			ctx:  ctx,
			args: args,
		},
	}
}

func (cmd *SliceCmd) Val() []interface{} {
	return cmd.val
}

func (cmd *SliceCmd) Result() ([]interface{}, error) {
	return cmd.val, cmd.err
}

func (cmd *SliceCmd) setReply(val interface{}, err error) {
	cmd.val, cmd.err = redis.Values(val, err)
}

func (cmd *SliceCmd) Scan(dest ...interface{}) ([]interface{}, error) {
	if cmd.err != nil {
		return nil, cmd.err
	}
	return redis.Scan(cmd.val, dest...)
}

func (cmd *SliceCmd) ScanSlice(src []interface{}, dest interface{}, fieldNames ...string) error {
	if cmd.err != nil {
		return cmd.err
	}
	return redis.ScanSlice(src, dest, fieldNames...)
}

//------------------------------------------------------------------------------

type StringStringMapCmd struct {
	baseCmd

	val []interface{}
}

var _ Cmder = (*StringStringMapCmd)(nil)

func NewStringStringMapCmd(ctx context.Context, args ...interface{}) *StringStringMapCmd {
	return &StringStringMapCmd{
		baseCmd: baseCmd{
			ctx:  ctx,
			args: args,
		},
	}
}

// func (cmd *StringStringMapCmd) Val() []interface{} {
// 	return cmd.val
// }

func (cmd *StringStringMapCmd) Result() (map[string]string, error) {
	return redis.StringMap(cmd.val, cmd.err)
}

func (cmd *StringStringMapCmd) setReply(val interface{}, err error) {
	cmd.val, cmd.err = redis.Values(val, err)
}

func (cmd *StringStringMapCmd) ScanStruct(dest interface{}) error {
	if cmd.err != nil {
		return cmd.err
	}
	return redis.ScanStruct(cmd.val, dest)
}

//------------------------------------------------------------------------------

type StatusCmd struct {
	baseCmd

	val string
}

var _ Cmder = (*StatusCmd)(nil)

func NewStatusCmd(ctx context.Context, args ...interface{}) *StatusCmd {
	return &StatusCmd{
		baseCmd: baseCmd{
			ctx:  ctx,
			args: args,
		},
	}
}

func (cmd *StatusCmd) Val() string {
	return cmd.val
}

func (cmd *StatusCmd) Result() (string, error) {
	return cmd.val, cmd.err
}

func (cmd *StatusCmd) setReply(val interface{}, err error) {
	cmd.val, cmd.err = redis.String(val, err)
}

//------------------------------------------------------------------------------

type IntCmd struct {
	baseCmd

	val int64
}

var _ Cmder = (*IntCmd)(nil)

func NewIntCmd(ctx context.Context, args ...interface{}) *IntCmd {
	return &IntCmd{
		baseCmd: baseCmd{
			ctx:  ctx,
			args: args,
		},
	}
}

func (cmd *IntCmd) Val() int64 {
	return cmd.val
}

func (cmd *IntCmd) Result() (int64, error) {
	return cmd.val, cmd.err
}

func (cmd *IntCmd) Uint64() (uint64, error) {
	return uint64(cmd.val), cmd.err
}

func (cmd *IntCmd) setReply(val interface{}, err error) {
	cmd.val, cmd.err = redis.Int64(val, err)
}

//------------------------------------------------------------------------------

type DurationCmd struct {
	baseCmd

	val       time.Duration
	precision time.Duration
}

var _ Cmder = (*DurationCmd)(nil)

func NewDurationCmd(ctx context.Context, precision time.Duration, args ...interface{}) *DurationCmd {
	return &DurationCmd{
		baseCmd: baseCmd{
			ctx:  ctx,
			args: args,
		},
		precision: precision,
	}
}

func (cmd *DurationCmd) Val() time.Duration {
	return cmd.val
}

func (cmd *DurationCmd) Result() (time.Duration, error) {
	return cmd.val, cmd.err
}

func (cmd *DurationCmd) setReply(val interface{}, err error) {
	var n int64
	n, cmd.err = redis.Int64(val, err)
	if cmd.err != nil {
		return
	}
	switch n {
	// -2 if the key does not exist
	// -1 if the key exists but has no associated expire
	case -2, -1:
		cmd.val = time.Duration(n)
	default:
		cmd.val = time.Duration(n) * cmd.precision
	}
}

//------------------------------------------------------------------------------

type BoolCmd struct {
	baseCmd

	val bool
}

var _ Cmder = (*BoolCmd)(nil)

func NewBoolCmd(ctx context.Context, args ...interface{}) *BoolCmd {
	return &BoolCmd{
		baseCmd: baseCmd{
			ctx:  ctx,
			args: args,
		},
	}
}

func (cmd *BoolCmd) Val() bool {
	return cmd.val
}

func (cmd *BoolCmd) Result() (bool, error) {
	return cmd.val, cmd.err
}

func (cmd *BoolCmd) setReply(val interface{}, err error) {
	cmd.err = err
	if cmd.err != nil {
		return
	}

	switch v := val.(type) {
	case int64:
		cmd.val = v == 1
	case string:
		cmd.val = v == "OK"
	case nil:
		// `SET key value NX` returns nil when key already exists. But
		// `SETNX key value` returns bool (0/1). So convert nil to bool.
		cmd.val = false
	default:
		cmd.err = fmt.Errorf("got %T, wanted int64 or string", v)
	}
}

//------------------------------------------------------------------------------

type StringCmd struct {
	baseCmd

	val string
}

var _ Cmder = (*StringCmd)(nil)

func NewStringCmd(ctx context.Context, args ...interface{}) *StringCmd {
	return &StringCmd{
		baseCmd: baseCmd{
			ctx:  ctx,
			args: args,
		},
	}
}

func (cmd *StringCmd) Val() string {
	return cmd.val
}

func (cmd *StringCmd) Result() (string, error) {
	return cmd.val, cmd.err
}

func (cmd *StringCmd) Int() (int, error) {
	if cmd.err != nil {
		return 0, cmd.err
	}
	return strconv.Atoi(cmd.val)
}

func (cmd *StringCmd) Int64() (int64, error) {
	if cmd.err != nil {
		return 0, cmd.err
	}
	return strconv.ParseInt(cmd.val, 10, 64)
}

func (cmd *StringCmd) Uint64() (uint64, error) {
	if cmd.err != nil {
		return 0, cmd.err
	}
	return strconv.ParseUint(cmd.val, 10, 64)
}

func (cmd *StringCmd) Float32() (float32, error) {
	if cmd.err != nil {
		return 0, cmd.err
	}
	f, err := strconv.ParseFloat(cmd.val, 32)
	if err != nil {
		return 0, err
	}
	return float32(f), nil
}

func (cmd *StringCmd) Float64() (float64, error) {
	if cmd.err != nil {
		return 0, cmd.err
	}
	return strconv.ParseFloat(cmd.val, 64)
}

func (cmd *StringCmd) Time() (time.Time, error) {
	if cmd.err != nil {
		return time.Time{}, cmd.err
	}
	return time.Parse(time.RFC3339Nano, cmd.val)
}

func (cmd *StringCmd) setReply(val interface{}, err error) {
	cmd.val, cmd.err = redis.String(val, err)
}

//------------------------------------------------------------------------------

type FloatCmd struct {
	baseCmd

	val float64
}

var _ Cmder = (*FloatCmd)(nil)

func NewFloatCmd(ctx context.Context, args ...interface{}) *FloatCmd {
	return &FloatCmd{
		baseCmd: baseCmd{
			ctx:  ctx,
			args: args,
		},
	}
}

func (cmd *FloatCmd) Val() float64 {
	return cmd.val
}

func (cmd *FloatCmd) Result() (float64, error) {
	return cmd.val, cmd.err
}

func (cmd *FloatCmd) setReply(val interface{}, err error) {
	cmd.val, cmd.err = redis.Float64(val, err)
}

//------------------------------------------------------------------------------

type StringSliceCmd struct {
	baseCmd

	val []string
}

var _ Cmder = (*StringSliceCmd)(nil)

func NewStringSliceCmd(ctx context.Context, args ...interface{}) *StringSliceCmd {
	return &StringSliceCmd{
		baseCmd: baseCmd{
			ctx:  ctx,
			args: args,
		},
	}
}

func (cmd *StringSliceCmd) Val() []string {
	return cmd.val
}

func (cmd *StringSliceCmd) Result() ([]string, error) {
	return cmd.val, cmd.err
}

func (cmd *StringSliceCmd) setReply(val interface{}, err error) {
	cmd.val, cmd.err = redis.Strings(val, err)
}

//------------------------------------------------------------------------------

type ZSliceCmd struct {
	baseCmd

	val []Z
}

var _ Cmder = (*ZSliceCmd)(nil)

func NewZSliceCmd(ctx context.Context, args ...interface{}) *ZSliceCmd {
	return &ZSliceCmd{
		baseCmd: baseCmd{
			ctx:  ctx,
			args: args,
		},
	}
}

func (cmd *ZSliceCmd) Val() []Z {
	return cmd.val
}

func (cmd *ZSliceCmd) Result() ([]Z, error) {
	return cmd.val, cmd.err
}

func (cmd *ZSliceCmd) setReply(val interface{}, err error) {
	var arr []interface{}
	arr, cmd.err = redis.Values(val, err)
	if cmd.err != nil {
		return
	}

	cmd.val = make([]Z, len(arr)/2)
	for i := range cmd.val {
		v, _ := redis.String(arr[2*i], nil)
		s, _ := redis.Float64(arr[2*i+1], nil)
		cmd.val[i] = Z{s, v}
	}
}

// TODO:
// type ScanCmd struct {
// 	baseCmd
// }
