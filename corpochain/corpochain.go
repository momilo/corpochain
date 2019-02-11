package main

import (
	"github.com/tav/golly/optparse"
)


const logo = `
   _____ ____  _____  _____   ____   _____ _    _          _____ _   _ 
  / ____/ __ \|  __ \|  __ \ / __ \ / ____| |  | |   /\   |_   _| \ | |
 | |   | |  | | |__) | |__) | |  | | |    | |__| |  /  \    | | |  \| |
 | |   | |  | |  _  /|  ___/| |  | | |    |  __  | / /\ \   | | | .   |
 | |___| |__| | | \ \| |    | |__| | |____| |  | |/ ____ \ _| |_| |\  |
  \_____\____/|_|  \_\_|     \____/ \_____|_|  |_/_/    \_\_____|_| \_|
`


func main() {
	cmds := map[string]func([]string, string){
		"genkeys":   cmdGenKeys,
		"genload":   cmdGenLoad,
		"init":      cmdInit,
		"interpret": cmdInterpret,
		"run":       cmdRun,
	}
	info := map[string]string{
		"genkeys":   "Generate new keys for a node",
		"genload":   "Run a node with transaction load",
		"init":      "Initialise a new CORPOCHAIN network",
		"interpret": "Interpret a node's block graph",
		"run":       "Run a node in a CORPOCHAIN network",
	}
	optparse.Commands("btcgo", "0.0.1", cmds, info, logo)
}
