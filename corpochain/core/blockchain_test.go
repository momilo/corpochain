package core

import (
	"fmt"
	"log"
	"os/exec"
	"testing"
)

func TestWalletCreate(t *testing.T) {
	walletsManager, err := NewWalletsManager("master")
	if err != nil {
		log.Panic(err)
	}

	newWalletAddress, err := walletsManager.CreateWallet()
	if err != nil {
		log.Panic(err)
	}

	newWallet := walletsManager.GetWallet(newWalletAddress)

	fmt.Println("New wallet: public and private keys:", newWallet.PublicKey, newWallet.PrivateKey)
	walletsManager.SessionClose()
}

func TestMasterWallet(t *testing.T) {
	out, err := exec.Command("bash", "-c", "ls -a").Output()
	fmt.Println("ls -a:", string(out))

	// clear wallets database
	if err := exec.Command("bash", "-c", "> wallet_master.Db").Run(); err != nil {
		log.Panic(err)
	}

	// clear blockchain database
	if err := exec.Command("bash", "-c", "> blockchain_master.Db").Run(); err != nil {
		log.Panic(err)
	}

	walletsManager, err := NewWalletsManager("master")
	if err != nil {
		log.Panic(err)
	}

	genesisWalletAddress, err := walletsManager.CreateWallet()
	if err != nil {
		log.Panic(err)
	}

	blockchainSession, err := NewBlockchain(genesisWalletAddress, "master")
	if err != nil {
		log.Panic(err)
	}

	UTXOset, err := NewUTXOSet(blockchainSession)
	if err != nil {
		log.Panic(err)
	}

	if err := UTXOset.Reindex(); err != nil {
		log.Println("reindex failed")
		log.Panic(err)
	}

	balance, err := UTXOset.GetBalance(genesisWalletAddress)
	if err != nil {
		log.Panic(err)
	}
	fmt.Printf("Balance of '%s': %d\n", genesisWalletAddress, balance)

	testWalletAddress, err := walletsManager.CreateWallet()
	if err != nil {
		log.Panic(err)
	}

	fromWallet := walletsManager.GetWallet(genesisWalletAddress)

	tx, err := NewUTXOTransaction(fromWallet, genesisWalletAddress, testWalletAddress, 100, blockchainSession)
	if err != nil {
		log.Panic(err)
	}
	fmt.Println("tx", tx)

	minedBlock, err := blockchainSession.AddBlock([]*Transaction{tx})
	if err != nil {
		log.Panic(err)
	}

	if err := UTXOset.Update(minedBlock); err != nil {
		fmt.Println("unable to update UTXOset")
		return
	}

	// genesis wallet address
	balance, err = UTXOset.GetBalance(genesisWalletAddress)
	if err != nil {
		log.Panic(err)
	}
	fmt.Printf("Balance of '%s': %d\n", genesisWalletAddress, balance)

	// test wallet address
	balance, err = UTXOset.GetBalance(testWalletAddress)
	if err != nil {
		log.Panic(err)
	}
	fmt.Printf("Balance of '%s': %d\n", testWalletAddress, balance)

	walletsManager.SessionClose()
	blockchainSession.SessionClose()
}
