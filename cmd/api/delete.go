package main

type DeleteCmd struct {
	Model string `arg:"" name:"model" help:"Model to delete"`
}

func (cmd *DeleteCmd) Run(ctx *Globals) error {
	if err := ctx.client.DeleteModel(ctx.ctx, cmd.Model); err != nil {
		return err
	}
	return nil
}
