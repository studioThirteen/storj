package UITests

import (
	assert2 "github.com/stretchr/testify/assert"
	"gotest.tools/assert"
	"testing"
	"time"
)
var email= "test@11whqbr4dcdexddedflxcgscsrdkdcfs2kdls3wdhhffledef2.com"
var pass= "123qwe"
var access_grant_name = "qwerty"
func Test_e2e_user_create_login_access_in_browser(t *testing.T)  {
	page, browser := setup_browser()
	defer browser.MustClose()

	// creating new user
	page.MustElement("div.login-container__register-button").MustClick()
	time.Sleep(1*time.Second)
	page.MustElementX("(//*[@class=\"headerless-input\"])[1]").MustWaitStable().MustWaitEnabled().MustInput("test2")
	page.MustElementX("(//*[@class=\"headerless-input\"])[2]").MustInput(email)
	page.MustElementX("(//*[@type=\"password\"])[1]").MustInput(pass)
	page.MustElementX("(//*[@type=\"password\"])[2]").MustInput(pass)
	page.MustElement("span.checkmark").MustClick()
	page.MustElement("div#createAccountButton").MustClick()

	time.Sleep(2*time.Second)
	// checking elements on congrats screen
	header:= page.MustElement("h2.register-success-area__form-container__title").MustText()
	assert.Equal(t,header,"Account Created!")
	text1:= page.MustElement("span.register-success-area__form-container__sub-title").MustText()
	assert.Equal(t,text1,"Check your email to complete registration.")
	countertext:= page.MustElement("p.register-success-area__form-container__text").MustText()
	assert2.Contains(t, countertext, "Didn’t receive a verification email?")
	resendbutton:= page.MustElement("div.register-success-area__form-container__button-container").MustText()
	assert.DeepEqual(t,resendbutton,"Resend Email")
	loginbutton:= page.MustElement("div.register-container__register-button").MustText()
	assert.DeepEqual(t,loginbutton,"Login")

	// continue to login and onboarding flow
	page.MustElement("div.register-container__register-button").MustClick()
	page.MustElement(".headerless-input").MustInput(email)
	page.MustElement("[type=password]").MustInput(pass)
	page.MustElement("div.login-area__submit-area__login-button").MustClick()

	// check welcome to Storj screen elements
	welcomeHeader:= page.MustElement("h1.overview-area__title").MustText()
	assert.DeepEqual(t, welcomeHeader, "Welcome to Storj.\nLet’s Get Started.")
	followtext:= page.MustElement("p.overview-area__sub-title").MustText()
	assert.DeepEqual(t, followtext, "Follow the docs to start storing data using method below.")
	image:= *page.MustElement(".overview-area__steps-area__step__image").MustAttribute("src")
	assert2.Contains(t, image, "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAADcAAAAmCAMA")
	cliTool:= page.MustElement("h2.overview-area__steps-area__step__title").MustText()
	assert.DeepEqual(t,cliTool,"CLI Tool")
	clitext:= page.MustElement("p.overview-area__steps-area__step__info").MustText()
	assert.Equal(t,clitext,"Quickly upload data directly through the command line interface.")
	clilink:= *page.MustElement("a.overview-area__steps-area__step__link").MustAttribute("href")
	assert.DeepEqual(t,clilink,"https://documentation.storj.io/setup/cli")
	nexttext:= page.MustElement("p.overview-area__next-label").MustText()
	assert.DeepEqual(t, nexttext, "Next")
	createAGbutton:= page.MustElement("div.overview-area__cta.container").MustText()
	assert.DeepEqual(t,createAGbutton, "Create an Access Grant")

	// continue to next page
	page.MustElement("div.overview-area__cta.container").MustClick()

	// create an access grant screen - check elements
	createHeader:= page.MustElement("h1.onboarding-access__title").MustText()
	assert.DeepEqual(t, createHeader,"Create an Access Grant")
	createtext:= page.MustElement("p.onboarding-access__sub-title").MustText()
	assert.DeepEqual(t, createtext, "Access Grants are keys that allow access to upload, delete, and view your project’s data.")
	progressbar:= page.MustElement("div.progress-bar").MustVisible()
	assert.Assert(t,progressbar)
	createname:= page.MustElement("h1.name-step__title").MustText()
	assert.DeepEqual(t,createname, "Name Your Access Grant")
	nametext:= page.MustElement("p.name-step__sub-title").MustText()
	assert.DeepEqual(t,nametext, "Enter a name for your new Access grant to get started.")
	accessTitle:= page.MustElement("h3.label-container__main__label").MustText()
	assert.DeepEqual(t, accessTitle,"Access Grant Name")
	accessInput:= *page.MustElement("input.headered-input").MustAttribute("placeholder")
	assert.Equal(t, accessInput,"Enter a name here...")
	nextButton:= page.MustElement("div.container").MustText()
	assert.DeepEqual(t,nextButton,"Next")

	// set access name and continue
	page.MustElement("input.headered-input").MustInput(access_grant_name)
	page.MustElement("div.container").MustClick()

	// Access permission - check elements
	perTitle:= page.MustElement("h1.permissions__title").MustText()
	assert.Equal(t, perTitle, "Access Permissions")
	perText:= page.MustElement("p.permissions__sub-title").MustText()
	assert.Equal(t, perText, "Assign permissions to this Access Grant.")
	amountCheckboxes:= len(page.MustElementsX("//*[@type=\"checkbox\"]"))
	assert.Equal(t,amountCheckboxes,4)
	downloadPerm:= page.MustElementX("(//*[@class=\"permissions__content__left__item__label\"])[1]").MustText()
	assert.Equal(t, downloadPerm, "Download")
	uploadPerm:= page.MustElementX("(//*[@class=\"permissions__content__left__item__label\"])[2]").MustText()
	assert.Equal(t, uploadPerm, "Upload")
	listPerm:= page.MustElementX("(//*[@class=\"permissions__content__left__item__label\"])[3]").MustText()
	assert.Equal(t, listPerm, "List")
	deletePerm:= page.MustElementX("(//*[@class=\"permissions__content__left__item__label\"])[4]").MustText()
	assert.Equal(t, deletePerm, "Delete")
	durationText:= page.MustElement("p.permissions__content__right__duration-select__label").MustText()
	assert.Equal(t, durationText, "Duration")
	durationDrop:= page.MustElement("div.duration-selection__toggle-container")
	assert.Equal(t,durationDrop.MustText(),"Forever")
	// check if datepicker appears
	durationDrop.MustClick()
	time.Sleep(1*time.Second)
	datepicker:= page.MustElement("div.duration-picker").MustVisible()
	assert.Assert(t, datepicker)
	durationDrop.MustClick()
	bucketsText:= page.MustElement("p.permissions__content__right__buckets-select__label").MustText()
	assert.Equal(t, bucketsText, "Buckets")
	bucketsDrop:= page.MustElement("div.buckets-selection")
	assert.Equal(t,bucketsDrop.MustText(),"All")
	conBro:= page.MustElement("div.permissions__button.container")
	assert.Equal(t, conBro.MustText(), "Continue in Browser")
	conCLI:= page.MustElement(".permissions__cli-link")
	assert.Equal(t, conCLI.MustText(), "Continue in CLI")
	// continue in Browser
	conBro.MustClick()
	// check success notification
	broNotification:= page.MustElement("p.notification-wrap__text-area__message").MustText()
	assert.Equal(t, "Permissions were set successfully",broNotification)
	time.Sleep(2*time.Second)
	// encryption passphrase screen elements checking
	passCheckbox:= page.MustElement("input#pass-checkbox")
	passCheckbox.MustClick()
	encrHeader:= page.MustElement("h1.generate-container__title").MustText()
	assert.Equal(t, encrHeader, "Encryption Passphrase")
	encrWarnTitle:= page.MustElement("p.generate-container__warning__header__label").MustText()
	assert.Equal(t, encrWarnTitle, "Save Your Encryption Passphrase")
	encrWarnMessage:= page.MustElement("p.generate-container__warning__message").MustText()
	assert.Equal(t, encrWarnMessage, "You’ll need this passphrase to access data in the future. This is the only time it will be displayed. Be sure to write it down.")
	encrPassType:= page.MustElement("p.generate-container__choosing__label")
	assert.Equal(t,encrPassType.MustText(),"Choose Passphrase Type")
	generateTab:= page.MustElement("p.generate-container__choosing__right__option.left-option")
	assert.Equal(t,generateTab.MustText(),"Generate Phrase")
	createTab:= page.MustElementX("(//*[@class=\"generate-container__choosing__right__option\"])")
	assert.Equal(t,createTab.MustText(),"Create Phrase")
	////checkout to create passphrase tab
	//createTab.MustClick()
	//createTitle:= page.MustElement("h3.label-container__main__label").MustText()
	//assert.Equal(t,createTitle,"Create Your Passphrase")
	//createInput:= *page.MustElement("input.headered-input").MustAttribute("placeholder")
	//assert.Equal(t, createInput,"Strong passphrases contain 12 characters or more")
	//
	//
	//// checkout to Generate passphrase and check success notification
	//generateTab.MustClick()
	passPhrase:= page.MustElement("p.generate-container__value-area__mnemonic__value")
	passPhrase.MustVisible()
	passCopy:= page.MustElement("div.generate-container__value-area__mnemonic__button.container")
	assert.Equal(t, passCopy.MustText(),"Copy")
	passCheckboxText:= page.MustElement("label.generate-container__check-area").MustText()
	assert.DeepEqual(t, passCheckboxText," Yes, I wrote this down or saved it somewhere.")
	passNext:= page.MustElement("div.generate-container__next-button.container")
	assert.Equal(t,passNext.MustText(),"Next")
	time.Sleep(1*time.Second)
	page.MustElement("div.generate-container__next-button.container").MustWaitVisible().MustWaitEnabled().MustWaitInteractable().MustClick()
	notification:= page.MustElement("p.notification-wrap__text-area__message").MustWaitEnabled().MustText()
	assert.Equal(t, "Access Grant was generated successfully",notification)
	// check AG
	agrantTitle:=page.MustElement("h1.generate-grant__title").MustText()
	assert.DeepEqual(t,agrantTitle,"Generate Access Grant")
	agrantWarnTitle:= page.MustElement("h2.generate-grant__warning__header__label").MustText()
	assert.DeepEqual(t,agrantWarnTitle,"This Information is Only Displayed Once")
	agrantWarnMessage:= page.MustElement(".generate-grant__warning__message").MustText()
	assert.Equal(t, agrantWarnMessage, "Save this information in a password manager, or wherever you prefer to store sensitive information.")
	agrantAreaTitle:= page.MustElement("h3.generate-grant__grant-area__label").MustText()
	assert.Equal(t, agrantAreaTitle, "Access Grant")
	agrantKey:= page.MustElement(".generate-grant__grant-area__container__value").MustVisible()
	assert.Assert(t, agrantKey)
	agrantCopy:= page.MustElement("div.generate-grant__grant-area__container__button.container")
	assert.Equal(t,agrantCopy.MustText(),"Copy")
	////check copy notification
	//agrantCopy.MustClick()
	//time.Sleep(1*time.Second)
	//copyNotification:= page.MustElement(".notification-wrap__text-area").MustWaitVisible().MustText()
	//assert.Equal(t, "Token was copied successfully",copyNotification)
	// Gateway credentials droplist
	gateCredDrop:= page.MustElement("h3.generate-grant__gateway-area__toggle__label").MustWaitStable()
	assert.Equal(t, gateCredDrop.MustText(),"Gateway Credentials")
	// open drop-list
	//gateCredDrop.MustClick()
	////check elements
	//gateCredInBeta:= page.MustElement("p.generate-grant__gateway-area__container__beta__message").MustText()
	//assert.Equal(t, gateCredInBeta, "This feature is currently in Beta")
	//gateCredInBetaLink:= page.MustElement("a.generate-grant__gateway-area__container__beta__link")
	//assert.DeepEqual(t, gateCredInBetaLink.MustText(),"Learn More >" )
	//assert.Equal(t, *gateCredInBetaLink.MustAttribute("href"),"https://forum.storj.io/t/gateway-mt-beta-looking-for-testers/11324")
	//gateCredInfoTitle:= page.MustElement(".generate-grant__gateway-area__container__warning__title").MustText()
	//assert.Equal(t, gateCredInfoTitle, "Using Gateway Credentials Enables Server-Side Encryption.")
	//gateCredInfo:= page.MustElement("p.generate-grant__gateway-area__container__warning__disclaimer").MustText()
	//assert.Equal(t,gateCredInfo,"By generating gateway credentials, you are opting in to Server-side encryption")
	//generateButton:= page.MustElement(".generate-grant__gateway-area__container__warning__button.container.blue-white")
	//assert.Equal(t, generateButton.MustText(),"Generate Gateway Credentials")
	// generate gateway credentials
	// generateButton.MustClick()


	// finish access grant generation
	doneButton:= page.MustElement("div.generate-grant__done-button.container")
	assert.Equal(t,doneButton.MustText(),"Done")
	page.MustElementX("(//*[@class=\"label\"])[2]").MustClick()
	time.Sleep(1*time.Second)
	page.MustElementX("(//*[@class=\"navigation-area__item-container__link__title\"])[2]").MustClick()
	createdAGInList:= page.MustElement("p.name").MustText()
	assert.Equal(t,createdAGInList,access_grant_name)


}
