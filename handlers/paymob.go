package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	apiKey        = "ZXlKaGJHY2lPaUpJVXpVeE1pSXNJblI1Y0NJNklrcFhWQ0o5LmV5SmpiR0Z6Y3lJNklrMWxjbU5vWVc1MElpd2ljSEp2Wm1sc1pWOXdheUk2TVRBMk5ETTNOaXdpYm1GdFpTSTZJbWx1YVhScFlXd2lmUS5WOTRRejZFUEF2d1h5b3ptenhXQ3JjMko1ZUNfZUJTOTVVdDZ6dnlGSUpna3Z4aDFBaWoxYTFUSnlPNFhwdk11Zmt0c2cxLUNrNjFUNGR6S1FNUmxpdw==" // اجعل هذا سريًا في ملفات .env في مشاريع حقيقية
	integrationID = 5219297
	iframeID      = "944570"
)

type RequestData struct {
	Amount int    `json:"amount"`
	Email  string `json:"email"`
}

func getAuthToken() (string, error) {
	body, _ := json.Marshal(map[string]string{"api_key": apiKey})
	resp, err := http.Post("https://accept.paymob.com/api/auth/tokens", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var res struct {
		Token string `json:"token"`
	}
	json.NewDecoder(resp.Body).Decode(&res)
	return res.Token, nil
}

func createOrder(token string, amount int) (int, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"auth_token":      token,
		"delivery_needed": false,
		"amount_cents":    amount * 100,
		"items":           []interface{}{},
	})
	resp, err := http.Post("https://accept.paymob.com/api/ecommerce/orders", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var res struct {
		ID int `json:"id"`
	}
	json.NewDecoder(resp.Body).Decode(&res)
	return res.ID, nil
}

func getPaymentKey(token string, orderID, amount int, email string) (string, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"auth_token":   token,
		"amount_cents": amount * 100,
		"expiration":   3600,
		"order_id":     orderID,
		"billing_data": map[string]string{
			"first_name":   "Test",
			"last_name":    "User",
			"email":        email,
			"phone_number": "01000000000",
			"apartment":    "NA",
			"floor":        "NA",
			"street":       "NA",
			"building":     "NA",
			"city":         "Cairo",
			"country":      "EG",
			"state":        "Cairo",
		},
		"currency":       "EGP",
		"integration_id": integrationID,
	})

	resp, err := http.Post("https://accept.paymob.com/api/acceptance/payment_keys", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var res struct {
		Token string `json:"token"`
	}
	json.NewDecoder(resp.Body).Decode(&res)
	return res.Token, nil
}

func PayHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(w, r)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqData RequestData
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	json.Unmarshal(body, &reqData)

	authToken, err := getAuthToken()
	if err != nil {
		http.Error(w, "Auth failed", http.StatusInternalServerError)
		return
	}

	orderID, err := createOrder(authToken, reqData.Amount)
	if err != nil {
		http.Error(w, "Order failed", http.StatusInternalServerError)
		return
	}

	paymentToken, err := getPaymentKey(authToken, orderID, reqData.Amount, reqData.Email)
	if err != nil {
		http.Error(w, "Payment key failed", http.StatusInternalServerError)
		return
	}

	iframeURL := fmt.Sprintf("https://accept.paymob.com/api/acceptance/iframes/%s?payment_token=%s", iframeID, paymentToken)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"url": iframeURL})
}

func enableCors(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}
