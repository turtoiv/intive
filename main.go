package main

import (
    "fmt"
    "os"
	_ "github.com/go-sql-driver/mysql"
	"intive/funds"
	"net/http"
	"encoding/json"
	"strconv"
	"crypto/sha256"
	"crypto/subtle"
	"sync"
)

var executedPayments = make(map[string]funds.Transaction)
var mutex = &sync.RWMutex{}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("config file not provided")
		return
	} 

	configName := os.Args[1]
	err := funds.InitDB(configName)

    if err != nil {
		fmt.Printf("unable to connect to db")
        panic(err.Error())
    }
	
	http.HandleFunc("/funds", basicAuth(DisplayFunds))
	http.HandleFunc("/transactions", basicAuth(DisplayTransactions))
	http.HandleFunc("/transfer", basicAuth(TransferFunds))
    http.ListenAndServe(":8000", nil)

}

func basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		
		if ok {
			usernameHash := sha256.Sum256([]byte(username))
			passwordHash := sha256.Sum256([]byte(password))
			expectedUsernameHash := sha256.Sum256([]byte("test"))
			expectedPasswordHash := sha256.Sum256([]byte("test"))

			usernameMatch := (subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1)
			passwordMatch := (subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1)

			if usernameMatch && passwordMatch {
				next.ServeHTTP(w, r)
				return
			}
		}

		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

func DisplayFunds(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
        http.Error(w, "Method is not supported.", http.StatusNotFound)
        return
	}

	userIDs, ok := r.URL.Query()["userid"]
	if ok != true {
		fmt.Println("Query param not found")
		http.Error(w, "Query param not found", http.StatusNotFound)
		return
	}
	
	id,err := strconv.Atoi(userIDs[0])
	if err != nil {
		fmt.Println("UserId param should be integer")
		http.Error(w, "UserId should be integer", http.StatusBadRequest)
		return
	}

	amount,err := funds.ListFunds(id)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	s := fmt.Sprintf("%f", amount)

	fmt.Fprintf(w, "%s", s)

}

func DisplayTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
        http.Error(w, "Method is not supported.", http.StatusNotFound)
        return
	}

	userIDs, ok := r.URL.Query()["userid"]
	if ok != true {
		fmt.Println("Query param not found")
		http.Error(w, "Query param not found", http.StatusNotFound)
		return
	}

	id,err := strconv.Atoi(userIDs[0])
	if err != nil {
		fmt.Println("UserId param should be integer")
		http.Error(w, "UserId param should be integer", http.StatusBadRequest)
		return
	}

	transactions, err := funds.ListTransactions(id)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	var response []byte
	if len(transactions) == 0 {
		response = []byte("no transactions found")
	} else {
		response, err = json.Marshal(transactions)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	
	fmt.Fprintf(w, "%s", string(response))
}

func TransferFunds(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
        http.Error(w, "Method is not supported.", http.StatusNotFound)
        return
	}

	//we use this key to provide basic idempotency to the request
	clientKey := r.Header.Get("Key")
	if clientKey == "" {
		fmt.Fprintf(w, "client key is missing")
		http.Error(w, "Client key is missing", http.StatusBadRequest)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var trans funds.Transaction
	err := decoder.Decode(&trans)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//if the transaction was found exit the request
	if executedPayments[clientKey] == trans {
		fmt.Fprintf(w, "successfull transfered the amount")
		return
	}

	err = funds.TransferFunds(trans.Source, trans.Destination, trans.Amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//add transaction to the existing payments
	mutex.Lock()
	executedPayments[clientKey] = trans;
	mutex.Unlock()

	fmt.Fprintf(w, "successfull transfered the amount")
}
