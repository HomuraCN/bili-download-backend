package service

import (
	"bilibili-video-stream/internal/dao" // 1. 引入 dao 包读取 Cookie
	"bilibili-video-stream/internal/model"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// 常量定义，用于 BV 转 AV
const (
	XOR_CODE  = 23442827791579
	MASK_CODE = 2251799813685247
	ALPHABET  = "FcwAPNKTMug3GV5Lj7EJnHpWsx4tb8haYeviqBz6rkCy12mUSDQX9RdoZf"
)

var decodeMap map[byte]int

func init() {
	decodeMap = make(map[byte]int)
	for i := 0; i < len(ALPHABET); i++ {
		decodeMap[ALPHABET[i]] = i
	}
}

// ----------------- 业务入口 -----------------

// ResolveVideo 根据完整的视频 URL (如 https://www.bilibili.com/video/BV...) 解析下载地址
func ResolveVideo(inputUrl string) (*model.ResolveResult, error) {
	// 1. 解析 BV/AV 号
	avid, err := BvToAv(inputUrl)
	if err != nil {
		return nil, fmt.Errorf("id parse error: %v", err)
	}

	// 2. 获取 CID 和标题
	cid, title, err := GetCid(avid)
	if err != nil {
		return nil, fmt.Errorf("cid fetch error: %v", err)
	}

	// 3. 调用核心获取逻辑
	return fetchStreamWbi(avid, cid, title)
}

// ResolveVideoUrl 根据 BVID 和 CID 直接解析下载地址
func ResolveVideoUrl(bvidStr, cidStr string) (*model.ResolveResult, error) {
	avid, err := BvToAv(bvidStr)
	if err != nil {
		return nil, fmt.Errorf("bvid parse error: %v", err)
	}

	cid, err := strconv.Atoi(cidStr)
	if err != nil {
		return nil, fmt.Errorf("invalid cid: %v", err)
	}

	return fetchStreamWbi(avid, cid, "video_"+cidStr)
}

// ----------------- 核心逻辑提取 -----------------

// fetchStreamWbi 统一的 WBI 签名和流地址获取逻辑
func fetchStreamWbi(avid string, cid int, title string) (*model.ResolveResult, error) {
	// WBI 签名
	params := map[string]string{
		"avid":  avid,
		"cid":   strconv.Itoa(cid),
		"qn":    "0",
		"fnval": "80", // dash
		"fnver": "0",
		"fourk": "1", // 请求 4K
	}

	query, err := SignAndGetWbiQuery(params)
	if err != nil {
		return nil, fmt.Errorf("wbi sign error: %v", err)
	}

	playApiUrl := "https://api.bilibili.com/x/player/wbi/playurl?" + query

	client := &http.Client{}
	req, _ := http.NewRequest("GET", playApiUrl, nil)

	// 关键 Header，伪装成浏览器
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://www.bilibili.com")

	// --- 关键修改：注入 Cookie (SESSDATA) 以获取高画质 ---
	// 从 dao.Store 中读取已登录的 SESSDATA
	sessData := dao.Store.GetSessData()
	if sessData != "" {
		// B站要求 Cookie 格式: SESSDATA=xxx;
		req.Header.Set("Cookie", "SESSDATA="+sessData)
		fmt.Println("ResolveVideo: Using SESSDATA for high quality.")
	} else {
		fmt.Println("ResolveVideo: No SESSDATA found, quality might be limited.")
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	var playResp model.PlayUrlResponse
	json.Unmarshal(bodyBytes, &playResp)

	if playResp.Code != 0 {
		return nil, fmt.Errorf("bilibili api error: %s", playResp.Message)
	}

	if playResp.Data.Dash.Video == nil {
		if len(playResp.Data.Durl) > 0 {
			return &model.ResolveResult{
				VideoUrl: playResp.Data.Durl[0].Url,
				AudioUrl: "",
				FileName: cleanFileName(title),
			}, nil
		}
		return nil, errors.New("no dash video found")
	}

	// 选择最佳流 (最高画质/音质)
	videos := playResp.Data.Dash.Video
	// 按 ID 降序排序 (ID越大画质越高, e.g. 120=4K, 80=1080P)
	sort.Slice(videos, func(i, j int) bool {
		if videos[i].Id != videos[j].Id {
			return videos[i].Id > videos[j].Id
		}
		return videos[i].Bandwidth > videos[j].Bandwidth
	})

	audios := playResp.Data.Dash.Audio
	sort.Slice(audios, func(i, j int) bool {
		return audios[i].Bandwidth > audios[j].Bandwidth
	})

	videoUrl := videos[0].BaseUrl
	audioUrl := ""
	if len(audios) > 0 {
		audioUrl = audios[0].BaseUrl
	}

	// 打印一下选中的清晰度 ID，方便调试
	fmt.Printf("Selected Video Quality ID: %d (Bandwidth: %d)\n", videos[0].Id, videos[0].Bandwidth)

	return &model.ResolveResult{
		VideoUrl: videoUrl,
		AudioUrl: audioUrl,
		FileName: cleanFileName(title),
	}, nil
}

// ----------------- 辅助工具函数 -----------------

func BvToAv(bvidStr string) (string, error) {
	if strings.HasPrefix(strings.ToLower(bvidStr), "av") {
		return bvidStr[2:], nil
	}
	if _, err := strconv.Atoi(bvidStr); err == nil {
		return bvidStr, nil
	}

	re := regexp.MustCompile(`(BV[a-zA-Z0-9]+)`)
	matches := re.FindStringSubmatch(bvidStr)
	var bvid string
	if len(matches) > 1 {
		bvid = matches[1]
	} else if strings.HasPrefix(bvidStr, "BV") {
		bvid = bvidStr
	} else {
		return "", errors.New("invalid BV format")
	}

	runes := []rune(bvid)
	if len(runes) < 12 {
		return "", errors.New("invalid BV length")
	}

	runes[3], runes[9] = runes[9], runes[3]
	runes[4], runes[7] = runes[7], runes[4]

	var temp int64 = 0
	for i := 3; i < len(runes); i++ {
		idx, ok := decodeMap[byte(runes[i])]
		if !ok {
			return "", errors.New("invalid char")
		}
		temp = temp*58 + int64(idx)
	}

	temp = (temp & MASK_CODE) ^ XOR_CODE
	return fmt.Sprintf("%d", temp), nil
}

func GetCid(avid string) (int, string, error) {
	url := "https://api.bilibili.com/x/web-interface/view?aid=" + avid

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var info model.VideoInfoResponse
	json.Unmarshal(body, &info)

	if info.Code != 0 {
		return 0, "", fmt.Errorf("api error: %s", info.Message)
	}

	return info.Data.Cid, info.Data.Title, nil
}

func cleanFileName(name string) string {
	invalid := []string{"<", ">", ":", "\"", "/", "\\", "|", "?", "*"}
	for _, char := range invalid {
		name = strings.ReplaceAll(name, char, "_")
	}
	return name
}
