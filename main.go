package main

import (
	"github.com/dtylman/gowd"
	"gitlab.com/elixxir/client/interfaces/contact"
	"gitlab.com/elixxir/client/single"

	"fmt"
	"time"

	"github.com/dtylman/gowd/bootstrap"
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

	//creates a new bootstrap fluid container
	body = bootstrap.NewContainer(false)

	// add some elements using the object model
	div := bootstrap.NewElement("div", "well")
	row := bootstrap.NewRow(bootstrap.NewColumn(bootstrap.ColumnSmall, 3, div))
	body.AddElement(row)

	div.AddHTML(`
	<label for="fname">Ethereum address:</label><br>
	<input type="text" id="ethaddr" name="ethaddr"><br>
	<label for="lname">Message:</label><br>
	<input type="text" id="message" name="message"><br><br>`, nil)

	// add a button
	btn := bootstrap.NewButton(bootstrap.ButtonPrimary, "Send")
	btn.OnEvent(gowd.OnClick, btnClicked)
	row.AddElement(bootstrap.NewColumn(bootstrap.ColumnSmall, 3, bootstrap.NewElement("div", "well", btn)))

	/*
	// add some other elements from HTML
	div.AddHTML(`<div class="dropdown">
	<button class="btn btn-primary dropdown-toggle" type="button" data-toggle="dropdown">Dropdown Example
	<span class="caret"></span></button>
	<ul class="dropdown-menu" id="dropdown-menu">
	<li><a href="#">HTML</a></li>
	<li><a href="#">CSS</a></li>
	<li><a href="#">JavaScript</a></li>
	</ul>
	</div>`, nil)
	
	*/

	//start the ui loop
	gowd.Run(body)
}

// happens when the 'start' button is clicked
func btnClicked(sender *gowd.Element, event *gowd.EventElement) {
	// adds test to the body
	text := body.AddElement(gowd.NewStyledText("Sending message...", gowd.BoldText))

	// makes the body stop responding to user events
	body.Disable()

	ethAddr := body.Find("ethaddr").GetValue()
	sendText := body.Find("message").GetValue()

	//send the message
	message := fmt.Sprintf("%s:%s",ethAddr,sendText)

	// Inline function to print message from client to page, callback for upcoming function
	replyFunc := func(payload []byte, err error){
		if err != nil {
			body.AddHTML(fmt.Sprintf("<textarea readonly style\"width:100%;\">{}</textarea>", err.Error()), nil)
		} else {
			body.AddHTML(fmt.Sprintf("<textarea readonly style\"width:100%;\">{}</textarea>", string(payload)), nil)
		}
		sender.SetText("Start")
		body.RemoveElement(text)
		body.Enable()
	}

	err := singleMngr.TransmitSingleUse(botContact, []byte(message),
		"xxCoinGame", 10, replyFunc, 30*time.Second)
	if err!=nil{
		body.AddHTML(fmt.Sprintf("<textarea readonly style\"width:100%;\">{}</textarea>", string(err.Error())), nil)
		sender.SetText("Start")
		body.RemoveElement(text)
		body.Enable()
	}
}