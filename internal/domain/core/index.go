package core

import (
	"context"
	"sync"

	"github.com/rendau/dop/adapters/logger"
)

type St struct {
	lg           logger.Lite
	imgMaxWidth  int
	imgMaxHeight int
	testing      bool

	ctx       context.Context
	ctxCancel context.CancelFunc

	Static *Static
	Img    *Img

	wg sync.WaitGroup
}

func New(
	lg logger.Lite,
	imgMaxWidth int,
	imgMaxHeight int,
	testing bool,
) *St {
	c := &St{
		lg:           lg,
		imgMaxWidth:  imgMaxWidth,
		imgMaxHeight: imgMaxHeight,
		testing:      testing,
	}

	c.ctx, c.ctxCancel = context.WithCancel(context.Background())

	c.Static = NewStatic(c)
	c.Img = NewImg(c)

	return c
}

func (c *St) Start() {
}

func (c *St) StopAndWaitJobs() {
	c.ctxCancel()
	c.wg.Wait()
}
