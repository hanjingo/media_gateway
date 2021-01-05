package gateway

type SearchReq struct {
	Tag   []string `json:"tag"`
	Page  int      `json:"page"`
	Limit int      `json:"limit"`
}

type SearchRsp struct {
	Result uint32 `json:"result"`
}

type SearchDetail struct {
	Tag     []string `json:"tag"`
	Page    int      `json:"page"`
	Limit   int      `json:"limit"`
	MaxPage int      `json:"max_page"`
	Results []string `json:"results"`
}

type NewFileReq struct {
	Hash string   `json:"hash"`
	Tag  []string `json:"tag"`
}

type NewFileRsp struct {
	Result uint32 `json:"result"`
}
