package model

type IndexDoc struct {
	Id   uint32 `json:"id,omitempty"`
	Text string `json:"text,omitempty"`
	Url  string `json:"url,omitempty"`
}

type StorageIndexDoc struct {
	*IndexDoc
	// doc seg result
	Keys []string `json:"keys,omitempty"`
}

type InvertedIndex struct {
	segResults map[string]int
}

type IndexRelated struct {
	Id      uint32   `json:"id,omitempty"`
	KeyWord string   `json:"keyword,omitempty"`
	Success []string `json:"success,omitempty"`
}

type ResponseDoc struct {
	IndexDoc
	OriginalText string   `json:"originalText,omitempty"`
	Score        float64  `json:"score,omitempty"` // 得分
	Keys         []string `json:"keys,omitempty"`
}

type SearchLog struct {
	Id    uint32 `json:"id,omitempty"`
	Query string `json:"query,omitempty"`
	Time  uint64 `json:"time,omitempty"`
}
