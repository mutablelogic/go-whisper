package main

type DeleteCmd struct {
	Model string `arg:"" help:"Model id to delete"`
}

func (cmd *DeleteCmd) Run(ctx *Globals) error {
	if err := ctx.service.DeleteModelById(cmd.Model); err != nil {
		return err
	}
	return ModelsCmd{}.Run(ctx)
}
