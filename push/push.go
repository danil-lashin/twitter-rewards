package push

import (
	"bytes"
	"encoding/json"
	"github.com/danil-lashin/twitter-rewards/config"
	"io/ioutil"
	"net/http"
)

type MinterPush struct {
	bearer string
	cli    *http.Client
}

type createRequest struct {
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	Coin      string `json:"coin"`
}

func (r *createRequest) Marshal() []byte {
	result, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}

	return result
}

type createResponse struct {
	Data struct {
		Address string `json:"address"`
		Id      string `json:"id"`
	} `json:"data"`
}

type activateResponse struct {
	Data struct {
		Balance struct {
			Amount float64 `json:"amount"`
		} `json:"balance"`
	} `json:"data"`
}

func NewMinterPush(cfg *config.Config) *MinterPush {
	return &MinterPush{bearer: cfg.MinterPushBearer, cli: &http.Client{}}
}

func (p MinterPush) Create(from string, to string, coin string) (address, id string) {
	requestData := createRequest{
		Sender:    from,
		Recipient: to,
		Coin:      coin,
	}

	req, _ := http.NewRequest("POST", "https://api.minterpush.com/create", bytes.NewBuffer(requestData.Marshal()))
	req.Header.Set("Authorization", "Bearer "+p.bearer)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.cli.Do(req)
	if err != nil {
		panic(err)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()

	var pushData createResponse
	if err := json.Unmarshal(body, &pushData); err != nil {
		panic(err)
	}

	return pushData.Data.Address, pushData.Data.Id
}

func (p MinterPush) Activate(id string) float64 {
	req, _ := http.NewRequest("POST", "https://api.minterpush.com/create/"+id, bytes.NewBuffer([]byte(`{}`)))
	req.Header.Set("Authorization", "Bearer "+p.bearer)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.cli.Do(req)
	if err != nil {
		panic(err)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()

	var response activateResponse
	if err := json.Unmarshal(body, &response); err != nil {
		panic(err)
	}

	return response.Data.Balance.Amount
}
