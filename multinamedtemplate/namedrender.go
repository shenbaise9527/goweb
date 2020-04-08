package multinamedtemplate

import (
	"text/template"

	"github.com/gin-gonic/gin/render"
)

// NamedRender接口.
type NamedRender interface {
	render.HTMLRender
	AddFromFiles(tmplname string, files ...string) *template.Template
	AddFromFilesByNamed(tmplname, definedname string, files ...string) *template.Template
}
