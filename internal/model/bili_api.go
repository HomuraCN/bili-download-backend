package model

// BiliBaseResp B站接口的基础响应结构
type BiliBaseResp struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	TTL     int         `json:"ttl"`
	Data    interface{} `json:"data"`
}

// QRCodeData 实际的数据内容
// 恢复为 snake_case 以匹配前端 Login.vue 的读取方式
type QRCodeData struct {
	Url       string `json:"url"`
	QRCodeKey string `json:"qrcode_key"` // 前端代码写死读取 .qrcode_key
}

// QRCodeResponse 用于嵌套的响应结构
// 前端读取路径为: res.data.data.data.url
// Result(res.data) -> Data(res.data.data) -> Data(res.data.data.data) -> Url
type QRCodeResponse struct {
	Data QRCodeData `json:"data"`
}

// PollData 扫码结果
type PollData struct {
	Url          string `json:"url"`
	RefreshToken string `json:"refresh_token"`
	Timestamp    int64  `json:"timestamp"`
	Code         int    `json:"code"`
	Message      string `json:"message"`
}

type NavData struct {
	WbiImg WbiImg `json:"wbi_img"`
}

type WbiImg struct {
	ImgUrl string `json:"img_url"`
	SubUrl string `json:"sub_url"`
}

// NavResponse 用于获取 WBI img_key 和 sub_key
type NavResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		WbiImg struct {
			ImgUrl string `json:"img_url"`
			SubUrl string `json:"sub_url"`
		} `json:"wbi_img"`
	} `json:"data"`
}

// VideoInfoResponse 用于获取 CID
type VideoInfoResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Bvid  string `json:"bvid"`
		Aid   int    `json:"aid"`
		Cid   int    `json:"cid"`
		Title string `json:"title"`
	} `json:"data"`
}

// PlayUrlResponse 用于获取视频流地址 (DASH格式)
type PlayUrlResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Dash struct {
			Duration int `json:"duration"`
			Video    []struct {
				Id        int    `json:"id"`
				BaseUrl   string `json:"baseUrl"`
				Bandwidth int    `json:"bandwidth"`
				Codecid   int    `json:"codecid"`
			} `json:"video"`
			Audio []struct {
				Id        int    `json:"id"`
				BaseUrl   string `json:"baseUrl"`
				Bandwidth int    `json:"bandwidth"`
			} `json:"audio"`
		} `json:"dash"`
		// 兼容 durl 模式 (虽然主要用 dash)
		Durl []struct {
			Url string `json:"url"`
		} `json:"durl"`
	} `json:"data"`
}

// ResolveResult 返回给前端的最终结果
type ResolveResult struct {
	VideoUrl string `json:"videoUrl"`
	AudioUrl string `json:"audioUrl"`
	FileName string `json:"fileName"`
}
