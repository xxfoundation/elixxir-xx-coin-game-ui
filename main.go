package main

import (
	"github.com/dtylman/gowd"
	"github.com/dtylman/gowd/bootstrap"
	"gitlab.com/elixxir/client/interfaces/contact"
	"gitlab.com/elixxir/client/single"
	"time"
)

var password string
var session string
var ndfPath string
var singleMngr *single.Manager
var botContact contact.Contact

var body *gowd.Element

func main() {

	//_, singleMngr = initClient()

	// creates a new bootstrap fluid container
	body = bootstrap.NewContainer(false)

	// add some elements using the object model
	div := bootstrap.NewElement("div", "well")
	row := bootstrap.NewRow(bootstrap.NewColumn(bootstrap.ColumnSmall, 3, div))
	body.AddElement(row)

	row.SetAttribute("style", "font-size:1.5em")

	ethAddr := bootstrap.NewFormInput(bootstrap.InputTypeText, "Ethereum Address:")
	ethAddr.Element.Kids[1].SetAttribute("style", "font-family:'Roboto Mono', 'Courier New', Courier, monospace;")
	sendText := bootstrap.NewFormInput(bootstrap.InputTypeText, "Message:")

	div.AddElement(ethAddr.Element)
	div.AddElement(sendText.Element)

	// add a button
	btn := bootstrap.NewButton(bootstrap.ButtonPrimary, "Send")
	btnEvent := func(sender *gowd.Element, event *gowd.EventElement) {
		btnClicked(sender, event, ethAddr.Element, sendText.Element)
	}
	btn.OnEvent(gowd.OnClick, btnEvent)
	div.AddElement(btn)

	// div.AddHTML(`
	// <label for="fname">Ethereum address:</label><br>
	// <input type="text" id="ethaddr" name="ethaddr"><br>
	// <label for="lname">Message:</label><br>
	// <input type="text" id="message" name="message"><br><br>`, nil)
	//
	// // add a button
	// btn := bootstrap.NewButton(bootstrap.ButtonPrimary, "Send")
	// btn.OnEvent(gowd.OnClick, btnClicked)
	// row.AddElement(bootstrap.NewColumn(bootstrap.ColumnSmall, 3, bootstrap.NewElement("div", "well", btn)))

	// start the ui loop
	gowd.Run(body)
}

// happens when the 'start' button is clicked
func btnClicked(sender *gowd.Element, event *gowd.EventElement, ethAddr, sendText *gowd.Element) {
	div := bootstrap.NewElement("div", "well")
	row := bootstrap.NewRow(bootstrap.NewColumn(bootstrap.ColumnSmall, 3, div))
	body.AddElement(row)

	// adds test to the body
	text := div.AddElement(gowd.NewStyledText("Sending message...", gowd.BoldText))

	// makes the body stop responding to user events
	body.Disable()

	// Send the message
	message := ethAddr.GetValue() + ";" + sendText.GetValue()

	text.SetText(message)

	// Inline function to print message from client to page, callback for upcoming function
	replyFunc := func(payload []byte, err error) {
		if err != nil {
			text.SetText(err.Error())
		} else {
			text.SetText(string(payload))
		}
		sender.SetText("Start")
		body.RemoveElement(text)
		body.Enable()
	}

	err := singleMngr.TransmitSingleUse(botContact, []byte(message),
		"xxCoinGame", 10, replyFunc, 30*time.Second)
	if err != nil {
		text.SetText(err.Error())
		sender.SetText("Start")
		body.RemoveElement(text)
		body.Enable()
	}
}
