package vatsim

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"vatusa-cobalt/config"
	"vatusa-cobalt/db"
	"vatusa-cobalt/dbconn"
)

const (
	CONNECT_BASE_URL_PROD = "https://auth.vatsim.net"
	CONNECT_BASE_URL_DEV  = "https://auth-dev.vatsim.net"
	CONNECT_URI_AUTHORIZE = "/oauth/authorize"
	CONNECT_URI_TOKEN     = "/oauth/token"
	CONNECT_URI_USER      = "/api/user"
	CONNECT_SCOPES        = "full_name email vatsim_details"
	CONNECT_REDIRECT_URI  = "/login/connect"
	COBALT_BASE_URL       = "https://cobalt.vatusa.net"
)

func ConnectBaseURL() string {
	if config.IsDevelopment() {
		return CONNECT_BASE_URL_DEV
	}
	return CONNECT_BASE_URL_PROD
}

func ConnectAuthorizeURL() string {
	return ConnectBaseURL() + CONNECT_URI_AUTHORIZE
}

func ConnectTokenURL() string {
	return ConnectBaseURL() + CONNECT_URI_TOKEN
}

func ConnectUserURL() string {
	return ConnectBaseURL() + CONNECT_URI_USER
}

func ConnectRedirectURI() string {
	if config.IsDevelopment() {
		return "http://localhost:8000/cobalt" + CONNECT_REDIRECT_URI
	}
	return COBALT_BASE_URL + CONNECT_REDIRECT_URI
}

func ConnectAuthorizeParams() string {
	return fmt.Sprintf(
		"?response_type=code&client_id=%s&redirect_uri=%s&scope=%s&prompt=login,consent",
		config.ConnectClientId(),
		ConnectRedirectURI(),
		CONNECT_SCOPES,
	)
}

func ConnectFullURL() string {
	return ConnectAuthorizeURL() + ConnectAuthorizeParams()
}

type ConnectAccessToken struct {
	Scopes       []string `json:"scopes"`
	TokenType    string   `json:"token_type"`
	ExpiresIn    int64    `json:"expires_in"`
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
}

func FetchToken(code string) (*ConnectAccessToken, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", config.ConnectClientId())
	data.Set("client_secret", config.ConnectClientSecret())
	data.Set("code", code)
	data.Set("redirect_uri", ConnectRedirectURI())

	client := &http.Client{}
	req, err := http.NewRequest("POST", ConnectTokenURL(), strings.NewReader(data.Encode()))

	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var token ConnectAccessToken
	err = json.Unmarshal(body, &token)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

type ConnectUserData struct {
	Data ConnectUserDataData `json:"data"`
}

type ConnectUserDataData struct {
	CID      string                  `json:"cid"`
	Personal ConnectUserDataPersonal `json:"personal"`
	Vatsim   ConnectUserDataVatsim   `json:"vatsim"`
	Oauth    ConnectUserDataOauth    `json:"oauth"`
}

type ConnectUserDataOauth struct {
	TokenValid string `json:"token_valid"`
}

type ConnectUserDataPersonal struct {
	NameFirst string `json:"name_first"`
	NameLast  string `json:"name_last"`
	Email     string `json:"email"`
}

type ConnectUserDataVatsim struct {
	Rating      ConnectUserDataRating `json:"rating"`
	PilotRating ConnectUserDataRating `json:"pilotrating"`
	Division    ConnectUserDataUnit   `json:"division"`
	Region      ConnectUserDataUnit   `json:"region"`
	SubDivision ConnectUserDataUnit   `json:"subdivision"`
}

type ConnectUserDataRating struct {
	Id    int    `json:"id"`
	Long  string `json:"long"`
	Short string `json:"short"`
}

type ConnectUserDataUnit struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func FetchUserData(token *ConnectAccessToken) (*ConnectUserDataData, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", ConnectUserURL(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var data ConnectUserData
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}
	return &data.Data, nil
}

func StoreVatsimUserRecordConnect(data *ConnectUserDataData) error {
	cid, err := strconv.Atoi(data.CID)
	if err != nil {
		return errors.New("error extracting cid")
	}
	params := db.UpsertVatsimUserConnectParams{
		Cid:         int64(cid),
		NameFirst:   data.Personal.NameFirst,
		NameLast:    data.Personal.NameLast,
		Email:       data.Personal.Email,
		Rating:      int32(data.Vatsim.Rating.Id),
		Pilotrating: int32(data.Vatsim.PilotRating.Id),
		RegionID:    data.Vatsim.Region.Id,
		DivisionID:  data.Vatsim.Division.Id,
		SubdivisionID: sql.NullString{
			String: data.Vatsim.SubDivision.Id,
			Valid:  true,
		},
	}
	err = dbconn.Queries().UpsertVatsimUserConnect(context.Background(), params)
	if err != nil {
		return err
	}
	err = VatsimUserUpdated(int64(cid))
	if err != nil {
		return err
	}
	return nil
}
