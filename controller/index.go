package controller

import (
	"html/template"
	"io/fs"
	"net/http"
	"strings"

	"github.com/caarlos0/httperr"
)

// Index 这个函数用于处理对特定页面（可能是首页或某个主要页面）的请求，主要负责生成并返回该页面的内容。
func Index(filesystem fs.FS, version string) http.Handler {
	// 从给定的文件系统（filesystem）中解析模板文件。这里可能是读取 HTML 模板文件，用于生成页面的结构和布局。
	indexTemplate, err := template.ParseFS(filesystem, base, index)
	if err != nil {
		panic(err)
	}

	// httperr.NewF返回一个http.Handler类型的对象(一个匿名函数)
	// 当有 HTTP 请求到来时，这个匿名函数会被调用。
	return httperr.NewF(func(w http.ResponseWriter, r *http.Request) error {
		// 使用解析得到的模板对象（indexTemplate）执行模板渲染操作，
		// 将一个包含版本信息（version）的映射传递给模板，
		// 模板会根据这个数据填充相应的位置，最终将生成的页面内容写入到 HTTP 响应写入器（w）中，返回给客户端。
		return indexTemplate.Execute(w, map[string]string{"Version": version})
	})
}

// HandleForm 该函数主要用于处理表单提交后的请求，通常用于重定向到特定的页面或执行特定的操作
func HandleForm() http.HandlerFunc {
	// 创建 HTTP 请求处理函数
	return func(w http.ResponseWriter, r *http.Request) {
		// 从请求的表单值中获取名为 “repository” 的值，并去除其前缀 “https://github.com/”，提取出可能是仓库名称或特定标识的部分。
		repo := strings.TrimPrefix(r.FormValue("repository"), "https://github.com/")
		// 执行重定向操作，将客户端重定向到由提取出的repo变量确定的页面。http.StatusSeeOther表示使用 303 状态码进行重定向，
		// 通常用于 POST 请求后的重定向，以避免浏览器重复提交表单
		http.Redirect(w, r, repo, http.StatusSeeOther)
	}
}
