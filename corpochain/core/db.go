package core

import (
	"github.com/boltdb/bolt"

	log "github.com/sirupsen/logrus"
)

const (
	dbFile = "blockchain_%s.Db"
	blocksByte = "blocks"
	lastHashKeyByte = "l"
)

var (
	blocksKey []byte
	lastHashKey []byte
)

func init() {
	blocksKey = []byte(blocksByte)
	lastHashKey = []byte(lastHashKeyByte)
}

type dbSession struct {
	db *bolt.DB
}

func NewDbSession(db *bolt.DB) *dbSession {
	return &dbSession{db}
}

func (db *dbSession) Close() {
	db.db.Close()
}

func (db *dbSession) GetLastHash() []byte {
	var lastHash []byte

	db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(blocksKey)
		lastHash = b.Get(lastHashKey)
		return nil
	})

	return lastHash
}

func (db *dbSession) InsertNewBlock(newBlock *Block, lastHash []byte) error {
	err := db.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(blocksKey)
		if err := b.Put(newBlock.Hash, newBlock.Serialize()); err != nil {
			return err
		}
		if err := b.Put(lastHashKey, newBlock.Hash); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (db *dbSession) CreateBlockchain(address string) ([]byte, error) {
	log.Infoln("CreateBlockchain called")
	var tipHash []byte
	cbtx := NewCoinbaseTx(address, genesisCoinbaseData)

	err := db.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(blocksKey)
		if b == nil {
			b, err := tx.CreateBucket(blocksKey)
			if err != nil {
				return err
			}
			genesis := NewGenesisBlock(cbtx)
			if err = b.Put(genesis.Hash, genesis.Serialize()); err != nil {
				return err
			}

			if err = b.Put(lastHashKey, genesis.Hash); err != nil {}
				return err
			tipHash = genesis.Hash
		}
		return nil
	})
	if err != nil {
		return tipHash, err
	}
	return tipHash, nil
}

func (db *dbSession) GetBlockchain() []byte {
	log.Infoln("GetBlockchain called")
	var tipHash []byte
	db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(blocksKey)
		if b != nil {
			tipHash = b.Get(lastHashKey)
		}
		return nil
	})
	log.Infoln("returning tipHash of: ", tipHash)
	return tipHash
}

func (db *dbSession) GetBlockByHash(hash []byte) *Block {
	log.Infoln("GetBlockByHash called")
	var block *Block
	db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(blocksKey)
		encodedBlock := b.Get(hash)
		block = DeserializeBlock(encodedBlock)
		return nil
	})
	return block
}
