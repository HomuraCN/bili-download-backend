package service

import (
	"bilibili-video-stream/internal/model"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// 混合密钥表
var mixinKeyEncTab = []int{
	46, 47, 18, 2, 53, 8, 23, 32, 15, 50, 10, 31, 58, 3, 45, 35, 27, 43, 5, 49,
	33, 9, 42, 19, 29, 28, 14, 39, 12, 38, 41, 13, 37, 48, 7, 16, 24, 55, 40,
	61, 26, 17, 0, 1, 60, 51, 30, 4, 22, 25, 54, 21, 56, 59, 6, 63, 57, 62, 11,
	36, 20, 34, 44, 52,
}

// GetMixinKey 对 imgKey 和 subKey 进行字符重排
func GetMixinKey(orig string) string {
	var sb strings.Builder
	for _, idx := range mixinKeyEncTab {
		if idx < len(orig) {
			sb.WriteByte(orig[idx])
		}
	}
	return sb.String()[:32]
}

// GetWbiKeys 获取最新的 img_key 和 sub_key
func GetWbiKeys() (string, string, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://api.bilibili.com/x/web-interface/nav", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var navResp model.NavResponse
	json.Unmarshal(body, &navResp)

	imgUrl := navResp.Data.WbiImg.ImgUrl
	subUrl := navResp.Data.WbiImg.SubUrl

	if imgUrl == "" || subUrl == "" {
		return "", "", errors.New("failed to get wbi keys")
	}

	imgKey := extractKey(imgUrl)
	subKey := extractKey(subUrl)
	return imgKey, subKey, nil
}

func extractKey(urlStr string) string {
	parts := strings.Split(urlStr, "/")
	fileName := parts[len(parts)-1]
	key := strings.Split(fileName, ".")[0]
	return key
}

// SignAndGetWbiQuery 核心方法：为参数签名并返回查询字符串
func SignAndGetWbiQuery(params map[string]string) (string, error) {
	imgKey, subKey, err := GetWbiKeys()
	if err != nil {
		return "", err
	}
	mixinKey := GetMixinKey(imgKey + subKey)

	currTime := fmt.Sprintf("%d", time.Now().Unix())
	params["wts"] = currTime

	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var queryStr strings.Builder
	for i, k := range keys {
		if i > 0 {
			queryStr.WriteString("&")
		}
		encodedValue := url.QueryEscape(params[k])
		queryStr.WriteString(k + "=" + encodedValue)
	}

	rawStr := queryStr.String() + mixinKey
	hash := md5.Sum([]byte(rawStr))
	wRid := hex.EncodeToString(hash[:])

	return queryStr.String() + "&w_rid=" + wRid, nil
}
