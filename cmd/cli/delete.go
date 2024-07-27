package main

type DeleteCmd struct {
	Id string `arg:"" required:"" help:"Model Identifier" type:"string"`
}

func (cmd *DeleteCmd) Run(ctx *Globals) error {
	err := ctx.api.DeleteModel(ctx.ctx, cmd.Id)
	if err != nil {
		return err
	}
	return nil
}
