package main

import (
	"gitlab.com/elixxir/client/api"
	"gitlab.com/elixxir/client/interfaces/params"
	"io/ioutil"
	"gitlab.com/elixxir/client/single"
	"os"
	jww "github.com/spf13/jwalterweatherman"
	"time"
)

func initClient() (*api.Client, *single.Manager) {
	createClient()

	pass := password
	storeDir := session

	netParams := params.GetDefaultNetwork()
	client, err := api.Login(storeDir, []byte(pass), netParams)
	if err != nil {
		jww.FATAL.Panicf("%+v", err)
	}

	_, err = client.StartNetworkFollower()
	if err != nil {
		jww.FATAL.Panicf("%+v", err)
	}

	// Wait until connected or crash on timeout
	connected := make(chan bool, 10)
	client.GetHealth().AddChannel(connected)
	waitUntilConnected(connected)

	// Make single-use manager and start receiving process
	singleMng := single.NewManager(client)
	client.AddService(singleMng.StartProcesses)

	return client, singleMng
}


func createClient() *api.Client {
	pass := password
	storeDir := session

	//create a new client if none exist
	if _, err := os.Stat(storeDir); os.IsNotExist(err) {
		// Load NDF
		ndfJSON, err := ioutil.ReadFile(ndfPath)
		if err != nil {
			jww.FATAL.Panicf(err.Error())
		}

		err = api.NewClient(string(ndfJSON), storeDir,
			[]byte(pass), "")
		if err != nil {
			jww.FATAL.Panicf("%+v", err)
		}
	}

	netParams := params.GetDefaultNetwork()
	client, err := api.OpenClient(storeDir, []byte(pass), netParams)
	if err != nil {
		jww.FATAL.Panicf("%+v", err)
	}
	return client
}


func waitUntilConnected(connected chan bool) {
	timeoutTimer := time.NewTimer(90 * time.Second)
	isConnected := false
	//Wait until we connect or panic if we can't by a timeout
	for !isConnected {
		select {
		case isConnected = <-connected:
			jww.INFO.Printf("Network Status: %v\n",
				isConnected)
			break
		case <-timeoutTimer.C:
			jww.FATAL.Panic("timeout on connection")
		}
	}

	// Now start a thread to empty this channel and update us
	// on connection changes for debugging purposes.
	go func() {
		prev := true
		for {
			select {
			case isConnected = <-connected:
				if isConnected != prev {
					prev = isConnected
					jww.INFO.Printf(
						"Network Status Changed: %v\n",
						isConnected)
				}
				break
			}
		}
	}()
}