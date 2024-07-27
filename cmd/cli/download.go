package main

import "fmt"

type DownloadCmd struct {
	Path string `arg:"" required:"" help:"Model Path" type:"string"`
}

func (cmd *DownloadCmd) Run(ctx *Globals) error {
	type progress struct {
		Status    string `json:"status" writer:",width:60"`
		Total     int64  `json:"total,omitempty" writer:",right,width:12,"`
		Completed int64  `json:"completed,omitempty" writer:",right,width:12,"`
		Percent   string `json:"percent,omitempty" writer:",width:8,right"`
	}
	model, err := ctx.api.DownloadModel(ctx.ctx, cmd.Path, func(status string, cur, total int64) {
		percent := ""
		if cur < total {
			percent = fmt.Sprintf("%.1f%%", float32(cur)*100/float32(total))
		}
		if status != "" {
			ctx.writer.Write(progress{
				Status:    status,
				Completed: cur,
				Total:     total,
				Percent:   percent,
			})
		}
	})
	if err != nil {
		return err
	}
	return ctx.writer.Write(model)
}
