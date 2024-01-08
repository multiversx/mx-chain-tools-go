package domain

type Transaction struct {
	TxHash         string `json:"txHash"`
	GasLimit       int    `json:"gasLimit"`
	GasPrice       int    `json:"gasPrice"`
	GasUsed        int    `json:"gasUsed"`
	MiniBlockHash  string `json:"miniBlockHash"`
	Nonce          int    `json:"nonce"`
	Receiver       string `json:"receiver"`
	ReceiverAssets struct {
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	} `json:"receiverAssets"`
	ReceiverShard int    `json:"receiverShard"`
	Round         int    `json:"round"`
	Sender        string `json:"sender"`
	SenderShard   int    `json:"senderShard"`
	Signature     string `json:"signature"`
	Status        string `json:"status"`
	Value         string `json:"value"`
	Fee           string `json:"fee"`
	Timestamp     int    `json:"timestamp"`
	Data          string `json:"data"`
	Function      string `json:"function"`
	Action        struct {
		Category    string `json:"category"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Arguments   struct {
			Transfers []struct {
				Type     string `json:"type"`
				Name     string `json:"name"`
				Ticker   string `json:"ticker"`
				SvgUrl   string `json:"svgUrl"`
				Token    string `json:"token"`
				Decimals int    `json:"decimals"`
				Value    string `json:"value"`
			} `json:"transfers"`
			Receiver       string   `json:"receiver"`
			FunctionName   string   `json:"functionName"`
			FunctionArgs   []string `json:"functionArgs"`
			ReceiverAssets struct {
				Name string   `json:"name"`
				Tags []string `json:"tags"`
			} `json:"receiverAssets"`
		} `json:"arguments"`
	} `json:"action"`
	Logs struct {
		Id            string `json:"id"`
		Address       string `json:"address"`
		AddressAssets struct {
			Name string   `json:"name"`
			Tags []string `json:"tags"`
		} `json:"addressAssets"`
		Events []struct {
			Identifier string   `json:"identifier"`
			Address    string   `json:"address"`
			Topics     []string `json:"topics"`
			Order      int      `json:"order"`
		} `json:"events"`
	} `json:"logs"`
	Operations []struct {
		Id             string `json:"id"`
		Action         string `json:"action"`
		Type           string `json:"type"`
		EsdtType       string `json:"esdtType"`
		Identifier     string `json:"identifier"`
		Ticker         string `json:"ticker"`
		Name           string `json:"name"`
		Sender         string `json:"sender"`
		Receiver       string `json:"receiver"`
		Value          string `json:"value"`
		Decimals       int    `json:"decimals"`
		SvgUrl         string `json:"svgUrl"`
		ReceiverAssets struct {
			Name string   `json:"name"`
			Tags []string `json:"tags"`
		} `json:"receiverAssets"`
		ValueUSD float64 `json:"valueUSD"`
	} `json:"operations"`
}
