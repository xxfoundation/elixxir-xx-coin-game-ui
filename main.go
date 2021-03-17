package main

import (
	"github.com/dtylman/gowd"
	"github.com/dtylman/gowd/bootstrap"
	"gitlab.com/elixxir/client/interfaces/contact"
	"gitlab.com/elixxir/client/single"
	"time"
)

var password = "password"
var session = ".session"
var ndfPath = "ndf.json"
var botContactPath = "botContact.bin"
var botContact contact.Contact

var singleMngr *single.Manager

var body *gowd.Element

func main() {

	botContact = readBotContact()

	_, singleMngr = initClient()

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
	btn := bootstrap.NewButton(bootstrap.ButtonPrimary, "Send over xx")
	rtnDiv := bootstrap.NewElement("div", "well")
	rtnRow := bootstrap.NewRow(bootstrap.NewColumn(bootstrap.ColumnSmall, 3, rtnDiv))
	body.AddElement(rtnRow)
	btnEvent := func(sender *gowd.Element, event *gowd.EventElement) {
		btnClicked(sender, event, ethAddr.Element.Kids[1], sendText.Element.Kids[1], rtnDiv)
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

var lastElement *gowd.Element

// happens when the 'start' button is clicked
func btnClicked(sender *gowd.Element, event *gowd.EventElement, ethAddr,
	sendText *gowd.Element, div *gowd.Element) {

	sender.SetAttribute("disabled", "true")
	body.Render()

	if lastElement!=nil{
		div.RemoveElement(lastElement)
	}

	// adds test to the body

	// makes the body stop responding to user events
	body.Disable()

	// Send the message
	message := ethAddr.GetValue() + ";" + sendText.GetValue()

	defer func() {
		sender.RemoveAttribute("disabled")
		body.Render()
		body.Enable()
	}()

	//text.SetText(message)
	replyString := make(chan string)
	// Inline function to print message from client to page, callback for upcoming function
	replyFunc := func(payload []byte, err error) {
		var result string
		if err != nil {
			result = err.Error()
		} else {
			result = string(payload)
		}
		replyString <- result
		//sender.SetText("Start")
		//body.RemoveElement(text)
		//body.Enable()
	}

	err := singleMngr.TransmitSingleUse(botContact, []byte(message),
		"xxCoinGame", 10, replyFunc, 30*time.Second)
	if err != nil {
		//body.Enable()
		lastElement = div.AddElement(gowd.NewStyledText(err.Error(), gowd.BoldText))
		//sender.SetText("Start")
		//body.RemoveElement(text)
		//body.Enable()
		return
	}

	result := <- replyString


	lastElement = div.AddElement(gowd.NewStyledText(result, gowd.BoldText))


}
