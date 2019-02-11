package core

import (
	"encoding/hex"

	"github.com/boltdb/bolt"

	log "github.com/sirupsen/logrus"
)

const (
	utxoBucket = "UTXOBucket"
)

type UTXOSet struct {
	blockchainSession *Blockchain
}

func NewUTXOSet(blockchainSession *Blockchain) (*UTXOSet, error) {
	UTXSOset := &UTXOSet{blockchainSession}
	bucketName := []byte(utxoBucket)

	err := UTXSOset.blockchainSession.db.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket(bucketName)
		return err
	})
	if err != nil {
		return nil, err
	}
	return UTXSOset, nil
}

func (u UTXOSet) Reindex() error {
	bucketName := []byte(utxoBucket)

	err := u.blockchainSession.db.db.Update(func(tx *bolt.Tx) error {
		if err := tx.DeleteBucket(bucketName); err != nil {
			return err
		}
		_, err := tx.CreateBucket(bucketName)
		return err
	})
	if err != nil {
		return err
	}

	UTXO := u.blockchainSession.FindUTXO()
	err = u.blockchainSession.db.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		for txID, outs := range UTXO {
			key, err := hex.DecodeString(txID)
			if err = b.Put(key, outs.Serialize()); err != nil {
				log.Infoln("Reindexing error while inserting outs for txID=", txID)
			}
		}
		return nil
	})
	return err
}

func (u UTXOSet) GetBalance(address string) (int, error) {
	log.Infoln("GetBalance for address called: ", address)
	pubKeyHash := Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]

	UTXOs, err := u.FindUTXO(pubKeyHash)
	if err != nil {
		return 0, err
	}

	balance := 0
	for _, out := range UTXOs {
		balance += out.Value
	}

	return balance, err
}

func (u UTXOSet) FindUTXO(pubKeyHash []byte) ([]TXOutput, error) {
	var UTXOs []TXOutput

	err := u.blockchainSession.db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			outs := DeserializeOutputs(v)

			for _, out := range outs.Outputs {
				if out.IsLockedWith(pubKeyHash) {
					UTXOs = append(UTXOs, out)
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return UTXOs, nil
}

func (u UTXOSet) Update(block *Block) error {
	log.Infoln("Update called")
	err := u.blockchainSession.db.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))

		for _, blockchainTx := range block.Transactions {
			if blockchainTx.IsCoinbase() == false {
				for _, vin := range blockchainTx.Vin {
					updatedOuts := TXOutputs{}
					outsBytes := b.Get(vin.Txid)
					outs := DeserializeOutputs(outsBytes)

					for outIdx, out := range outs.Outputs {
						if outIdx != vin.Vout {
							updatedOuts.Outputs = append(updatedOuts.Outputs, out)
						}
					}

					if len(updatedOuts.Outputs) == 0 {
						if err := b.Delete(vin.Txid); err != nil {
							tx.Rollback()
							return err
						}
					} else {
						if err := b.Put(vin.Txid, updatedOuts.Serialize()); err != nil {
							tx.Rollback()
							return err
						}
					}
				}
			}

			newOutputs := TXOutputs{}
			for _, out := range blockchainTx.Vout {
				newOutputs.Outputs = append(newOutputs.Outputs, out)
			}

			if err := b.Put(blockchainTx.ID, newOutputs.Serialize()); err != nil {
				tx.Rollback()
				return err
			}
		}
		return nil
	})
	return err
}
