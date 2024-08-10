package main

type PingCmd struct{}

func (cmd *PingCmd) Run(ctx *Globals) error {
	if err := ctx.client.Ping(ctx.ctx); err != nil {
		return err
	}
	return ctx.writer.Writeln("OK")
}
