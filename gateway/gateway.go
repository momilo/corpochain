package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"sync"
	"time"

	pb "corpochain/protocol"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"golang.org/x/net/context"
	"github.com/gorilla/mux"
	"github.com/tav/golly/optparse"
)

const (
	nodeServiceAddress     = "node-service:1234"
	inititalPoolSize = 10
)

type CorpoChainGateway struct {
	clientPool []CorpoChainClient
}

type corpoChainClient struct {
	connection *grpc.ClientConn
	m        sync.Mutex
	working bool
}

type CorpoChainClient interface {
	Dial() error
	Disconnect()
	GetConnection() pb.BtcgoClient
	Release()
}

func (g *CorpoChainGateway) getClientFromPool() CorpoChainClient {
	for _, client := range g.clientPool {
		if c := client.GetConnection(); c != nil {
			return client
		}
	}
	if client := g.increaseClientPool(); client != nil {
		return client
	} else {
		log.Errorln("unable to increase pool size")
	}
	return nil
}

func (c *corpoChainClient) Dial() error {
	// get grpc client
	conn, err := grpc.Dial(nodeServiceAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Unable to establish connection with node: %v", err)
		return err
	}
	c.connection = conn
	return nil
}

func NewCorpoChainClient() CorpoChainClient {
	return &corpoChainClient{}
}

func (c *corpoChainClient) GetConnection() pb.BtcgoClient {
	if !c.working {
		c.working = true
		return pb.NewBtcgoClient(c.connection)
	} else {
		return nil
	}
}

func (c *corpoChainClient) Disconnect() {
	c.connection.Close()
}

func (c *corpoChainClient) Release() {
	c.working = false
}

func (g *CorpoChainGateway) initialiseClientPool() {
	for i := 0; i < inititalPoolSize; i++ {
		g.increaseClientPool()
	}
}

func (g *CorpoChainGateway) increaseClientPool() CorpoChainClient {
	cpc := NewCorpoChainClient()
	if err := cpc.Dial(); err != nil {
		log.Errorf("Unable to dial to node %+v", err)
		return nil
	}
	g.clientPool = append(g.clientPool, cpc)
	return cpc
}

func (g *CorpoChainGateway) closeClientPool() {
	for _, client := range g.clientPool {
		client.Disconnect()
	}
}

func main() {
	opts := optparse.New("Usage: CORPOCHAIN run NETWORK_NAME NODE_ID [OPTIONS] \n\n")
	consoleLog := opts.Flags("--console-log").Label("LEVEL").String("Set the minimum console log level")
	fileLog := opts.Flags("--file-log").Label("LEVEL").String("Set the minimum file log level")

	if *consoleLog != "" {
		switch *consoleLog {
		case "debug":
			log.SetLevel(log.DebugLevel)
		case "error":
			log.SetLevel(log.ErrorLevel)
		case "fatal":
			log.SetLevel(log.FatalLevel)
		case "info":
			log.SetLevel(log.InfoLevel)
		default:
			log.Fatal("Unknown --console-log level: " + *consoleLog)
		}
	} else {
		log.SetLevel(log.ErrorLevel)
	}
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{})

	if *fileLog != "" {
		switch *fileLog {
		case "debug":
			log.SetLevel(log.DebugLevel)
		case "error":
			log.SetLevel(log.ErrorLevel)
		case "fatal":
			log.SetLevel(log.FatalLevel)
		case "info":
			log.SetLevel(log.InfoLevel)
		default:
			log.Fatal("Unknown --file-log level: " + *fileLog)
		}
		f, err := os.OpenFile(*fileLog, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			return
		}
		log.SetOutput(f)
	}

	gateway := &CorpoChainGateway{}
	gateway.initialiseClientPool()
	defer gateway.closeClientPool()

	r := registerGatewayEndpoints(gateway)
	go http.ListenAndServe(":8080", r)

	killSig := make(chan os.Signal, 1)
	signal.Notify(killSig, syscall.SIGINT, syscall.SIGTERM)

	<-killSig
	log.Infoln("Closing gateway")
}

func registerGatewayEndpoints(gateway *CorpoChainGateway) *mux.Router {
	r := mux.NewRouter()
	log.Infoln("Initialising endpoints")
	r.HandleFunc("/send", gateway.send).Methods("POST")
	r.HandleFunc("/getBalance/{address}", gateway.getBalance).Methods("GET")
	r.HandleFunc("/createWallet", gateway.createWallet).Methods("GET")
	return r
}

func (gateway *CorpoChainGateway) send(w http.ResponseWriter, r *http.Request) {
	var msg pb.Transaction
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(body, &msg); err != nil {
		log.Errorln("unable to unmarshal post request with error: ", err)
	}
	log.Println("Transaction receiver with payload: \n", msg)
	// get context for proxy
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	c := gateway.getClientFromPool()
	defer c.Release()
	if c == nil {
		log.Errorln("Unable to establish connection")
		return
	}

	// call node
	responseTx, err := c.GetConnection().Send(ctx, &msg)
	if err != nil {
		log.Errorln("Send failed: ", err)
		return
	}

	// print response from server
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, responseTx)
	log.Infoln("Received from grpc ", responseTx)
}

func (gateway *CorpoChainGateway) getBalance(w http.ResponseWriter, r *http.Request) {
	log.Infoln("GetBalance called")
	vars := mux.Vars(r)

	msg := pb.Address{Address:vars["address"]}
	log.Infoln("Transaction msg: ", msg)

	// get context for client
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	c := gateway.getClientFromPool()
	defer c.Release()
	if c == nil {
		log.Errorln("Unable to establish connection")
		return
	}

	// call server from client
	balance, err := c.GetConnection().GetBalance(ctx, &msg)
	if err != nil {
		log.Errorln("get balance failed: ",err)
		fmt.Fprintln(w, err.Error())
		return
	}

	// print response from server
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, balance)
	log.Infoln("received from grpc ", balance)
}

func (gateway *CorpoChainGateway) createWallet(w http.ResponseWriter, r *http.Request) {
	log.Infoln("createWallet called")

	// get context for client
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	c := gateway.getClientFromPool()
	defer c.Release()
	if c == nil {
		log.Errorln("Unable to establish connection")
		return
	}

	// call server from client
	walletAddress, err := c.GetConnection().CreateWallet(ctx, &pb.Empty{})
	if err != nil {
		log.Errorln("create wallet failed: ",err)
		fmt.Fprintln(w, err.Error())
		return
	}

	// print response from server
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, walletAddress)
	log.Infoln("received from grpc ", walletAddress)
}