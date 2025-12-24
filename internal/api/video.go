package api

import (
	"bilibili-video-stream/internal/model"
	"bilibili-video-stream/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// DownloadVideoWBI 处理前端的下载请求
// 实际上是进行 WBI 签名解析，返回视频的真实下载直链 (VideoUrl, AudioUrl)
// 前端拿到这两个 URL 后，自行在浏览器中下载并合并
func DownloadVideoWBI(c *gin.Context) {
	url := c.Query("url")
	// fileName := c.Query("fileName") // 可选，目前后端主要负责解析出默认标题

	if url == "" {
		c.JSON(http.StatusOK, model.Fail("缺少参数: url"))
		return
	}

	// 调用 Service 层解析
	result, err := service.ResolveVideo(url)
	if err != nil {
		c.JSON(http.StatusOK, model.Fail("解析失败: "+err.Error()))
		return
	}

	// 返回成功结果
	c.JSON(http.StatusOK, model.Success(result))
}
