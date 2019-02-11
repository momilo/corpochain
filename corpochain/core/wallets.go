package core

import (
	"fmt"

	"github.com/boltdb/bolt"

	log "github.com/sirupsen/logrus"
)

const (
	walletsDbFile = "wallet_%s.Db"
	walletsKey = "wallets"
)

var walletsBucketKey = []byte(walletsKey)

// WalletsManager stores a collection of wallets
type WalletsManager struct {
	Wallets map[string]*Wallet
	Db      *bolt.DB
}

// NewWalletsManager creates WalletsManager and fills it from a file if it exists
func NewWalletsManager(nodeID string) (*WalletsManager, error) {
	log.Infoln("NewWalletsManager called")
	walletsDbFile := fmt.Sprintf(walletsDbFile, nodeID)
	db, err := bolt.Open(walletsDbFile, 0600, nil)
	if err != nil {
		return nil, err
	}

	wallets := WalletsManager{Db: db}
	wallets.Wallets = make(map[string]*Wallet)

	if err := wallets.LoadWallets(); err != nil {
		return nil, err
	}
	log.Infof("NewWalletsManager finished with wallets loaded: %+v", wallets.Wallets)
	return &wallets, err
}

func (ws *WalletsManager) SessionClose() {
	ws.Db.Close()
}

// CreateWallet adds a Wallet to WalletsManager and updates database
func (ws *WalletsManager) CreateWallet() (string, error) {
	log.Infoln("CreateWallet called")

	wallet, _ := NewWallet()
	addressByte, _ := wallet.GetAddress()
	address := fmt.Sprintf("%s", addressByte)
	log.Infoln("generated address: ", address)

	err := ws.Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(walletsBucketKey)
		return b.Put([]byte(address), wallet.Serialize())
	})
	if err != nil {
		return "", err
	}
	log.Infoln("new wallet added to db")

	ws.Wallets[address] = wallet
	log.Infof("wallet added to wallets: %+v\n", ws.Wallets)
	return address, nil
}

// GetWallet returns a Wallet by its address
func (ws WalletsManager) GetWallet(address string) *Wallet {
	return ws.Wallets[address]
}

func (ws WalletsManager) LoadWallets() error {
	if err := ws.CreateBucketIfMissing(); err != nil {
		return err
	}

	log.Infoln("Loading wallets")
	err := ws.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(walletsBucketKey)
		err := b.ForEach(func(k, v []byte) error {
			ws.Wallets[string(k)] = DeserializeWallet(v)
			return nil
		})
		return err
	})
	return err
}

func (ws WalletsManager) CreateBucketIfMissing() error {
	err := ws.Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(walletsBucketKey)
		// create bucket if not present
		if b == nil {
			_, err := tx.CreateBucket(walletsBucketKey)
			log.Println("bucket created")
			return err
		}
		return nil
	})
	return err
}
