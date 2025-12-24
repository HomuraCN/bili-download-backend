package model

// VideoStreamResponse B站 playurl 接口的响应
type VideoStreamResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    VideoStreamData `json:"data"`
}

type VideoStreamData struct {
	AcceptQuality []int `json:"accept_quality"` // 支持的清晰度列表
	// Dash 格式 (推荐，音视频分离)
	Dash *DashInfo `json:"dash"`
	// Durl 格式 (老旧格式，flv/mp4)
	Durl []DurlInfo `json:"durl"`
}

type DashInfo struct {
	Duration int         `json:"duration"`
	Video    []MediaInfo `json:"video"`
	Audio    []MediaInfo `json:"audio"`
}

type MediaInfo struct {
	Id        int      `json:"id"`        // 清晰度 ID (80=1080P, 64=720P...)
	BaseUrl   string   `json:"baseUrl"`   // 主下载链接
	BackupUrl []string `json:"backupUrl"` // 备用链接
	MimeType  string   `json:"mimeType"`
	Codecid   int      `json:"codecid"` // 编码 (7=AVC, 12=HEVC/H.265)
}

type DurlInfo struct {
	Url string `json:"url"`
}
