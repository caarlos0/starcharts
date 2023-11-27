package chart

import (
	"github.com/golang/freetype/truetype"
	"sync"
)

var (
	fontLock sync.Mutex
	fontDef  *truetype.Font
)

func GetFont() *truetype.Font {
	if fontDef == nil {
		fontLock.Lock()
		defer fontLock.Unlock()
		if fontDef == nil {
			loadedFont, err := truetype.Parse(Roboto)
			if err != nil {
				panic(err)
			}

			fontDef = loadedFont
		}
	}

	return fontDef
}
