package btcgo
//
//import (
//	"encoding/json"
//	"log"
//)
//
//type JSONTransaction struct {
//	fromAddress string `json:"fromAddress"`
//	toAddress string `json:"toAddress"`
//	amount int64 `json:"amount"`
//}
//
//func (m *Transaction) UnmarshalJSON(b []byte) error {
//	var jsonTx JSONTransaction
//	if err := json.Unmarshal(b, &jsonTx); err != nil {
//		return err
//	}
//	log.Printf("unmarshaled: %v\n", jsonTx)
//	m.FromAddress.Address = jsonTx.fromAddress
//	m.ToAddress.Address = jsonTx.toAddress
//	m.Amount.Amount = jsonTx.amount
//
//	return nil
//}
