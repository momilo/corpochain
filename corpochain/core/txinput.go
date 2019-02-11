package core

import (
	"bytes"
	"log"
)

type TXInput struct {
	Txid      []byte //this is the id of tx
	Vout      int
	Signature []byte
	PubKey    []byte
}

func (in *TXInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash, err := HashPublicKey(in.PubKey)
	if err != nil {
		log.Println(err)
		return false
	}
	return bytes.Compare(lockingHash, pubKeyHash) == 0
}
