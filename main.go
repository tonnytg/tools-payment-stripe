package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stripe/stripe-go/v72/webhook"
	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/paymentintent"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func main() {
	stripe.Key = "sk_xxxyyyzzz"

	fs := http.FileServer(http.Dir("."))
	http.Handle("/", fs)
	http.HandleFunc("/create-payment-intent", handleCreatePaymentIntent)
	http.HandleFunc("/payments", handleWebhook)

	addr := "localhost:9002"
	log.Printf("Listening on %s ...", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

type item struct {
	ID     string `json:"id"`
	Amount int64  `json:"amount"`
}

type createPaymentIntentRequest struct {
	Items      []item `json:"items"` // You can use to calculate the amount of itens
	CourseName string `json:"course_name"`
	CourseId   string `json:"course_id"`
	Price      string `json:"price"`
	Email      string `json:"email"`
}

// calculateOrderAmount calculates the total order amount based on the items
func calculateOrderAmount(items []item) int64 {
	var total int64
	for _, item := range items {
		total += item.Amount
	}
	return total
}

func convertAmountToInt64(amount string) int64 {
	amount = strings.Replace(amount, ".", "", -1)

	amountInt, err := strconv.ParseInt(amount, 10, 64)
	if err != nil {
		log.Printf("strconv.ParseInt: %v", err)
		return -1
	}

	return amountInt
}

func handleCreatePaymentIntent(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		return
	}

	log.Println("handleCreatePaymentIntent")

	var req createPaymentIntentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("json.NewDecoder.Decode: %v", err)
		return
	}

	//Amount:   stripe.Int64(convertOrderAmountToInt64(req.Price)),
	//Amount:   stripe.Int64(calculateOrderAmount(req.Items)),
	amount := convertAmountToInt64(req.Price)
	if amount == -1 {
		writeJSONErrorMessage(w, "Invalid amount", http.StatusBadRequest)
		log.Printf("Invalid amount")
		return
	}

	courseName := req.CourseName
	courseId := req.CourseId
	if courseName == "" {
		writeJSONErrorMessage(w, "Invalid course_name", http.StatusBadRequest)
		log.Printf("Invalid course_name: %s", courseName)
		return
	}
	if courseId == "" {
		writeJSONErrorMessage(w, "Invalid course_id", http.StatusBadRequest)
		log.Printf("Invalid course_id: %s", courseId)
		return
	}

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amount),
		Currency: stripe.String(string(stripe.CurrencyUSD)),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
		ReceiptEmail: stripe.String(req.Email),
		// add courseId
		Metadata: map[string]string{
			"course_name": courseName,
			"course_id":   courseId,
		},
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("pi.New: %v", err)
		return
	}

	writeJSON(w, struct {
		ClientSecret string `json:"clientSecret"`
	}{
		ClientSecret: pi.ClientSecret,
	})
}

func handlePayments(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		return
	}

	log.Println("handlePayments")
	log.Println(r.Method)

	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	newStr := buf.String()
	log.Println(newStr)
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("json.NewEncoder.Encode: %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := io.Copy(w, &buf); err != nil {
		log.Printf("io.Copy: %v", err)
		return
	}
}

func writeJSONError(w http.ResponseWriter, v interface{}, code int) {
	w.WriteHeader(code)
	writeJSON(w, v)
	return
}

func writeJSONErrorMessage(w http.ResponseWriter, message string, code int) {
	resp := struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}{
		Error: struct {
			Message string `json:"message"`
		}{
			Message: message,
		},
	}
	writeJSONError(w, resp, code)
}

func checkSignature(payload []byte, header string) error {
	endpointSecret := os.Getenv("WEBHOOK_SECRET")
	if endpointSecret == "" {
		return fmt.Errorf("WEBHOOK_SECRET environment variable not set")
	}
	event, err := webhook.ConstructEvent(payload, header, endpointSecret)
	if err != nil {
		return fmt.Errorf("webhook.ConstructEvent: %v", err)
	}
	// Handle the event
	switch event.Type {
	case "payment_intent.succeeded":
		//data := event.Data.Object.(map[string]interface{})

		data := event.Data.Object

		log.Printf("PaymentIntent was successful: %v", data["id"])
	default:
		log.Printf("Unhandled event type: %s", event.Type)
	}
	return nil
}

func handleWebhook(w http.ResponseWriter, req *http.Request) {
	const MaxBodyBytes = int64(65536)
	req.Body = http.MaxBytesReader(w, req.Body, MaxBodyBytes)
	payload, err := ioutil.ReadAll(req.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading request body: %v\n", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	// This is your Stripe CLI webhook secret for testing your endpoint locally.
	endpointSecret := "WEBHOOK_SECRET"
	// Pass the request body and Stripe-Signature header to ConstructEvent, along
	// with the webhook signing key.
	event, err := webhook.ConstructEvent(payload, req.Header.Get("Stripe-Signature"),
		endpointSecret)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error verifying webhook signature: %v\n", err)
		w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature
		return
	}
	var amount float64
	if event.Data.Object["amount_received"] != nil {
		amount = event.Data.Object["amount_received"].(float64)
	}
	switch event.Type {
	case "payment_intent.succeeded":
		jsonBody := map[string]interface{}{
			"id":              event.Data.Object["id"],
			"receipt_email":   event.Data.Object["receipt_email"],
			"course_name":     event.Data.Object["metadata"].(map[string]interface{})["course_name"],
			"course_id":       event.Data.Object["metadata"].(map[string]interface{})["course_id"],
			"amount_received": amount,
			"event_type":      event.Type,
			"status":          event.Data.Object["status"],
		}
		// pretty print
		jsonStr, _ := json.MarshalIndent(jsonBody, "", "  ")
		fmt.Fprintf(os.Stdout, "%v\n", string(jsonStr))
		log.Println("Liberando curso")
	case "payment_intent.created":
		fmt.Fprintf(os.Stdout, "PaymentIntent created: %v\n", event.Data.Object["id"])
	case "charge.updated":
		fmt.Fprintf(os.Stdout, "Charge updated: %v\n", event.Data.Object["id"])
	default:
		jsonBody := map[string]interface{}{
			"id":              event.Data.Object["id"],
			"receipt_email":   event.Data.Object["receipt_email"],
			"course_name":     event.Data.Object["metadata"].(map[string]interface{})["course_name"],
			"course_id":       event.Data.Object["metadata"].(map[string]interface{})["course_id"],
			"amount_received": amount,
			"event_type":      event.Type,
			"status":          event.Data.Object["status"],
		}
		// pretty print
		jsonStr, _ := json.MarshalIndent(jsonBody, "", "  ")
		fmt.Fprintf(os.Stdout, "%v\n", string(jsonStr))
	}

	w.WriteHeader(http.StatusOK)
}
