package authorization

type Certificate struct {
	SerialNumber string
}

type AuthInfo struct {
	Timestamp  string `json:"timestamp"`
	Nonce      string `json:"nonce"`
	Signature  string `json:"signature"`
	SerialNo   string `json:"serial_no"`
	MerchantID string `json:"merchant_id"`
}
