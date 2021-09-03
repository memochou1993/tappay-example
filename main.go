package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"text/template"
	"time"
)

var (
	client = &http.Client{
		Timeout: 10 * time.Second,
	}
	templates = template.Must(template.ParseFiles("index.html"))
	config    Config
)

type Config struct {
	PartnerKey string `yaml:"partner_key"`
	MerchantID string `yaml:"merchant_id"`
}

type Payload struct {
	PartnerKey string `json:"partner_key"`
	Prime      string `json:"prime"`
	Amount     int    `json:"amount"`
	MerchantID string `json:"merchant_id"`
	Details    string `json:"details"`
	Cardholder struct {
		PhoneNumber string `json:"phone_number"`
		Name        string `json:"name"`
		Email       string `json:"email"`
		ZipCode     string `json:"zip_code"`
		Address     string `json:"address"`
		NationalID  string `json:"national_id"`
	} `json:"cardholder"`
}

type Result struct {
	Status            int    `json:"status"`
	Msg               string `json:"msg"`
	Amount            int    `json:"amount"`
	Acquirer          string `json:"acquirer"`
	Currency          string `json:"currency"`
	RecTradeID        string `json:"rec_trade_id"`
	BankTransactionID string `json:"bank_transaction_id"`
	OrderNumber       string `json:"order_number"`
	AuthCode          string `json:"auth_code"`
	CardInfo          struct {
		Issuer      string `json:"issuer"`
		Funding     int    `json:"funding"`
		Type        int    `json:"type"`
		Level       string `json:"level"`
		Country     string `json:"country"`
		LastFour    string `json:"last_four"`
		BinCode     string `json:"bin_code"`
		IssuerZhTw  string `json:"issuer_zh_tw"`
		BankID      string `json:"bank_id"`
		CountryCode string `json:"country_code"`
	} `json:"card_info"`
	TransactionTimeMillis int64 `json:"transaction_time_millis"`
	BankTransactionTime   struct {
		StartTimeMillis string `json:"start_time_millis"`
		EndTimeMillis   string `json:"end_time_millis"`
	} `json:"bank_transaction_time"`
	BankResultCode           string `json:"bank_result_code"`
	BankResultMsg            string `json:"bank_result_msg"`
	CardIdentifier           string `json:"card_identifier"`
	MerchantID               string `json:"merchant_id"`
	IsRbaVerified            bool   `json:"is_rba_verified"`
	TransactionMethodDetails struct {
		TransactionMethodReference string `json:"transaction_method_reference"`
		TransactionMethod          string `json:"transaction_method"`
	} `json:"transaction_method_details"`
}

func init() {
	if err := parseConfig(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	http.HandleFunc("/", Index)
	http.HandleFunc("/api/pay", Pay)
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func parseConfig() error {
	file := "config.yaml"
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(b, &config)
}

func Index(w http.ResponseWriter, r *http.Request) {
	if err := templates.ExecuteTemplate(w, "index.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func Pay(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		response(w, http.StatusOK, nil)
		return
	}
	payload := Payload{
		PartnerKey: config.PartnerKey,
		Amount:     1,
		MerchantID: config.MerchantID,
	}
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		response(w, http.StatusBadRequest, nil)
		return
	}
	b, err := json.Marshal(payload)
	if err != nil {
		response(w, http.StatusInternalServerError, nil)
		return
	}
	resp, err := payByPrime(bytes.NewBuffer(b))
	if err != nil {
		response(w, http.StatusInternalServerError, nil)
		return
	}
	result := Result{}
	if err := json.Unmarshal(resp, &result); err != nil {
		response(w, http.StatusInternalServerError, nil)
		return
	}
	response(w, http.StatusOK, result)
}

func payByPrime(body io.Reader) (b []byte, err error) {
	url := "https://sandbox.tappaysdk.com/tpc/payment/pay-by-prime"
	req, _ := http.NewRequest(http.MethodPost, url, body)
	req.Header.Set("x-api-key", config.PartnerKey)
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("unexpected response code: %v", resp.StatusCode)
		return
	}
	defer closeBody(resp.Body)
	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	return
}

func closeBody(reader io.ReadCloser) {
	if err := reader.Close(); err != nil {
		log.Fatal(err)
	}
}

func response(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.WriteHeader(code)
	if v == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
