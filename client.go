package main

import (
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/api"
	"gitlab.com/elixxir/client/interfaces"
	"gitlab.com/elixxir/client/interfaces/contact"
	"gitlab.com/elixxir/client/interfaces/params"
	"gitlab.com/elixxir/client/single"
	"gitlab.com/elixxir/client/stoppable"
	"gitlab.com/xx_network/primitives/utils"
	"io/ioutil"
	"os"
	"time"
)

type ClientAPI interface {
	StartNetworkFollower() (<-chan interfaces.ClientError, error)
	GetHealth() interfaces.HealthTracker
	AddService(sp api.ServiceProcess)
	GetNodeRegistrationStatus() (int, int, error)
}

type SingleManager interface {
	StartProcesses() stoppable.Stoppable
	TransmitSingleUse(contact.Contact, []byte, string, uint8, single.ReplyComm,
		time.Duration) error
}

type TestClient struct {
}

func (tc TestClient) StartNetworkFollower() (<-chan interfaces.ClientError, error) {
	return nil, nil
}

func (tc TestClient) GetHealth() interfaces.HealthTracker {
	return nil
}

func (tc TestClient) AddService(api.ServiceProcess) {

}

var NodeRegistrationStatusTrack = 0

func (tc TestClient) GetNodeRegistrationStatus() (int, int, error) {
	NodeRegistrationStatusTrack++
	return NodeRegistrationStatusTrack, 30, nil
}

type TestSingle struct {
}

func (ts TestSingle) StartProcesses() stoppable.Stoppable {
	return nil
}

func (ts TestSingle) TransmitSingleUse(_ contact.Contact, payload []byte,
	_ string, _ uint8, callback single.ReplyComm, _ time.Duration) error {

	go func() {
		time.Sleep(5 * time.Second)
		// callback(payload, errors.New("ERROR"))
		callback(payload, nil)
	}()

	// return errors.New("ERROR")
	return nil
}

func initClient(test bool) (ClientAPI, SingleManager) {

	if test {
		time.Sleep(1 * time.Second)
		return TestClient{}, TestSingle{}
	}

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

	// Create a new client if none exist
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
	// Wait until we connect or panic if we can't by a timeout
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

func readBotContact() contact.Contact {

	// Read from file
	data, err := utils.ReadFile(botContactPath)
	jww.INFO.Printf("Contact file size read in: %d bytes", len(data))
	if err != nil {
		jww.FATAL.Panicf("Failed to read contact file: %+v", err)
	}

	// Unmarshal contact
	c, err := contact.Unmarshal(data)
	if err != nil {
		jww.FATAL.Panicf("Failed to unmarshal contact: %+v", err)
	}

	return c
}

func initLog() {
	jww.SetStdoutOutput(ioutil.Discard)
	logOutput, err := os.OpenFile(logPath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err.Error())
	}
	jww.SetLogOutput(logOutput)
	jww.SetStdoutThreshold(jww.LevelDebug)
	jww.SetLogThreshold(jww.LevelDebug)
}
