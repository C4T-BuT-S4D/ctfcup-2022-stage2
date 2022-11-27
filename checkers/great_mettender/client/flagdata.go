package client

type FlagData struct {
	Author    string            `json:"author,omitempty"`
	TenderID  string            `json:"tenderID,omitempty"`
	BidToUser map[string]string `json:"bidToUser,omitempty"`
}
