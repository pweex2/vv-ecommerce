package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type GatewayHandler struct {
	orderTarget     *url.URL
	inventoryTarget *url.URL
	paymentTarget   *url.URL
}

func NewGatewayHandler(orderURL, inventoryURL, paymentURL string) *GatewayHandler {
	oTarget, _ := url.Parse(orderURL)
	iTarget, _ := url.Parse(inventoryURL)
	pTarget, _ := url.Parse(paymentURL)

	return &GatewayHandler{
		orderTarget:     oTarget,
		inventoryTarget: iTarget,
		paymentTarget:   pTarget,
	}
}

// Proxy 是核心方法，它创建一个 ReverseProxy 并把请求转发出去
func (h *GatewayHandler) Proxy(target *url.URL) gin.HandlerFunc {
	return func(c *gin.Context) {
		proxy := httputil.NewSingleHostReverseProxy(target)

		// 自定义 Director 来修改请求（如果需要）
		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originalDirector(req)
			// 这里可以添加额外的 Header，比如把网关收到的 UserID 传下去
			// req.Header.Set("X-User-ID", c.GetString("user_id"))

			// 重要：我们需要重写 Host Header，否则有些后端服务会拒绝请求
			req.Host = target.Host
		}

		// 自定义 ErrorHandler
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			fmt.Printf("Proxy error: %v\n", err)
			c.JSON(http.StatusBadGateway, gin.H{"error": "Service unavailable"})
		}

		// ModifyResponse: 统一包装成功响应为 {code: 0, msg: "success", data: ...}
		proxy.ModifyResponse = func(resp *http.Response) error {
			// 只处理 200 OK 且是 JSON 的响应
			if resp.StatusCode == http.StatusOK && strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return err
				}
				resp.Body.Close()

				// 构造统一响应结构
				wrapper := struct {
					Code    int             `json:"code"`
					Message string          `json:"message"`
					Data    json.RawMessage `json:"data"`
				}{
					Code:    0,
					Message: "success",
					Data:    json.RawMessage(body),
				}

				newBody, err := json.Marshal(wrapper)
				if err != nil {
					return err
				}

				// 重置 Body 和 Content-Length
				resp.Body = io.NopCloser(bytes.NewReader(newBody))
				resp.ContentLength = int64(len(newBody))
				resp.Header.Set("Content-Length", strconv.Itoa(len(newBody)))
			}
			return nil
		}

		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

func (h *GatewayHandler) OrderProxy() gin.HandlerFunc {
	return h.Proxy(h.orderTarget)
}

func (h *GatewayHandler) InventoryProxy() gin.HandlerFunc {
	return h.Proxy(h.inventoryTarget)
}

func (h *GatewayHandler) PaymentProxy() gin.HandlerFunc {
	return h.Proxy(h.paymentTarget)
}
