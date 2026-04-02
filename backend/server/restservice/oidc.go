package restservice

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	dbops "isc.org/stork/server/database"
	dbmodel "isc.org/stork/server/database/model"
	dbsession "isc.org/stork/server/database/session"
)

type OIDCControl struct {
	oauth2Config   oauth2.Config
	tokenVerifier  *oidc.IDTokenVerifier
	sessionManager *dbsession.SessionMgr
	db             *dbops.PgDB
}

// this will be configured
var (
	CLIENT_ID          string = ""
	AUTH_ENDPOINT_URL  string = ""
	TOKEN_ENDPOINT_URL string = ""
	CLIENT_SECRET      string = ""
	REDIRECT_URL       string = "http://localhost:8080/oidc/callback" // this should be built based on core rest api settings: TLSCertificate (http/https scheme), Host, Port, BaseURL
)

func NewOIDCControl(db *dbops.PgDB) *OIDCControl {
	ctx := context.Background()

	op, err := oidc.NewProvider(ctx, AUTH_ENDPOINT_URL) // this tries OIDC discovery so AUTH_ENDPOINT_URL is considered the issuer e.g. "https://gitlab.isc.org"
	if err != nil {
		log.Error("OIDC discovery failed")
		// if discovery fails we may want to construct OP config manually
		// opc := oidc.ProviderConfig{
		// 	AuthURL: AUTH_ENDPOINT_URL,
		// 	TokenURL: TOKEN_ENDPOINT_URL,
		// }
	}
	tokenVerifier := op.Verifier(&oidc.Config{
		ClientID: CLIENT_ID,
	})
	oauth2Config := oauth2.Config{
		ClientID:     CLIENT_ID,
		ClientSecret: CLIENT_SECRET,
		RedirectURL:  REDIRECT_URL,
		Endpoint:     op.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "email", "profile"},
	}
	return &OIDCControl{
		oauth2Config:  oauth2Config,
		tokenVerifier: tokenVerifier,
		db:            db,
	}
}

func (ctl *OIDCControl) SetSessionManager(sessionManager *dbsession.SessionMgr) {
	ctl.sessionManager = sessionManager
}

func generateRandBytes(n int) (bytes []byte, err error) {
	bytes = make([]byte, n)
	_, err = rand.Read(bytes)
	if err != nil {
		bytes = nil
		return
	}
	return
}

func generateRandBase64Str() (hash string, err error) {
	bytes, err := generateRandBytes(32)
	if err != nil {
		return
	}
	hash = base64.RawURLEncoding.EncodeToString(bytes)
	return
}

func generatePKCE() (codeVerifier string, codeChallenge string, err error) {
	codeVerifier, err = generateRandBase64Str()
	if err != nil {
		codeVerifier = ""
		return
	}
	hash := sha256.Sum256([]byte(codeVerifier))
	codeChallenge = base64.RawURLEncoding.EncodeToString(hash[:])
	return
}

func (ctl *OIDCControl) LoginHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	state, err := generateRandBase64Str()
	if err != nil {
		http.Error(w, "oidc error generating random state", http.StatusInternalServerError)
		return
	}
	nonce, err := generateRandBase64Str()
	if err != nil {
		http.Error(w, "oidc error generating random nonce", http.StatusInternalServerError)
		return
	}
	codeVerifier, codeChallenge, err := generatePKCE()
	if err != nil {
		http.Error(w, "oidc error generating random PKCE", http.StatusInternalServerError)
		return
	}

	ctl.sessionManager.StoreOidcData(ctx, state, nonce, codeVerifier)

	authURL := ctl.oauth2Config.AuthCodeURL(
		state,
		oidc.Nonce(nonce),
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)

	http.Redirect(w, r, authURL, http.StatusFound)
}

func (ctl *OIDCControl) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	expectedState, expectedNonce, codeVerifier := ctl.sessionManager.GetOidcData(ctx)

	if r.URL.Query().Get("state") != expectedState {
		http.Error(w, "oidc received invalid state", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	token, err := ctl.oauth2Config.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
	if err != nil {
		http.Error(w, "oidc error exchanging token", http.StatusInternalServerError)
		return
	}
	idTokenJWT, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "oidc error missing id_token in token endpoint response", http.StatusInternalServerError)
		return
	}
	idToken, err := ctl.tokenVerifier.Verify(ctx, idTokenJWT)
	if err != nil {
		http.Error(w, "oidc error invalid id_token", http.StatusInternalServerError)
		return
	}

	if idToken.Nonce != expectedNonce {
		http.Error(w, "oidc invalid nonce", http.StatusBadRequest)
		return
	}

	var claims struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	err = idToken.Claims(&claims)
	if err != nil {
		http.Error(w, "oidc error parsing claims", http.StatusInternalServerError)
		return
	}
	// at this point, oidc auth is considered successful
	groups := []*dbmodel.SystemGroup{}
	groups = append(groups, &dbmodel.SystemGroup{
		ID: dbmodel.SuperAdminGroupID,
	})
	systemUser := &dbmodel.SystemUser{
		Login:                  claims.Email,
		Email:                  claims.Email,
		Lastname:               claims.Name,
		Name:                   claims.Name,
		Groups:                 groups,
		AuthenticationMethodID: "oidc",
		ExternalID:             claims.Sub,
		ChangePassword:         false,
	}
	dbSimpleUser, err := dbmodel.GetUserByExternalID(ctl.db, "oidc", claims.Sub)
	if err != nil {
		http.Error(w, "oidc error system user cannot be fetched from db", http.StatusInternalServerError)
		return
	}
	if dbSimpleUser == nil {
		_, err := dbmodel.CreateUser(ctl.db, systemUser)
		if err != nil {
			http.Error(w, "oidc error creating system user in DB ", http.StatusInternalServerError)
			return
		}
		dbSimpleUser, err = dbmodel.GetUserByExternalID(ctl.db, "oidc", claims.Sub)
		if err != nil {
			http.Error(w, "oidc error system user cannot be fetched from db", http.StatusInternalServerError)
			return
		}
		systemUser.ID = dbSimpleUser.ID
	} else {
		systemUser.ID = dbSimpleUser.ID
		_, err = dbmodel.UpdateUser(ctl.db, systemUser)
		if err != nil {
			http.Error(w, "oidc error updating system user in db", http.StatusInternalServerError)
			return
		}
	}
	err = ctl.sessionManager.LoginHandler(ctx, systemUser)
	if err != nil {
		http.Error(w, "oidc error creating session in SM", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}
