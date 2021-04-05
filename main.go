package main

import (
	"github.com/dtylman/gowd"
	"github.com/dtylman/gowd/bootstrap"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/interfaces/contact"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var password = "password"
var session = ".session"
var ndfPath = "ndf.json"
var logPath = "xx-coin-game.log"
var botContactPath = "botContact.bin"
var botContact contact.Contact

var singleMngr SingleManager
var client ClientAPI

var body *gowd.Element

const testMode = false

func main() {

	initLog()

	// creates a new bootstrap fluid container
	body = bootstrap.NewContainer(false)

	// add some elements using the object model
	div := bootstrap.NewElement("div", "well")
	div.SetAttribute("style", "font-size:1.5em;margin-top:25px;")
	body.AddElement(div)

	logo := bootstrap.NewElement("img", "")
	logo.SetAttribute("src", "img/xx_logo.svg")
	logo.SetAttribute("style", "float:right;margin: -10px -10px 0 0;")
	logo.SetAttribute("id", "logo")
	div.AddElement(logo)

	progressBarTitle := bootstrap.NewElement(gowd.Heading4, "")
	progressBarTitle.SetAttribute("style", "text-align:center;")
	progressBarTitle.SetText("Connecting to network")
	spinner := bootstrap.NewElement("span", "spinner-grow")
	spinner.SetAttribute("role", "status")
	spinner.SetAttribute("style", "width: 3rem; height: 3rem;text-align:center;margin:0 auto;")
	spinnerContainer := bootstrap.NewElement("div", "")
	spinnerContainer.SetAttribute("style", "text-align: center")
	spinnerContainer.AddElement(spinner)
	div.AddElement(progressBarTitle)
	div.AddElement(spinnerContainer)

	err := body.Render()
	if err != nil {
		jww.ERROR.Printf("Failed to render body: %+v", err)
	}

	botContact = readBotContact()
	client, singleMngr = initClient(testMode)

	div.RemoveElement(spinnerContainer)
	div.RemoveElement(progressBarTitle)
	progressBarTitle = bootstrap.NewElement(gowd.Heading4, "")
	progressBarTitle.SetAttribute("style", "text-align:center;")
	progressBarTitle.SetText("Generating Keys and Registering with Network")
	div.AddElement(progressBarTitle)
	progressBar := bootstrap.NewProgressBar()
	progressBar.Kids[0].SetAttribute("style", "background:#037281;")
	div.AddElement(progressBar.Element)

	registeredNodes, totalNodes, err := client.GetNodeRegistrationStatus()
	if err != nil {
		jww.FATAL.Panicf("Failed to get node registration status: %+v", err)
	}
	go func() {
		for {
			time.Sleep(100 * time.Millisecond)
			registeredNodes, _, err = client.GetNodeRegistrationStatus()
			if err != nil {
				jww.FATAL.Panicf("Failed to get node registration status: %+v", err)
			}
			max := (totalNodes * 8) / 10
			progressBar.SetText(strconv.Itoa(registeredNodes) + "/" + strconv.Itoa(totalNodes))
			err = progressBar.SetValue(registeredNodes, totalNodes)
			if err != nil {
				jww.ERROR.Printf("Failed to set progress bar value: %+v", err)
			}
			err = body.Render()
			if err != nil {
				jww.ERROR.Printf("Failed to render body: %+v", err)
			}

			if registeredNodes >= max {
				break
			}
		}

		div.RemoveElement(progressBarTitle)
		div.RemoveElement(progressBar.Element)

		printForm(div)
	}()

	// Start the ui loop
	err = gowd.Run(body)
	if err != nil {
		jww.ERROR.Printf("Failed to start ui loop: %+v", err)
	}
}

func printForm(div *gowd.Element) {
	ethAddr := bootstrap.NewFormInput(bootstrap.InputTypeText, "Ethereum Address:")
	ethAddr.Element.Kids[1].SetAttribute("style", "font-family:'Roboto Mono', Consolas, 'Courier New', Courier, monospace;")
	sendText := bootstrap.NewFormInput(bootstrap.InputTypeText, "Message:")

	if testMode {
		ethAddr.SetValue("0x89205A3A3b2A69De6Dbf7f01ED13B2108B2c43e7")
	}

	div.AddElement(ethAddr.Element)
	div.AddElement(sendText.Element)

	// add a button
	btn := bootstrap.NewButton(bootstrap.ButtonPrimary, "Send over xx")
	btn.SetAttribute("style", "background:#037281;background-color:#037281")
	rtnDiv := bootstrap.NewElement("div", "well")
	body.AddElement(rtnDiv)
	btnEvent := func(sender *gowd.Element, event *gowd.EventElement) {
		btnClicked(sender, event, ethAddr.Element.Kids[1], sendText.Element.Kids[1], rtnDiv)
	}
	btn.OnEvent(gowd.OnClick, btnEvent)
	div.AddElement(btn)

	err := body.Render()
	if err != nil {
		jww.ERROR.Printf("Failed to render body: %+v", err)
	}
}

var lastElement, btnMessage, ethAddrFailure *gowd.Element

// happens when the 'start' button is clicked
func btnClicked(sender *gowd.Element, _ *gowd.EventElement, ethAddr,
	sendText *gowd.Element, div *gowd.Element) {

	// makes the body stop responding to user events
	body.Disable()

	if ethAddrFailure != nil {
		ethAddr.Parent.RemoveElement(ethAddrFailure)
	}

	if !validEthereumAddress(ethAddr.GetValue()) {
		ethAddrFailure = bootstrap.NewElement("span", bootstrap.AlertWarning+" very-small")
		ethAddrFailure.SetText("Must be valid Ethereum address.")
		ethAddr.Parent.AddElement(ethAddrFailure)
		return
	}

	sender.SetAttribute("disabled", "")
	spinner := bootstrap.NewElement("span", "spinner-grow")
	spinner.SetAttribute("role", "status")
	spinner.SetAttribute("style", "width: 3rem; height: 3rem;text-align:center;margin:0 auto;")
	div.SetAttribute("style", "text-align: center")
	div.AddElement(spinner)

	if lastElement != nil {
		div.RemoveElement(lastElement)
	}
	if btnMessage != nil {
		sender.Parent.RemoveElement(btnMessage)
	}

	err := body.Render()
	if err != nil {
		jww.ERROR.Printf("Failed to render body: %+v", err)
	}

	// Send the message
	message := ethAddr.GetValue() + ";" + sendText.GetValue()

	defer func() {
		time.Sleep(1 * time.Second)
		sender.RemoveAttribute("disabled")
		div.RemoveElement(spinner)
		div.RemoveAttribute("style")
		err := body.Render()
		if err != nil {
			jww.ERROR.Printf("Failed to render body: %+v", err)
		}
		body.Enable()
	}()

	replyString := make(chan string, 1)

	// Inline function to print message from client to page, callback for upcoming function
	replyFunc := func(payload []byte, err error) {
		sender.Parent.RemoveElement(btnMessage)
		btnMessage = bootstrap.NewElement("span", "")
		var result string
		if err != nil {
			btnMessage.SetText("Failed to receive response.")
			btnMessage.SetAttribute("style", "font-size:0.75em;padding:0.5em;color:#d62424;")
			result = "ERROR: " + err.Error()
		} else {
			btnMessage.SetText("Received response.")
			btnMessage.SetAttribute("style", "font-size:0.75em;padding:0.5em;color:#24d627;")
			result = string(payload)
		}
		replyString <- result
		sender.Parent.AddElement(btnMessage)
	}

	err = singleMngr.TransmitSingleUse(botContact, []byte(message),
		"xxCoinGame", 10, replyFunc, 30*time.Second)
	if err != nil {
		lastElement = div.AddElement(gowd.NewStyledText(err.Error(), gowd.BoldText))

		if btnMessage != nil {
			sender.Parent.RemoveElement(btnMessage)
		}

		btnMessage = bootstrap.NewElement("span", "")
		btnMessage.SetText("Message failed to send.")
		btnMessage.SetAttribute("style", "font-size:0.75em;padding:0.5em;color:#d62424;")
		sender.Parent.AddElement(btnMessage)
		err := body.Render()
		if err != nil {
			jww.ERROR.Printf("Failed to render body: %+v", err)
		}
		return
	} else {

		if btnMessage != nil {
			sender.Parent.RemoveElement(btnMessage)
		}

		time.Sleep(1 * time.Second)
		btnMessage = bootstrap.NewElement("span", "")
		btnMessage.SetAttribute("style", "font-size:0.75em;padding:0.5em;color:#24d627;")
		sender.Parent.AddElement(btnMessage)
		btnMessage.SetText("Message sent successfully. Waiting for response.")
		err = body.Render()
		if err != nil {
			jww.ERROR.Printf("Failed to render body: %+v", err)
		}
	}

	result := <-replyString
	lastElement = div.AddElement(gowd.NewStyledText(result, gowd.BoldText))
	lastElementStyle := "font-size:1.25em;line-height:1.25em;"
	lastElement.SetAttribute("style", lastElementStyle)
	if strings.Contains(result, "ERROR: ") {
		lastElement.SetAttribute("style", lastElementStyle+"color:#d62424;")
	}
}

func validEthereumAddress(address string) bool {
	r, err := regexp.Compile("^0x[0-9a-fA-F]{40}$")
	if err != nil {
		jww.ERROR.Print(err)
	}

	return r.MatchString(address)
}
