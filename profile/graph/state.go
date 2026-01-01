package graph

// Result 代表社交媒体账户搜索结果。
type Result struct {
	Name   string `json:"name"`
	Link   string `json:"link"`
	Exists bool   `json:"exists"`
}

// State 定义了画像生成图的状态。
type State struct {
	Username         string      `json:"username"`
	UserID           string      `json:"user_id"`            // 从 Tavily 搜索结果中提取的 user ID
	TavilySearchData string      `json:"tavily_search_data"` // Tavily API 返回的原始搜索结果
	SocialData       []Result    `json:"social_data"`
	ProfileData      string      `json:"profile_data"`
	ProfileText      string      `json:"profile_text"`
	LogChan          chan string `json:"-"`
}
