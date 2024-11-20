package models

import (
	"fmt"
)

type Culture struct {
	Lang string `json:"lang"`
}

/*
	`heading.%v` => `heading.{{lang}}`

	exampe:
	 `heading.tm`
*/
func (c *Culture) Stringf(format string) string {
	return fmt.Sprintf(format, c.Lang)
}

func (c *Culture) ToTranslation(str string) Translation {
	t := Translation{}

	switch c.Lang {
	case "ru":
		t.Ru = str
	case "en":
		t.En = str
	case "tr":
		t.Tr = str
	case "tm":
		t.Tm = str
	}

	return t
}
