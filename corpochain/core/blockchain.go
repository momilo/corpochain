package core

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
	log "github.com/sirupsen/logrus"
)

type Block struct {
	Timestamp     int64
	Transactions  []*Transaction
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

type Blockchain struct {
	tip      []byte
	db       *dbSession
	Iterator *BlockchainIterator
}

type BlockchainIterator struct {
	currentHash []byte
	db          *dbSession
}

func NewBlockchain(address, nodeID string) (*Blockchain, error) {
	log.Infoln("NewBlockchain called")
	dbFile := fmt.Sprintf(dbFile, nodeID)
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Errorln("failed to open database")
		return nil, err
	}

	dbSession := NewDbSession(db)
	tipHash := dbSession.GetBlockchain()
	if tipHash == nil {
		tipHash, err = dbSession.CreateBlockchain(address)
		if err != nil {
			return nil, err
		}
	}

	iterator := &BlockchainIterator{tipHash, dbSession}

	blockchain := &Blockchain{tipHash, dbSession, iterator}

	return blockchain, nil
}

func (bc *Blockchain) SessionClose() {
	bc.db.Close()
}

func (i *BlockchainIterator) Next() *Block {
	block := i.db.GetBlockByHash(i.currentHash)
	i.currentHash = block.PrevBlockHash
	return block
}

func (bc *Blockchain) ResetIterator() {
	bc.Iterator = &BlockchainIterator{bc.tip, bc.db}
}

func (bc *Blockchain) AddBlock(transactions []*Transaction) (*Block, error) {
	lastHash := bc.db.GetLastHash()

	for _, tx := range transactions {
		valid, err := bc.VerifyTransaction(tx)
		if err != nil {
			return nil, err
		}
		if valid != true {
			log.Errorln("Invalid Transaction")
			return nil, errors.New("Invalid Transaction")
		}
	}

	newBlock := MineBlock(transactions, lastHash)

	if err := bc.db.InsertNewBlock(newBlock, lastHash); err != nil {
		return nil, err
	}
	bc.tip = newBlock.Hash
	return newBlock, nil
}

func MineBlock(transactions []*Transaction, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), transactions, prevBlockHash, []byte{}, 0}

	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	//validate proof of work
	log.Debugln("PoW: %s\n\n", strconv.FormatBool(NewProofOfWork(block).Validate()))

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

func NewGenesisBlock(coinbase *Transaction) *Block {
	return MineBlock([]*Transaction{coinbase}, []byte{})
}

func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]
}

func (bc *Blockchain) FindUnspentTransactions(pubKeyHash []byte) []Transaction {
	log.Infoln("FindUnspentTransactions called")

	var unspentTXs []Transaction
	spentTXOs := make(map[string][]int)
	bc.ResetIterator()
	defer bc.ResetIterator()

	for {
		block := bc.Iterator.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				// Was the output spent?
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				log.Infoln("looking at: ", out.Value, out.PubKeyHash)
				if out.IsLockedWith(pubKeyHash) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}

			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					if in.UsesKey(pubKeyHash) {
						inTxID := hex.EncodeToString(in.Txid)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return unspentTXs
}

func (bc *Blockchain) FindUTXO() map[string]TXOutputs {
	log.Infoln("FindUTXO called")
	UTXO := make(map[string]TXOutputs)
	spentTXOs := make(map[string][]int)
	bc.ResetIterator()
	defer bc.ResetIterator()

	for {
		block := bc.Iterator.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				// Was the output spent?
				if spentTXOs[txID] != nil {
					for _, spentOutIdx := range spentTXOs[txID] {
						if spentOutIdx == outIdx {
							continue Outputs
						}
					}
				}

				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}

			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					inTxID := hex.EncodeToString(in.Txid)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return UTXO
}

func (bc *Blockchain) FindTransactionByID(ID []byte) (*Transaction, error) {
	log.Infoln("FindTransactionByID called")
	bc.ResetIterator()
	defer bc.ResetIterator()
	for {
		block := bc.Iterator.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return tx, nil
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return nil, errors.New("Transaction is not found")
}

func (bc *Blockchain) SignTransaction(tx *Transaction, privateKey ecdsa.PrivateKey) error {
	prevTxs := make(map[string]*Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransactionByID(vin.Txid)
		if err != nil {
			return err
		}
		prevTxs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privateKey, prevTxs)
	return nil
}

func (bc *Blockchain) VerifyTransaction(tx *Transaction) (bool, error) {
	prevTXs := make(map[string]*Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransactionByID(vin.Txid)
		if err != nil {
			return false, err
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs), nil
}

func (bc *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	log.Infoln("FindSpendableOutputs called")

	pubKeyHash := Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]

	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTransactions(pubKeyHash)
	accumulated := 0

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Vout {
			if out.IsLockedWith(pubKeyHash) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOutputs
}
