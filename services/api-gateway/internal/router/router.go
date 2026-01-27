package router

import (
	"api-gateway/internal/handler"
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewRouter(h *handler.GatewayHandler) *gin.Engine {
	r := gin.Default() // 使用 Default，自带 Logger 和 Recovery

	// Health Check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "API Gateway is healthy"})
	})

	// API Version 1 Group
	v1 := r.Group("/api/v1")
	{
		// 1. Order Routes -> Order Service
		// 注意：我们这里直接把所有 /orders 开头的请求都转发过去
		// 如果后端是 GET /orders，网关也是 GET /api/v1/orders，需要 strip prefix 吗？
		// 现在的 httputil.NewSingleHostReverseProxy 不会自动 strip path。
		// 所以如果网关路径是 /api/v1/orders，后端收到的也是 /api/v1/orders。
		// 但我们的 Order Service 监听的是 /orders。
		// 所以我们需要重写 URL Path。

		orders := v1.Group("/orders")
		orders.Any("/*path", func(c *gin.Context) {
			// 重写 Path: /api/v1/orders/xxx -> /orders/xxx
			// 实际上 Gin 的 Group 已经处理了路由匹配，我们只需要把 Path 改对
			// 简单粗暴法：直接把请求转发给 OrderService 的根目录？不对。

			// 让我们换一种更简单的路由定义方式，针对具体接口定义，或者写一个 PathRewrite 中间件。
			// 为了演示清晰，我们先针对具体接口做映射。
		})

		// 重新定义路由映射 (Explicit Mapping)

		// Order Service
		// Public: 下单
		v1.POST("/orders", func(c *gin.Context) {
			c.Request.URL.Path = "/orders"
			h.OrderProxy()(c)
		})
		// Public: 查单
		v1.GET("/orders", func(c *gin.Context) {
			c.Request.URL.Path = "/orders"
			h.OrderProxy()(c)
		})

		// Inventory Service
		// Public: 商品列表
		v1.GET("/products", func(c *gin.Context) {
			// 前端叫 /products，后端叫 /inventories
			c.Request.URL.Path = "/inventories"
			h.InventoryProxy()(c)
		})
		// Public: 商品详情
		v1.GET("/products/:sku", func(c *gin.Context) {
			// 前端 /products/SKU123 -> 后端 /inventory/sku?sku=SKU123
			// 或者后端改一下路由？现在的后端是 GET /inventory/sku (Query Param)
			// 我们来适配一下：
			sku := c.Param("sku")
			c.Request.URL.Path = "/inventory/sku"
			q := c.Request.URL.Query()
			q.Add("sku", sku)
			c.Request.URL.RawQuery = q.Encode()

			h.InventoryProxy()(c)
		})

		// Payment Service
		// Public: 消费记录
		v1.GET("/payments", func(c *gin.Context) {
			c.Request.URL.Path = "/payments"
			h.PaymentProxy()(c)
		})
		// Public: 创建支付 (测试用)
		v1.POST("/payments", func(c *gin.Context) {
			c.Request.URL.Path = "/payments"
			h.PaymentProxy()(c)
		})

		// Inventory Service (Additional)
		// Internal/Admin: 创建库存
		v1.POST("/inventory/create", func(c *gin.Context) {
			c.Request.URL.Path = "/inventory/create"
			h.InventoryProxy()(c)
		})
	}

	return r
}
