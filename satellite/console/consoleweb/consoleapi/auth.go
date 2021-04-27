// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package consoleapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/zeebo/errs"
	"go.uber.org/zap"

	"storj.io/common/uuid"
	"storj.io/storj/private/post"
	"storj.io/storj/satellite/analytics"
	"storj.io/storj/satellite/console"
	"storj.io/storj/satellite/console/consoleweb/consoleql"
	"storj.io/storj/satellite/console/consoleweb/consolewebauth"
	"storj.io/storj/satellite/mailservice"
	"storj.io/storj/satellite/rewards"
)

var (
	// ErrAuthAPI - console auth api error type.
	ErrAuthAPI = errs.Class("console auth api error")

	// errNotImplemented is the error value used by handlers of this package to
	// response with status Not Implemented.
	errNotImplemented = errs.New("not implemented")
)

// Auth is an api controller that exposes all auth functionality.
type Auth struct {
	log                   *zap.Logger
	ExternalAddress       string
	LetUsKnowURL          string
	TermsAndConditionsURL string
	ContactInfoURL        string
	service               *console.Service
	analytics             *analytics.Service
	mailService           *mailservice.Service
	cookieAuth            *consolewebauth.CookieAuth
	partners              *rewards.PartnersService
}

// NewAuth is a constructor for api auth controller.
func NewAuth(log *zap.Logger, service *console.Service, mailService *mailservice.Service, cookieAuth *consolewebauth.CookieAuth, partners *rewards.PartnersService, analytics *analytics.Service, externalAddress string, letUsKnowURL string, termsAndConditionsURL string, contactInfoURL string) *Auth {
	return &Auth{
		log:                   log,
		ExternalAddress:       externalAddress,
		LetUsKnowURL:          letUsKnowURL,
		TermsAndConditionsURL: termsAndConditionsURL,
		ContactInfoURL:        contactInfoURL,
		service:               service,
		mailService:           mailService,
		cookieAuth:            cookieAuth,
		partners:              partners,
		analytics:             analytics,
	}
}

// Token authenticates user by credentials and returns auth token.
func (a *Auth) Token(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var err error
	defer mon.Task()(&ctx)(&err)

	var tokenRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err = json.NewDecoder(r.Body).Decode(&tokenRequest)
	if err != nil {
		a.serveJSONError(w, err)
		return
	}

	token, err := a.service.Token(ctx, tokenRequest.Email, tokenRequest.Password)
	if err != nil {
		a.log.Info("Error authenticating token request", zap.String("email", tokenRequest.Email), zap.Error(ErrAuthAPI.Wrap(err)))
		a.serveJSONError(w, err)
		return
	}

	a.cookieAuth.SetTokenCookie(w, token)

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(token)
	if err != nil {
		a.log.Error("token handler could not encode token response", zap.Error(ErrAuthAPI.Wrap(err)))
		return
	}
}

// Logout removes auth cookie.
func (a *Auth) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	defer mon.Task()(&ctx)(nil)

	a.cookieAuth.RemoveTokenCookie(w)

	w.Header().Set("Content-Type", "application/json")
}

// Register creates new user, sends activation e-mail.
func (a *Auth) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var err error
	defer mon.Task()(&ctx)(&err)

	var registerData struct {
		FullName         string `json:"fullName"`
		ShortName        string `json:"shortName"`
		Email            string `json:"email"`
		Partner          string `json:"partner"`
		PartnerID        string `json:"partnerId"`
		Password         string `json:"password"`
		SecretInput      string `json:"secret"`
		ReferrerUserID   string `json:"referrerUserId"`
		IsProfessional   bool   `json:"isProfessional"`
		Position         string `json:"position"`
		CompanyName      string `json:"companyName"`
		EmployeeCount    string `json:"employeeCount"`
		HaveSalesContact bool   `json:"haveSalesContact"`
	}

	err = json.NewDecoder(r.Body).Decode(&registerData)
	if err != nil {
		a.serveJSONError(w, err)
		return
	}

	secret, err := console.RegistrationSecretFromBase64(registerData.SecretInput)
	if err != nil {
		a.serveJSONError(w, err)
		return
	}

	if registerData.Partner != "" {
		info, err := a.partners.ByName(ctx, registerData.Partner)
		if err != nil {
			a.log.Warn("Invalid partner name", zap.String("Partner name", registerData.Partner), zap.String("User email", registerData.Email), zap.Error(err))
		} else {
			registerData.PartnerID = info.ID
		}
	}

	user, err := a.service.CreateUser(ctx,
		console.CreateUser{
			FullName:         registerData.FullName,
			ShortName:        registerData.ShortName,
			Email:            registerData.Email,
			PartnerID:        registerData.PartnerID,
			Password:         registerData.Password,
			IsProfessional:   registerData.IsProfessional,
			Position:         registerData.Position,
			CompanyName:      registerData.CompanyName,
			EmployeeCount:    registerData.EmployeeCount,
			HaveSalesContact: registerData.HaveSalesContact,
		},
		secret,
	)
	if err != nil {
		a.serveJSONError(w, err)
		return
	}

	trackCreateUserFields := analytics.TrackCreateUserFields{
		ID:          user.ID,
		AnonymousID: loadSession(r),
		FullName:    user.FullName,
		Email:       user.Email,
		Type:        analytics.Personal,
	}
	if user.IsProfessional {
		trackCreateUserFields.Type = analytics.Professional
		trackCreateUserFields.EmployeeCount = user.EmployeeCount
		trackCreateUserFields.CompanyName = user.CompanyName
		trackCreateUserFields.JobTitle = user.Position
	}
	a.analytics.TrackCreateUser(trackCreateUserFields)

	token, err := a.service.GenerateActivationToken(ctx, user.ID, user.Email)
	if err != nil {
		a.serveJSONError(w, err)
		return
	}

	link := a.ExternalAddress + "activation/?token=" + token
	userName := user.ShortName
	if user.ShortName == "" {
		userName = user.FullName
	}

	a.mailService.SendRenderedAsync(
		ctx,
		[]post.Address{{Address: user.Email, Name: userName}},
		&consoleql.AccountActivationEmail{
			ActivationLink: link,
			Origin:         a.ExternalAddress,
			UserName:       userName,
		},
	)

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(user.ID)
	if err != nil {
		a.log.Error("registration handler could not encode userID", zap.Error(ErrAuthAPI.Wrap(err)))
		return
	}
}

// loadSession looks for a cookie for the session id.
// this cookie is set from the reverse proxy if the user opts into cookies from Storj.
func loadSession(req *http.Request) string {
	sessionCookie, err := req.Cookie("webtraf-sid")
	if err != nil {
		return ""
	}
	return sessionCookie.Value
}

// UpdateAccount updates user's full name and short name.
func (a *Auth) UpdateAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var err error
	defer mon.Task()(&ctx)(&err)

	var updatedInfo struct {
		FullName  string `json:"fullName"`
		ShortName string `json:"shortName"`
	}

	err = json.NewDecoder(r.Body).Decode(&updatedInfo)
	if err != nil {
		a.serveJSONError(w, err)
		return
	}

	if err = a.service.UpdateAccount(ctx, updatedInfo.FullName, updatedInfo.ShortName); err != nil {
		a.serveJSONError(w, err)
	}
}

// GetAccount gets authorized user and take it's params.
func (a *Auth) GetAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var err error
	defer mon.Task()(&ctx)(&err)

	var user struct {
		ID               uuid.UUID `json:"id"`
		FullName         string    `json:"fullName"`
		ShortName        string    `json:"shortName"`
		Email            string    `json:"email"`
		PartnerID        uuid.UUID `json:"partnerId"`
		ProjectLimit     int       `json:"projectLimit"`
		IsProfessional   bool      `json:"isProfessional"`
		Position         string    `json:"position"`
		CompanyName      string    `json:"companyName"`
		EmployeeCount    string    `json:"employeeCount"`
		HaveSalesContact bool      `json:"haveSalesContact"`
	}

	auth, err := console.GetAuth(ctx)
	if err != nil {
		a.serveJSONError(w, err)
		return
	}

	user.ShortName = auth.User.ShortName
	user.FullName = auth.User.FullName
	user.Email = auth.User.Email
	user.ID = auth.User.ID
	user.PartnerID = auth.User.PartnerID
	user.ProjectLimit = auth.User.ProjectLimit
	user.IsProfessional = auth.User.IsProfessional
	user.CompanyName = auth.User.CompanyName
	user.Position = auth.User.Position
	user.EmployeeCount = auth.User.EmployeeCount
	user.HaveSalesContact = auth.User.HaveSalesContact

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(&user)
	if err != nil {
		a.log.Error("could not encode user info", zap.Error(ErrAuthAPI.Wrap(err)))
		return
	}
}

// DeleteAccount authorizes user and deletes account by password.
func (a *Auth) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	defer mon.Task()(&ctx)(&errNotImplemented)

	// We do not want to allow account deletion via API currently.
	a.serveJSONError(w, errNotImplemented)
}

// ChangeEmail auth user, changes users email for a new one.
func (a *Auth) ChangeEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var err error
	defer mon.Task()(&ctx)(&err)

	var emailChange struct {
		NewEmail string `json:"newEmail"`
	}

	err = json.NewDecoder(r.Body).Decode(&emailChange)
	if err != nil {
		a.serveJSONError(w, err)
		return
	}

	err = a.service.ChangeEmail(ctx, emailChange.NewEmail)
	if err != nil {
		a.serveJSONError(w, err)
		return
	}
}

// ChangePassword auth user, changes users password for a new one.
func (a *Auth) ChangePassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var err error
	defer mon.Task()(&ctx)(&err)

	var passwordChange struct {
		CurrentPassword string `json:"password"`
		NewPassword     string `json:"newPassword"`
	}

	err = json.NewDecoder(r.Body).Decode(&passwordChange)
	if err != nil {
		a.serveJSONError(w, err)
		return
	}

	err = a.service.ChangePassword(ctx, passwordChange.CurrentPassword, passwordChange.NewPassword)
	if err != nil {
		a.serveJSONError(w, err)
		return
	}
}

// ForgotPassword creates password-reset token and sends email to user.
func (a *Auth) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var err error
	defer mon.Task()(&ctx)(&err)

	params := mux.Vars(r)
	email, ok := params["email"]
	if !ok {
		err = errs.New("email expected")
		a.serveJSONError(w, err)
		return
	}

	user, err := a.service.GetUserByEmail(ctx, email)
	if err != nil {
		a.serveJSONError(w, err)
		return
	}

	recoveryToken, err := a.service.GeneratePasswordRecoveryToken(ctx, user.ID)
	if err != nil {
		a.serveJSONError(w, err)
		return
	}

	passwordRecoveryLink := a.ExternalAddress + "password-recovery/?token=" + recoveryToken
	cancelPasswordRecoveryLink := a.ExternalAddress + "cancel-password-recovery/?token=" + recoveryToken
	userName := user.ShortName
	if user.ShortName == "" {
		userName = user.FullName
	}

	contactInfoURL := a.ContactInfoURL
	letUsKnowURL := a.LetUsKnowURL
	termsAndConditionsURL := a.TermsAndConditionsURL

	a.mailService.SendRenderedAsync(
		ctx,
		[]post.Address{{Address: user.Email, Name: userName}},
		&consoleql.ForgotPasswordEmail{
			Origin:                     a.ExternalAddress,
			UserName:                   userName,
			ResetLink:                  passwordRecoveryLink,
			CancelPasswordRecoveryLink: cancelPasswordRecoveryLink,
			LetUsKnowURL:               letUsKnowURL,
			ContactInfoURL:             contactInfoURL,
			TermsAndConditionsURL:      termsAndConditionsURL,
		},
	)
}

// ResendEmail generates activation token by userID and sends email account activation email to user.
func (a *Auth) ResendEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var err error
	defer mon.Task()(&ctx)(&err)

	params := mux.Vars(r)
	id, ok := params["id"]
	if !ok {
		a.serveJSONError(w, err)
		return
	}

	userID, err := uuid.FromString(id)
	if err != nil {
		a.serveJSONError(w, err)
		return
	}

	user, err := a.service.GetUser(ctx, userID)
	if err != nil {
		a.serveJSONError(w, err)
		return
	}

	token, err := a.service.GenerateActivationToken(ctx, user.ID, user.Email)
	if err != nil {
		a.serveJSONError(w, err)
		return
	}

	link := a.ExternalAddress + "activation/?token=" + token
	userName := user.ShortName
	if user.ShortName == "" {
		userName = user.FullName
	}

	contactInfoURL := a.ContactInfoURL
	termsAndConditionsURL := a.TermsAndConditionsURL

	a.mailService.SendRenderedAsync(
		ctx,
		[]post.Address{{Address: user.Email, Name: userName}},
		&consoleql.AccountActivationEmail{
			Origin:                a.ExternalAddress,
			ActivationLink:        link,
			TermsAndConditionsURL: termsAndConditionsURL,
			ContactInfoURL:        contactInfoURL,
			UserName:              userName,
		},
	)
}

// serveJSONError writes JSON error to response output stream.
func (a *Auth) serveJSONError(w http.ResponseWriter, err error) {
	w.WriteHeader(a.getStatusCode(err))

	var response struct {
		Error string `json:"error"`
	}

	response.Error = err.Error()

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		a.log.Error("failed to write json error response", zap.Error(ErrAuthAPI.Wrap(err)))
	}
}

// getStatusCode returns http.StatusCode depends on console error class.
func (a *Auth) getStatusCode(err error) int {
	switch {
	case console.ErrValidation.Has(err):
		return http.StatusBadRequest
	case console.ErrUnauthorized.Has(err):
		return http.StatusUnauthorized
	case console.ErrEmailUsed.Has(err):
		return http.StatusConflict
	case errors.Is(err, errNotImplemented):
		return http.StatusNotImplemented
	default:
		return http.StatusInternalServerError
	}
}
