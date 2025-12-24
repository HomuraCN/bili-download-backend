package main

import (
	"bilibili-video-stream/internal/api" // 1. 引入新创建的 api 包
	"bilibili-video-stream/internal/dao"
	"bilibili-video-stream/internal/model"
	"bilibili-video-stream/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. 初始化 Cookie 存储
	dao.InitStore("cookie.json")

	// 2. 初始化 Web 框架
	r := gin.Default()

	// 3. 配置 CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// --- 登录模块 ---

	// 获取二维码
	r.GET("/getQRCode", func(c *gin.Context) {
		data, err := service.GetQRCode()
		if err != nil {
			c.JSON(http.StatusOK, model.Fail(err.Error()))
			return
		}
		c.JSON(http.StatusOK, model.Success(data))
	})

	// 验证扫码状态
	r.GET("/verifyQRCode", func(c *gin.Context) {
		key := c.Query("qrcodekey")
		if key == "" {
			key = c.Query("qrcode_key")
		}
		if key == "" {
			c.JSON(http.StatusOK, model.Fail("缺少 qrcodekey 参数"))
			return
		}

		data, success, err := service.CheckQRCodeStatus(key)
		if err != nil {
			c.JSON(http.StatusOK, model.Fail(err.Error()))
			return
		}

		if success {
			c.JSON(http.StatusOK, model.Success(data))
		} else {
			c.JSON(http.StatusOK, model.Result{
				Code: 500,
				Msg:  data.Message,
				Data: data,
			})
		}
	})

	// --- 视频处理模块 ---

	// 接口：解析视频流
	r.GET("/video/resolve", func(c *gin.Context) {
		bvid := c.Query("bvid")
		cid := c.Query("cid")

		if bvid == "" || cid == "" {
			c.JSON(http.StatusOK, model.Fail("缺少 bvid 或 cid 参数"))
			return
		}

		data, err := service.ResolveVideoUrl(bvid, cid)
		if err != nil {
			c.JSON(http.StatusOK, model.Fail("解析失败: "+err.Error()))
			return
		}

		c.JSON(http.StatusOK, model.Success(data))
	})

	// --- 新增: 视频流代理接口 (解决 403/CORS 问题) ---
	r.GET("/proxy", api.ProxyHandler) // <--- 添加这一行

	// --- 新增: 下载与进度模块 (第一阶段) ---

	// 注册 WebSocket 路由，用于前端监听进度
	// 前端连接地址: ws://localhost:9961/progress
	r.GET("/progress", api.ProgressWebSocketHandler)

	// --- 新增: 视频下载接口 (第二阶段) ---
	// 对应 views/VideoDownload.vue 中的请求
	r.GET("/downloadVideoWBI", api.DownloadVideoWBI)

	// --- 辅助接口 ---

	r.GET("/api/cookie/view", func(c *gin.Context) {
		cookie, _ := dao.Store.LoadCookie()
		if cookie.SessData == "" {
			c.JSON(http.StatusOK, model.Fail("暂无 Cookie，请先扫码登录"))
		} else {
			displayCookie := cookie
			if len(displayCookie.SessData) > 10 {
				displayCookie.SessData = displayCookie.SessData[:10] + "..."
			}
			c.JSON(http.StatusOK, model.Success(displayCookie))
		}
	})

	// 启动服务
	r.Run(":9961")
}
