package main

import (
	"embed"
	"net/http"
	"os"
	"time"

	"github.com/apex/httplog"
	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
	"github.com/caarlos0/starcharts/config"
	"github.com/caarlos0/starcharts/controller"
	"github.com/caarlos0/starcharts/internal/cache"
	"github.com/caarlos0/starcharts/internal/github"
	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// FS 是文件的只读集合，通常使用 //go:embed指令进行初始化当声明时没有 //go:embed指令，FS 是一个空文件系统。
// FS是只读值，因此同时从多个goroutine中使用是安全的，并且将FS类型的值彼此分配也是安全的。
// FS实现了fs.FS，因此它可以与任何理解文件系统接口的包一起使用，包括net/http、text/template和html/template。
// 把它想象成一个 “宝箱”，里面装着网页需要的各种装饰和素材
//
//go:embed static/*
var static embed.FS
var version = "devel"

func main() {
	log.SetHandler(text.New(os.Stderr))
	// log.SetLevel(log.DebugLevel)
	config := config.Get()
	ctx := log.WithField("listen", config.Listen)
	options, err := redis.ParseURL(config.RedisURL) // redis参数解析url获取redis的addr等参数
	if err != nil {
		log.WithError(err).Fatal("invalid redis_url")
	}
	redis := redis.NewClient(options) // 根据解析后的参数新建redis客户端
	cache := cache.New(redis)         // 新建redis缓存，支持新增，查询，删除
	defer cache.Close()               // 关闭redis客户端

	github := github.New(config, cache) // 新建github客户端

	r := mux.NewRouter() // 新建路由，路由器注册要匹配的路由并调度处理程序。

	r.Path("/"). // 用户打开网站，就会看到这个首页
			Methods(http.MethodGet). // 请求方式
		// 这个函数用于处理对特定页面（可能是首页或某个主要页面）的请求，主要负责生成并返回该页面的内容。
		Handler(controller.Index(static, version)) // 处理函数

	r.Path("/"). // 处理表单数据、保存数据或者重定向到其他页面
			Methods(http.MethodPost).
		// 该函数主要用于处理表单提交后的请求，通常用于重定向到特定的页面或执行特定的操作。
		HandlerFunc(controller.HandleForm())

	r.PathPrefix("/static/").
		Methods(http.MethodGet).
		// 用于提供静态文件服务，比如网页的 CSS、JavaScript、图片等文件
		// http.FS(static)将名为static的资源（可能是一个实现了http.FileSystem接口的对象）转换为文件系统实现，
		// 然后http.FileServer创建一个处理程序，为 HTTP 请求提供以这个文件系统为根的文件服务。
		Handler(http.FileServer(http.FS(static)))

	// 上面是前端部分，下面是后端部分

	// 使用之前创建的 GitHub 客户端（github）和缓存（cache）来生成特定仓库的 SVG 图表并返回给用户
	r.Path("/{owner}/{repo}.svg"). // 其中owner和repo可能代表某个资源的所有者和仓库名称等信息。
					Methods(http.MethodGet).
		// 接收github和cache两个参数，可能用于生成或获取特定的 SVG 图表，
		Handler(controller.GetRepoChart(github, cache))

	// 这个函数会使用静态资源（static）、GitHub 客户端（github）、缓存（cache）和版本信息（version）来获取特定仓库的详细信息并返回给用户。
	// 可能包括仓库的描述、文件列表等信息
	r.Path("/{owner}/{repo}").
		Methods(http.MethodGet).
		Handler(controller.GetRepo(static, github, cache, version))

	// generic metrics
	// 请求计数器
	requestCounter := promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "starcharts",
		Subsystem: "http",
		Name:      "requests_total",
		Help:      "total requests",
	}, []string{"code", "method"})
	// 响应计数器
	responseObserver := promauto.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "starcharts",
		Subsystem: "http",
		Name:      "responses",
		Help:      "response times and counts",
	}, []string{"code", "method"})

	r.Methods(http.MethodGet).Path("/metrics").Handler(promhttp.Handler())

	// http服务端
	srv := &http.Server{
		Handler: httplog.New(
			promhttp.InstrumentHandlerDuration(
				responseObserver,
				promhttp.InstrumentHandlerCounter(
					requestCounter,
					r,
				),
			),
		),
		Addr:         config.Listen,
		WriteTimeout: 60 * time.Second,
		ReadTimeout:  60 * time.Second,
	}
	ctx.Info("starting up...")
	// http请求服务端监听
	ctx.WithError(srv.ListenAndServe()).Error("failed to start up server")
}
