package types

type TxPushMessage struct {
	TxID string `json:"tx_id"`
	Data string `json:"data"`
}

type SignatureMessage struct {
	TxID      string `json:"tx_id"`
	Signature string `json:"signature"`
	Sender    string `json:"sender"`
}
