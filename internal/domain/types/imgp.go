package types

import (
	"fmt"
)

type ImgParsSt struct {
	Method    string
	Width     int
	Height    int
	Blur      float64
	Grayscale bool
}

func (o *ImgParsSt) Reset() {
	o.Method = ""
	o.Width = 0
	o.Height = 0
	o.Blur = 0
	o.Grayscale = false
}

func (o *ImgParsSt) IsEmpty() bool {
	return *o == ImgParsSt{}
}

func (o *ImgParsSt) String() string {
	return fmt.Sprintf("m=%s&w=%d&h=%d&blur=%fgrayscale=%v", o.Method, o.Width, o.Height, o.Blur, o.Grayscale)
}
