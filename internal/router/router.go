package router

import (
	"net/http"
	"path/filepath"
	"strings"
)

// HandlerFunc 处理函数类型
type HandlerFunc func(http.ResponseWriter, *http.Request)

// MiddlewareFunc 中间件函数类型
type MiddlewareFunc func(HandlerFunc) HandlerFunc

// Router 高性能路由器
type Router struct {
	trees      map[string]*node // 每个 HTTP 方法对应一个路由树
	middleware []MiddlewareFunc // 全局中间件
	NotFound   HandlerFunc      // 404 处理器
}

// node 路由树节点
type node struct {
	path     string           // 路径片段
	handler  HandlerFunc      // 处理函数
	children map[string]*node // 子节点
	param    *node            // 参数节点 (:id)
	wildcard *node            // 通配符节点 (*)
	isParam  bool             // 是否为参数节点
	isWild   bool             // 是否为通配符节点
}

// New 创建新的路由器
func New() *Router {
	return &Router{
		trees: make(map[string]*node),
		NotFound: func(w http.ResponseWriter, r *http.Request) {
			http.NotFound(w, r)
		},
	}
}

// Use 添加全局中间件
func (r *Router) Use(middleware MiddlewareFunc) {
	r.middleware = append(r.middleware, middleware)
}

// GET 注册 GET 路由
func (r *Router) GET(path string, handler HandlerFunc) {
	r.Handle("GET", path, handler)
}

// POST 注册 POST 路由
func (r *Router) POST(path string, handler HandlerFunc) {
	r.Handle("POST", path, handler)
}

// PUT 注册 PUT 路由
func (r *Router) PUT(path string, handler HandlerFunc) {
	r.Handle("PUT", path, handler)
}

// DELETE 注册 DELETE 路由
func (r *Router) DELETE(path string, handler HandlerFunc) {
	r.Handle("DELETE", path, handler)
}

// Handle 注册路由
func (r *Router) Handle(method, path string, handler HandlerFunc) {
	if path[0] != '/' {
		panic("路径必须以 '/' 开头")
	}

	// 获取或创建方法树
	root := r.trees[method]
	if root == nil {
		root = &node{}
		r.trees[method] = root
	}

	// 添加路由到树中
	r.addRoute(root, path, handler)
}

// addRoute 添加路由到树中
func (r *Router) addRoute(root *node, path string, handler HandlerFunc) {
	// 特殊处理根路径
	if path == "/" {
		root.handler = handler
		return
	}

	segments := strings.Split(path, "/")[1:] // 去掉第一个空字符串
	current := root

	for _, segment := range segments {
		if segment == "" {
			continue
		}

		// 检查是否为参数路由
		if strings.HasPrefix(segment, ":") {
			paramName := segment[1:]
			if current.param == nil {
				current.param = &node{
					path:    paramName,
					isParam: true,
				}
			}
			current = current.param
		} else if segment == "*" {
			// 通配符路由
			if current.wildcard == nil {
				current.wildcard = &node{
					path:   "*",
					isWild: true,
				}
			}
			current = current.wildcard
		} else {
			// 普通路由
			if current.children == nil {
				current.children = make(map[string]*node)
			}
			if current.children[segment] == nil {
				current.children[segment] = &node{
					path: segment,
				}
			}
			current = current.children[segment]
		}
	}

	current.handler = handler
}

// ServeHTTP 实现 http.Handler 接口
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// 查找路由
	handler, params := r.findRoute(req.Method, req.URL.Path)
	
	if handler == nil {
		r.NotFound(w, req)
		return
	}

	// 设置路径参数到请求上下文
	if len(params) > 0 {
		// 简化处理：将参数添加到 URL 查询参数中
		q := req.URL.Query()
		for key, value := range params {
			q.Set(key, value)
		}
		req.URL.RawQuery = q.Encode()
	}

	// 应用中间件
	finalHandler := handler
	for i := len(r.middleware) - 1; i >= 0; i-- {
		finalHandler = r.middleware[i](finalHandler)
	}

	finalHandler(w, req)
}

// findRoute 查找路由
func (r *Router) findRoute(method, path string) (HandlerFunc, map[string]string) {
	root := r.trees[method]
	if root == nil {
		return nil, nil
	}

	segments := strings.Split(path, "/")[1:]
	return r.searchRoute(root, segments, make(map[string]string))
}

// searchRoute 搜索路由
func (r *Router) searchRoute(node *node, segments []string, params map[string]string) (HandlerFunc, map[string]string) {
	// 如果没有更多段，检查当前节点是否有处理器
	if len(segments) == 0 {
		return node.handler, params
	}

	// 过滤空段
	var filteredSegments []string
	for _, seg := range segments {
		if seg != "" {
			filteredSegments = append(filteredSegments, seg)
		}
	}

	// 如果过滤后没有段，返回根处理器
	if len(filteredSegments) == 0 {
		return node.handler, params
	}

	segments = filteredSegments

	segment := segments[0]
	remaining := segments[1:]

	// 1. 尝试精确匹配
	if node.children != nil {
		if child, exists := node.children[segment]; exists {
			if handler, p := r.searchRoute(child, remaining, params); handler != nil {
				return handler, p
			}
		}
	}

	// 2. 尝试参数匹配
	if node.param != nil {
		newParams := make(map[string]string)
		for k, v := range params {
			newParams[k] = v
		}
		newParams[node.param.path] = segment
		if handler, p := r.searchRoute(node.param, remaining, newParams); handler != nil {
			return handler, p
		}
	}

	// 3. 尝试通配符匹配
	if node.wildcard != nil {
		return node.wildcard.handler, params
	}

	return nil, nil
}

// Static 静态文件服务
func (r *Router) Static(prefix, dir string) {
	fileServer := http.StripPrefix(prefix, http.FileServer(http.Dir(dir)))
	
	r.GET(prefix+"*", func(w http.ResponseWriter, req *http.Request) {
		fileServer.ServeHTTP(w, req)
	})
}

// Group 路由组
type Group struct {
	router     *Router
	prefix     string
	middleware []MiddlewareFunc
}

// Group 创建路由组
func (r *Router) Group(prefix string) *Group {
	return &Group{
		router: r,
		prefix: prefix,
	}
}

// Use 为路由组添加中间件
func (g *Group) Use(middleware MiddlewareFunc) {
	g.middleware = append(g.middleware, middleware)
}

// GET 路由组 GET 方法
func (g *Group) GET(path string, handler HandlerFunc) {
	g.Handle("GET", path, handler)
}

// POST 路由组 POST 方法
func (g *Group) POST(path string, handler HandlerFunc) {
	g.Handle("POST", path, handler)
}

// PUT 路由组 PUT 方法
func (g *Group) PUT(path string, handler HandlerFunc) {
	g.Handle("PUT", path, handler)
}

// DELETE 路由组 DELETE 方法
func (g *Group) DELETE(path string, handler HandlerFunc) {
	g.Handle("DELETE", path, handler)
}

// Handle 路由组处理方法
func (g *Group) Handle(method, path string, handler HandlerFunc) {
	// 应用组中间件
	finalHandler := handler
	for i := len(g.middleware) - 1; i >= 0; i-- {
		finalHandler = g.middleware[i](finalHandler)
	}
	
	// 组合路径
	fullPath := filepath.Join(g.prefix, path)
	if fullPath == "" {
		fullPath = "/"
	}
	
	g.router.Handle(method, fullPath, finalHandler)
}

// Group 创建子路由组
func (g *Group) Group(prefix string) *Group {
	return &Group{
		router:     g.router,
		prefix:     filepath.Join(g.prefix, prefix),
		middleware: append([]MiddlewareFunc{}, g.middleware...), // 复制父组中间件
	}
}
