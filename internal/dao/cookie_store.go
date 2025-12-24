package dao

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// CookieData 对应 Java 中的 Cookie 实体类
// 我们只需要存最关键的 SESSDATA 和 DedeUserID
type CookieData struct {
	SessData   string `json:"sess_data"`
	DedeUserID string `json:"dede_user_id"` // 对应 B 站的 mid
	BiliJct    string `json:"bili_jct"`     // csrf token，有时候需要
	UpdateTime int64  `json:"update_time"`
}

// CookieStore 负责管理 Cookie 的读写
// 替代 MongoDB，我们直接读写本地的一个 json 文件
type CookieStore struct {
	filePath string
	mu       sync.RWMutex // 读写锁，防止并发读写文件冲突
}

var Store *CookieStore // 全局单例

// InitStore 初始化存储
func InitStore(path string) {
	Store = &CookieStore{
		filePath: path,
	}
	// 确保文件存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		_ = os.WriteFile(path, []byte("{}"), 0644)
	}
}

// SaveCookie 保存 Cookie 到文件
func (s *CookieStore) SaveCookie(data CookieData) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data.UpdateTime = time.Now().Unix()

	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, bytes, 0644)
}

// LoadCookie 从文件读取 Cookie
func (s *CookieStore) LoadCookie() (CookieData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var data CookieData
	bytes, err := os.ReadFile(s.filePath)
	if err != nil {
		return data, err
	}

	// 如果文件为空，直接返回空对象
	if len(bytes) == 0 {
		return data, nil
	}

	err = json.Unmarshal(bytes, &data)
	return data, err
}

// GetSessData 快捷获取 SessData 字符串
func (s *CookieStore) GetSessData() string {
	cookie, err := s.LoadCookie()
	if err != nil {
		return ""
	}
	return cookie.SessData
}
