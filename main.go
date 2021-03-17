package main

import (
	"github.com/dtylman/gowd"

	"fmt"
	"time"

	"github.com/dtylman/gowd/bootstrap"
)

var body *gowd.Element

func main() {
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
	// adds a text and progress bar to the body
	sender.SetText("Working...")
	text := body.AddElement(gowd.NewStyledText("Working...", gowd.BoldText))
	progressBar := bootstrap.NewProgressBar()
	body.AddElement(progressBar.Element)

	// makes the body stop responding to user events
	body.Disable()

	// clean up - remove the added elements
	defer func() {
		sender.SetText("Start")
		body.RemoveElement(text)
		body.RemoveElement(progressBar.Element)
		body.Enable()
	}()

	// render the progress bar
	for i := 0; i <= 123; i++ {
		progressBar.SetValue(i, 123)
		text.SetText(fmt.Sprintf("Working %v", i))
		time.Sleep(time.Millisecond * 20)
		// this will cause the body to be refreshed
		body.Render()
	}

	body.AddHTML(`<textarea readonly style="width:100%;">Client output would go here</textarea>`, nil)

}
