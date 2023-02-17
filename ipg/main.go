package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"strconv"
)

var baseUrl = "http://ipg.vandar.local"
var apiKey = "714e6f9c9fc38e37a5cb1b2f87b2c8a9a2273243"

// SendRequestParams contains the parameters for the send request
type SendRequestParams struct {
	APIKey          string `json:"api_key"`
	Amount          int    `json:"amount"`
	CallbackURL     string `json:"callback_url"`
	MobileNumber    string `json:"mobile_number,omitempty"`
	FactorNumber    string `json:"factorNumber,omitempty"`
	Description     string `json:"description,omitempty"`
	NationalCode    string `json:"national_code,omitempty"`
	ValidCardNumber string `json:"valid_card_number,omitempty"`
	Port            string `json:"port,omitempty"`
}

// SendResponseParams is the response of the send request
type SendResponseParams struct {
	Token   string `json:"token"`
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// VerifyParams contains the parameters for the verify request
type VerifyParams struct {
	APIKey string `json:"api_key"`
	Token  string `json:"token"`
}

// VerifyResponse is the response of the verify request
type VerifyResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

type TransactionResult struct {
	Status       int    `json:"status"`
	Amount       string `json:"amount"`
	TransID      int    `json:"transId"`
	RefNumber    string `*json:"refnumber"`
	TrackingCode string `json:"trackingCode"`
	FactorNumber string `json:"factorNumber"`
	Mobile       string `json:"mobile"`
	Description  string `json:"description"`
	CardNumber   string `*json:"cardNumber"`
	CID          string `*json:"CID"`
	CreatedAt    string `json:"createdAt"`
	PaymentDate  string `json:"paymentDate"`
	Code         int    `json:"code"`
	Message      string `json:"message"`
}

func sendRequest(amount int, mobileNumber string, callbackUrl string) (string, error) {
	var response SendResponseParams
	params := SendRequestParams{
		APIKey:          apiKey,
		Amount:          amount,
		CallbackURL:     callbackUrl,
		MobileNumber:    mobileNumber,
		FactorNumber:    strconv.Itoa(rand.Int()),
		Description:     "GO Lang example",
		NationalCode:    "",
		ValidCardNumber: "",
	}
	payload, err := json.Marshal(params)
	if err != nil {
		return "", err
	}
	resp, err := http.Post(baseUrl+"/api/v3/send", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", err
	}
	if response.Status == 1 {
		return response.Token, nil
	}
	return "", errors.New(response.Message)
}

func verifyRequest(token string) (VerifyResponse, error) {
	var response VerifyResponse
	params := VerifyParams{
		APIKey: apiKey,
		Token:  token,
	}
	payload, err := json.Marshal(params)
	if err != nil {
		return response, err
	}

	resp, err := http.Post(baseUrl+"/api/v3/verify", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return response, err
	}

	return response, nil
}

func getTransactionData(token string) (TransactionResult, error) {
	var response TransactionResult
	params := VerifyParams{
		APIKey: apiKey,
		Token:  token,
	}
	payload, err := json.Marshal(params)
	if err != nil {
		return response, err
	}

	resp, err := http.Post(baseUrl+"/api/v3/transaction", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return response, err
	}

	return response, nil
}
func main() {
	// Route 1: Redirect to new location
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serverUrl := fmt.Sprintf("http://%s", r.Host)
		fmt.Fprintf(w, "<a href='%s'>Please Click Here To Pay</a>", serverUrl+"/send")
	})

	http.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
		callbackUrl := fmt.Sprintf("http://%s/callback", r.Host)
		token, err := sendRequest(10000, "09125221691", callbackUrl)
		if err != nil {
			fmt.Println(err.Error())
		}
		http.Redirect(w, r, baseUrl+"/v3/"+token, http.StatusSeeOther)
	})

	// Route 2: Get query string and show in page
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		queryParams := r.URL.Query()
		token := queryParams.Get("token")
		payment_status := queryParams.Get("payment_status")
		fmt.Fprintf(w, "Token: %s\nPayment Status: %s", token, payment_status)
		if payment_status == "FAILED" {
			transactionResult, _ := getTransactionData(token)
			fmt.Fprintf(w, "\nFailed transaction data\n")
			jsonBytes, _ := json.Marshal(transactionResult)
			fmt.Fprintf(w, string(jsonBytes))
			//writeTransactionData(transactionResult, w)
		} else {
			verifyRequest(token)
		}
	})

	// Start the server on a dynamic port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	fmt.Printf("Server listening on http://localhost:%d\n", port)

	err = http.Serve(listener, nil)
	if err != nil {
		panic(err)
	}
}
