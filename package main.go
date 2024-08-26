package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

type Payment struct {
	ID          int     `json:"id"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Selected    bool    `json:"selected"`
}

type PaymentRequest struct {
	PaymentIDs    []int  `json:"payment_ids"`
	PaymentMethod string `json:"payment_method"`
	ReferenceNote string `json:"reference_note"`
}

var payments = []Payment{
	{ID: 1, Description: "Rent for July", Amount: 2500.00},
	{ID: 2, Description: "Rent for August", Amount: 2500.00},
	{ID: 3, Description: "Rent for September", Amount: 2500.00},
}

var paymentLock = &sync.Mutex{}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/payments", listPayments).Methods("GET")
	r.HandleFunc("/process-payment", processPayment).Methods("POST")

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

// List all payments
func listPayments(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payments)
}

// Process payment based on selected payments and method
func processPayment(w http.ResponseWriter, r *http.Request) {
	var req PaymentRequest

	// Decode the request body into the PaymentRequest struct
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate the request
	if len(req.PaymentIDs) == 0 || req.PaymentMethod == "" {
		http.Error(w, "Payment IDs and Payment Method are required", http.StatusBadRequest)
		return
	}

	if len(req.ReferenceNote) > 120 {
		http.Error(w, "Reference note cannot exceed 120 characters", http.StatusBadRequest)
		return
	}

	var totalAmount float64
	paymentLock.Lock()
	defer paymentLock.Unlock()

	// Process selected payments
	for _, id := range req.PaymentIDs {
		for i := range payments {
			if payments[i].ID == id {
				payments[i].Selected = true
				totalAmount += payments[i].Amount
			}
		}
	}

	// Build the response
	response := map[string]interface{}{
		"message":        "Payment processed successfully",
		"total_amount":   totalAmount,
		"payment_method": req.PaymentMethod,
		"reference_note": req.ReferenceNote,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
