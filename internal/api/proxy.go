package api

import (
	"bilibili-video-stream/internal/model"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ProxyHandler 视频流代理
// 前端请求: /proxy?url=http://...
// 后端动作: 带上 Referer 请求 B 站，然后把数据流 pipe 给前端
func ProxyHandler(c *gin.Context) {
	targetUrl := c.Query("url")
	if targetUrl == "" {
		c.JSON(http.StatusBadRequest, model.Fail("Missing url parameter"))
		return
	}

	// 1. 创建请求
	req, err := http.NewRequest("GET", targetUrl, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Fail("Invalid URL"))
		return
	}

	// 2. 伪装 Header (关键步骤)
	// B站检查 Referer，必须是 https://www.bilibili.com
	req.Header.Set("Referer", "https://www.bilibili.com")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	// 3. 发起请求
	// 注意：这里使用 DefaultClient，实际生产中建议使用带超时的 Client
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, model.Fail("Failed to request upstream"))
		return
	}
	defer resp.Body.Close()

	// 4. 将上游(B站)的响应头复制给前端
	// 这样前端能拿到 Content-Length (用于显示进度) 和 Content-Type
	for k, v := range resp.Header {
		// 过滤掉一些可能引起冲突的头
		if k != "Access-Control-Allow-Origin" {
			c.Writer.Header()[k] = v
		}
	}

	// 强制允许跨域 (虽然 main.go 里配了，但这里覆盖一下更保险)
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

	// 设置状态码
	c.Status(resp.StatusCode)

	// 5. 核心：管道转发 (Stream Copy)
	// 直接把 B 站发来的数据流写入到 ResponseWriter，不占用后端内存
	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		// 流传输过程中断，通常不需要特殊处理，日志记录即可
	}
}
