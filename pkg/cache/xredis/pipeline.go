package xredis

import "context"

type pipelineExecer func(context.Context, []Cmder) error

type Pipeliner interface {
	Cmdable
	Do(ctx context.Context, args ...interface{}) *Cmd
	Process(ctx context.Context, cmd Cmder) error
	Discard() // error
	Exec(ctx context.Context) ([]Cmder, error)
}

type Pipeline struct {
	cmdable

	exec pipelineExecer

	cmds []Cmder
}

var _ Pipeliner = (*Pipeline)(nil)

func (c *Pipeline) init() {
	c.cmdable = c.Process
}

func (c *Pipeline) Do(ctx context.Context, args ...interface{}) *Cmd {
	cmd := NewCmd(ctx, args...)
	_ = c.Process(ctx, cmd)
	return cmd
}

// Process queues the cmd for later execution.
func (c *Pipeline) Process(ctx context.Context, cmd Cmder) error {
	c.cmds = append(c.cmds, cmd)
	return nil
}

// Discard resets the pipeline and discards queued commands.
func (c *Pipeline) Discard() {
	c.cmds = c.cmds[:0]
}

// Exec executes all previously queued commands using one
// client-server roundtrip.
//
// Exec always returns list of commands and error of the first failed
// command if any.
func (c *Pipeline) Exec(ctx context.Context) ([]Cmder, error) {
	if len(c.cmds) == 0 {
		return nil, nil
	}

	cmds := c.cmds
	c.cmds = nil

	return cmds, c.exec(ctx, cmds)
}

func (c *Pipeline) Pipelined(ctx context.Context, fn func(Pipeliner) error) ([]Cmder, error) {
	if err := fn(c); err != nil {
		return nil, err
	}
	cmds, err := c.Exec(ctx)
	// _ = c.Close()
	return cmds, err
}

func (c *Pipeline) Pipeline() Pipeliner {
	return c
}
