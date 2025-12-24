package service

import (
	"bilibili-video-stream/internal/dao"
	"bilibili-video-stream/internal/model"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// GetQRCode 获取二维码
// 返回值修改为 model.QRCodeResponse 以满足前端嵌套层级需求
func GetQRCode() (model.QRCodeResponse, error) {
	var result model.QRCodeResponse
	var innerData model.QRCodeData

	resp, err := http.Get("https://passport.bilibili.com/x/passport-login/web/qrcode/generate")
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var biliResp model.BiliBaseResp
	if err := json.Unmarshal(body, &biliResp); err != nil {
		return result, err
	}

	if biliResp.Code != 0 {
		return result, errors.New(biliResp.Message)
	}

	// 映射数据
	if dataMap, ok := biliResp.Data.(map[string]interface{}); ok {
		if url, ok := dataMap["url"].(string); ok {
			innerData.Url = url
		}
		if key, ok := dataMap["qrcode_key"].(string); ok {
			innerData.QRCodeKey = key
		}
	}

	// 包装进 Data 字段
	result.Data = innerData
	return result, nil
}

// CheckQRCodeStatus 检查状态
// 修改：增加返回值 bool，代表是否登录成功
func CheckQRCodeStatus(qrKey string) (model.PollData, bool, error) {
	var pollResult model.PollData

	url := fmt.Sprintf("https://passport.bilibili.com/x/passport-login/web/qrcode/poll?qrcode_key=%s", qrKey)

	client := &http.Client{Timeout: 5 * time.Second}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return pollResult, false, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var biliResp model.BiliBaseResp
	json.Unmarshal(body, &biliResp)

	// 解析数据
	if dataMap, ok := biliResp.Data.(map[string]interface{}); ok {
		if u, ok := dataMap["url"].(string); ok {
			pollResult.Url = u
		}
		if rt, ok := dataMap["refresh_token"].(string); ok {
			pollResult.RefreshToken = rt
		}
		if msg, ok := dataMap["message"].(string); ok {
			pollResult.Message = msg
		}
		// 注意：B站接口返回的 code 是数字
		if c, ok := dataMap["code"].(float64); ok {
			pollResult.Code = int(c)
		}
	} else {
		// 如果没有 data 字段，说明可能出错了 (比如 key 过期 code 86038 可能在最外层)
		if biliResp.Code != 0 {
			pollResult.Code = biliResp.Code
			pollResult.Message = biliResp.Message
		}
	}

	// 逻辑判断：只有 Code 为 0 才是成功
	if pollResult.Code == 0 {
		// 提取 Cookie
		cookies := resp.Cookies()
		var saveCookie dao.CookieData

		for _, c := range cookies {
			if c.Name == "SESSDATA" {
				saveCookie.SessData = c.Value
			} else if c.Name == "DedeUserID" {
				saveCookie.DedeUserID = c.Value
			} else if c.Name == "bili_jct" {
				saveCookie.BiliJct = c.Value
			}
		}

		if saveCookie.SessData != "" {
			_ = dao.Store.SaveCookie(saveCookie)
			fmt.Println("登录成功！用户ID:", saveCookie.DedeUserID)
			return pollResult, true, nil
		}
	}

	return pollResult, false, nil
}
