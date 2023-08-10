package xredis

import (
	"context"
	"fmt"

	"kratos/pkg/cache/redis"
)

var ErrNil = redis.ErrNil

type pool interface {
	Get(ctx context.Context) redis.Conn
	Close() error
}

type Client struct {
	cmdable

	p        pool
	selfInit bool
}

type Config = redis.Config

func New(conf *Config, options ...redis.DialOption) (c *Client) {
	c = &Client{
		p:        redis.NewPool(conf, options...),
		selfInit: true,
	}
	c.cmdable = c.Process

	return
}

func NewWithPool(p *redis.Pool) (c *Client) {
	c = &Client{
		p: p,
	}
	c.cmdable = c.Process

	return
}

type poolWrap struct {
	r *redis.Redis
}

func (w poolWrap) Get(ctx context.Context) redis.Conn {
	return w.r.Conn(ctx)
}

func (w poolWrap) Close() error {
	return w.r.Close()
}

func NewWithRedis(r *redis.Redis) (c *Client) {
	c = &Client{
		p: poolWrap{r},
	}
	c.cmdable = c.Process
	return
}

func (c *Client) Close() error {
	if c.selfInit {
		return c.p.Close()
	}
	return nil
}

// Do creates a Cmd from the args and processes the cmd.
func (c *Client) Do(ctx context.Context, args ...interface{}) *Cmd {
	cmd := NewCmd(ctx, args...)
	_ = c.Process(ctx, cmd)
	return cmd
}

func (c *Client) Process(ctx context.Context, cmd Cmder) error {
	// 参数构建问题
	if cmd.Err() != nil {
		return cmd.Err()
	}

	conn := c.p.Get(ctx)
	cmd.setReply(conn.Do(cmd.Name(), cmd.Args()[1:]...))
	fmt.Printf("[Process] err:(%v)", cmd.Err())
	conn.Close()
	return cmd.Err()
}

func (c *Client) processPipeline(ctx context.Context, cmds []Cmder) (err error) {
	// 验证参数构建问题
	for _, cmd := range cmds {
		if err = cmd.Err(); err != nil {
			return
		}
	}

	conn := c.p.Get(ctx)
	defer conn.Close()

	for _, cmd := range cmds {
		err = conn.Send(cmd.Name(), cmd.Args()[1:]...)
		if err != nil {
			return
		}
	}

	err = conn.Flush()
	if err != nil {
		return
	}

	for _, cmd := range cmds {
		cmd.setReply(conn.Receive())
	}
	return
}

func (c *Client) Pipelined(ctx context.Context, fn func(Pipeliner) error) ([]Cmder, error) {
	return c.Pipeline().Pipelined(ctx, fn)
}

func (c *Client) Pipeline() Pipeliner {
	pipe := Pipeline{
		exec: c.processPipeline,
	}
	pipe.init()
	return &pipe
}
