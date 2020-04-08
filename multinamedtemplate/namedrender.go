package multinamedtemplate

import (
	"html/template"

	"github.com/gin-gonic/gin/render"
)

// NamedRender 接口.
type NamedRender interface {
	render.HTMLRender
	AddFromFiles(tmplname string, files ...string) *template.Template
	AddFromFilesByNamed(tmplname, definedname string, files ...string) *template.Template
}
