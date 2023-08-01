package model

const (
	MsgIDHeartBeat     = 1000
	MsgIDClientAndPush = 1001
	MsgIDSubscribe     = 2000
	MsgIDUnsubscribe   = 2001
	MsgIDStock         = 2003
)

type MsgStruct struct {
	MsgID int    `json:"msg_id"`
	Data  []byte `json:"data"`
}

type PushMsgStruct struct {
	MsgID    int    `json:"msg_id"`
	Receiver string `json:"receiver"`
	Data     []byte `json:"data"`
}

type SubscribeReqData struct {
	ClientID string `json:"client_id"`
	StockID  string `json:"stock_id"`
}

type UnsubscribeReqData struct {
	ClientID string `json:"client_id"`
	StockID  string `json:"stock_id"`
}

type PushStockData struct {
}

type ServerAddr struct {
	IP   string `json:"ip"`
	Port int    `json:"port"`
}

type StockDetail struct {
	Symbol        string `json:"symbol"`
	Name          string `json:"name"`
	EnName        string `json:"engname"`
	LastTrade     string `json:"lasttrade"`
	ChangePercent string `json:"changepercent"`
}

type StockListResult struct {
	Page       int           `json:"page"`
	Num        int           `json:"num"`
	TotalCount int           `json:"totalCount"`
	Data       []StockDetail `json:"data"`
}

type StockListRsp struct {
	ErrorCode int             `json:"error_code"`
	Reason    string          `json:"reason"`
	Result    StockListResult `json:"result"`
}
