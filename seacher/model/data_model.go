package model

type IndexDoc struct {
	Id   uint32 `json:"id, omitempty"`
	Text string `json:"text, omitempty"`
	Url  string `json:"url,omitempty"`
}

type StorageIndexDoc struct {
	*IndexDoc
	//doc seg result
	Keys []string `json:"keys,omitempty"`
}

type InvertedIndex struct {
	segResults map[string]int
}

//type ResponseDoc struct {
//	IndexDoc
//	OriginalText string   `json:"originalText,omitempty"`
//	Score        int      `json:"score,omitempty"` //得分
//	Keys         []string `json:"keys,omitempty"`
//}
