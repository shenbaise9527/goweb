package multinamedtemplate

import (
	"fmt"
	"html/template"

	"github.com/gin-gonic/gin/render"
)

type multiRender struct {
	name string
	tmpl *template.Template
}

// 命名的多模板.
type NamedMultiRender map[string]multiRender

// NewRender 创建render.
func NewRender() NamedMultiRender {
	return make(NamedMultiRender)
}

// AddFromFiles 根据模板文件创建模板对象.
func (nr NamedMultiRender) AddFromFiles(tmplname string, files ...string) *template.Template {
	return nr.AddFromFilesByNamed(tmplname, "", files...)
}

// AddFromFiles 根据模板文件创建模板对象.
func (nr NamedMultiRender) AddFromFilesByNamed(
	tmplname, definedname string, files ...string) *template.Template {
	tmpl := template.Must(template.ParseFiles(files...))
	multiR := multiRender{
		name: definedname,
		tmpl: tmpl,
	}

	if len(tmplname) == 0 {
		panic("template name cannot be empty")
	}

	if _, ok := nr[tmplname]; ok {
		panic(fmt.Sprintf("template %s already exists", tmplname))
	}

	nr[tmplname] = multiR

	return tmpl
}

// Instance 初始化render.
func (nr NamedMultiRender) Instance(name string, data interface{}) render.Render {
	r := nr[name]
	return render.HTML{
		Template: r.tmpl,
		Name:     r.name,
		Data:     data,
	}
}
