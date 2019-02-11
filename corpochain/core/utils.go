package core

import (
	"bytes"
	"encoding/gob"
)

func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	gob.NewEncoder(&result).Encode(b)
	return result.Bytes()
}

func DeserializeBlock(d []byte) *Block {
	var block Block
	gob.NewDecoder(bytes.NewReader(d)).Decode(&block)
	return &block
}

func (w *Wallet) Serialize() []byte {
	var result bytes.Buffer
	gob.NewEncoder(&result).Encode(w)
	return result.Bytes()
}

func DeserializeWallet(w []byte) *Wallet {
	var wallet Wallet
	gob.NewDecoder(bytes.NewReader(w)).Decode(&wallet)
	return &wallet
}
