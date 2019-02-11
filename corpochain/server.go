package main

import  (
	"os"
	"corpochain/corpochain/core"

	pb "corpochain/protocol"

	"golang.org/x/net/context"
	log "github.com/sirupsen/logrus"
)

const (
	master = "master"
)

type Node struct {
	blockchainSession *core.Blockchain
	walletsManager    *core.WalletsManager
	UTXOset           *core.UTXOSet
}

func NewServer() *Node {
	walletsManager, err := core.NewWalletsManager(master)
	if err != nil {
		log.Panic(err)
	}

	genesisWalletAddress, err := walletsManager.CreateWallet()
	if err != nil {
		log.Panic(err)
	}

	blockchainSession, err := core.NewBlockchain(genesisWalletAddress, master)
	if err != nil {
		os.Exit(0)
	}

	UTXOset, err := core.NewUTXOSet(blockchainSession)
	if err != nil {
		os.Exit(0)
	}

	if err := UTXOset.Reindex(); err != nil {
		os.Exit(0)
	}

	return &Node{blockchainSession, walletsManager, UTXOset}
}

func (s *Node) ShutDownElegantly() {
	s.blockchainSession.SessionClose()
	s.walletsManager.SessionClose()
}


func (s *Node) Send(ctx context.Context, msg *pb.Transaction) (*pb.Transaction, error) {
	log.Infoln("Received send request ", msg.String())

	go s.send(ctx, msg)

	log.Infoln("Mining for transaction started")
	return msg, nil
}

func (s *Node) GetBalance(ctx context.Context, msg *pb.Address) (*pb.Amount, error) {
	log.Infoln("Received get balance request", msg.String())
	wallet := s.walletsManager.GetWallet(msg.Address)
	if wallet == nil {
		log.Infoln("wallet not found")
		return &pb.Amount{0}, nil
	}

	balanceCh := make(chan pb.Amount, 1)
	go s.getBalance(ctx, msg, balanceCh)
	balance := <-balanceCh

	log.Infoln("GetBalance finished returning: ", balance)
	return &balance, nil
}

func (s *Node) CreateWallet(ctx context.Context, _ *pb.Empty) (*pb.Address, error) {
	log.Infoln("Received create wallet request")

	walletAddressCh := make(chan pb.Address, 1)
	go s.createWallet(ctx, walletAddressCh)
	walletAddress := <-walletAddressCh

	log.Infoln("CreateWallet finished returning: ", walletAddress)
	return &walletAddress, nil
}

func (s *Node) getBalance(ctx context.Context, msg *pb.Address, balanceCh chan pb.Amount) {
	log.Infoln("getBalance called for: ", msg.Address)

	balance, err := s.UTXOset.GetBalance(msg.Address)
	if err != nil {
		log.Infoln("wallet not found")
		return
	}
	log.Infoln("Balance of '%s': %d\n", msg.Address, balance)

	balanceCh <- pb.Amount{Amount: int64(balance)}
}

func (s *Node) send(ctx context.Context, msg *pb.Transaction) {
	if !core.ValidateAddress(msg.FromAddress.Address) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !core.ValidateAddress(msg.ToAddress.Address) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	fromWallet := s.walletsManager.GetWallet(msg.FromAddress.Address)
	if fromWallet == nil {
		log.Infoln("source wallet not found")
		return
	}

	toWallet := s.walletsManager.GetWallet(msg.FromAddress.Address)
	if toWallet == nil {
		log.Infoln("destination wallet not found")
		return
	}

	tx, err := core.NewUTXOTransaction(fromWallet, msg.FromAddress.Address, msg.ToAddress.Address, int(msg.Amount.Amount), s.blockchainSession)
	if err != nil {
		return
	}

	minedBlock, err := s.blockchainSession.AddBlock([]*core.Transaction{tx})
	if err != nil {
		log.Infoln("unable to mine this block")
		return
	}

	if err := s.UTXOset.Update(minedBlock); err != nil {
		log.Infoln("unable to update UTXOset")
		return
	}

	if log.IsLevelEnabled(log.DebugLevel) {
		dumpChain(s.blockchainSession)
	}
}

func (s *Node) createWallet(ctx context.Context, walletAddressCh chan pb.Address) {
	log.Infoln("go routine runs createWallet")

	walletAddress, err := s.walletsManager.CreateWallet()
	if err != nil {
		log.Panic(err)
	}

	walletAddressCh <- pb.Address{Address: walletAddress}
}

func dumpChain(bc *core.Blockchain) {
	log.Infoln("dumpChain called")
	bc.ResetIterator()
	for {
		block := bc.Iterator.Next()
		if len(block.PrevBlockHash) == 0 {
			break
		}
		log.Printf("Previous hash: %x\n Data: %v\n Hash: %x\n", block.PrevBlockHash, block.Transactions, block.Hash)
	}
}
