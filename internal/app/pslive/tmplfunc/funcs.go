// Package tmplfunc contains extension functions for templates
package tmplfunc

import (
	"html/template"
	"net/url"
)

func FuncMap(h HashedNamers) template.FuncMap {
	return template.FuncMap{
		"queryEscape":            url.QueryEscape,
		"derefBool":              DerefBool,
		"derefInt":               DerefInt,
		"derefFloat32":           DerefFloat32,
		"derefString":            DerefString,
		"appHashed":              h.AppHashed,
		"staticHashed":           h.StaticHashed,
	}
}
