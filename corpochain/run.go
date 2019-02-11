package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"
	"runtime"
	"runtime/pprof"

	pb "corpochain/protocol"

	"google.golang.org/grpc"
	"golang.org/x/net/context"
	"github.com/tav/golly/optparse"
	"github.com/tav/golly/process"
	log "github.com/sirupsen/logrus"
)

const (
	network = "tcp"
	defaultNodeAddress = "1234"
)

func cmdRun(args []string, usage string) {
	opts := optparse.New("Usage: corpochain run NETWORK_NAME NODE_ID [OPTIONS] \n\n  " + usage + "\n")
	consoleLog := opts.Flags("--console-log").Label("LEVEL").String("Set the minimum console log level")
	cpuProfile := opts.Flags("--cpu-profile").Label("PATH").String("Write a CPU profile to the given file before exiting")
	fileLog := opts.Flags("--file-log").Label("LEVEL").String("Set the minimum file log level")
	memProfile := opts.Flags("--mem-profile").Label("PATH").String("Write the memory profile to the given file before exiting")

	if *cpuProfile != "" {
		profileFile, err := os.Create(*cpuProfile)
		if err != nil {
			log.Fatal("Could not create CPU profile file", *cpuProfile, err)
		}
		pprof.StartCPUProfile(profileFile)
	}

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

	s, stopNodeFunc := startNode()
	killSig := make(chan os.Signal, 1)
	signal.Notify(killSig, syscall.SIGINT, syscall.SIGTERM)

	process.SetExitHandler(func() {
		if *memProfile != "" {
			f, err := os.Create(*memProfile)
			if err != nil {
				log.Fatal("Could not create memory profile file", *memProfile)
			}
			runtime.GC()
			if err := pprof.WriteHeapProfile(f); err != nil {
				log.Fatal("Could not write memory profile", err)
			}
			f.Close()
		}
		if s != nil {
			killSig <- syscall.SIGINT
		}
		if *cpuProfile != "" {
			pprof.StopCPUProfile()
		}
	})

	<-killSig
	stopNodeFunc()
	log.Println("Node shutted down")
}

func startNode() (*grpc.Server, context.CancelFunc) {
	log.Println("Starting the node")
	s := grpc.NewServer()
	nodeServer := NewServer()
	defer nodeServer.ShutDownElegantly()

	log.Println("Registering the node")
	pb.RegisterBtcgoServer(s, nodeServer)

	lis, err := net.Listen(network, ":"+defaultNodeAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	ctx, stopNodeFunc := context.WithCancel(context.Background())
	go func(ctx context.Context) {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}(ctx)

	log.Println("Node is up")
	return s, stopNodeFunc
}

func cmdGenKeys(args []string, usage string) {
	opts := optparse.New("Usage: btcgo run NETWORK_NAME NODE_ID [OPTIONS] \n\n  " + usage + "\n")
	consoleLog := opts.Flags("--console-log").Label("LEVEL").String("Set the minimum console log level")
	fileLog := opts.Flags("--file-log").Label("LEVEL").String("Set the minimum file log level")

	// TODO: cmdGenKeys
	log.Fatal("cmdGenKeys not implemented ", *consoleLog, *fileLog)
}

func cmdGenLoad(args []string, usage string) {
	opts := optparse.New("Usage: btcgo run NETWORK_NAME NODE_ID [OPTIONS] \n\n  " + usage + "\n")
	consoleLog := opts.Flags("--console-log").Label("LEVEL").String("Set the minimum console log level")
	fileLog := opts.Flags("--file-log").Label("LEVEL").String("Set the minimum file log level")

	// TODO: cmdGenLoad
	log.Fatal("cmdGenLoad not implemented ", *consoleLog, *fileLog)
}

func cmdInit(args []string, usage string) {
	opts := optparse.New("Usage: btcgo run NETWORK_NAME NODE_ID [OPTIONS] \n\n  " + usage + "\n")
	consoleLog := opts.Flags("--console-log").Label("LEVEL").String("Set the minimum console log level")
	fileLog := opts.Flags("--file-log").Label("LEVEL").String("Set the minimum file log level")

	// TODO: cmdInit
	log.Fatal("cmdInit not implemented ", *consoleLog, *fileLog)
}

func cmdInterpret(args []string, usage string) {
	opts := optparse.New("Usage: btcgo run NETWORK_NAME NODE_ID [OPTIONS] \n\n  " + usage + "\n")
	consoleLog := opts.Flags("--console-log").Label("LEVEL").String("Set the minimum console log level")
	fileLog := opts.Flags("--file-log").Label("LEVEL").String("Set the minimum file log level")

	// TODO: cmdInterpret
	log.Fatal("cmdInterpret not implemented ", *consoleLog, *fileLog)
}
