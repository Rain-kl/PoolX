package utils

import "html/template"

func UnescapeHTML(x string) interface{} {
	return template.HTML(x)
}
