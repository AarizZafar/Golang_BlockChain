package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"path"
	"strconv"

	"github.com/AarizZafar/goblockchain/utils"
	"github.com/AarizZafar/goblockchain/wallet"
)

const tempDir = "C:\\Users\\aariz\\codes\\Golang\\goblockchain\\wallet_server\\templates\\"

type WalletServer struct {
	port    uint16
	gateway string
}

func NewWalletServer(port uint16, gateway string) *WalletServer {
	return &WalletServer{port, gateway}
}

func (ws *WalletServer) Port() uint16 {
	return ws.port
}

func (ws *WalletServer) Gateway() string {
	return ws.gateway
}

func (ws *WalletServer) Index(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		t, _ := template.ParseFiles(path.Join(tempDir, "index.html"))
		t.Execute(w, "")
	default:
		log.Printf("ERROR : Invalid HTTP Method")
	}
}

func (ws * WalletServer) Wallet(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		w.Header().Add("Content-Type","application/json")
		myWallet := wallet.NewWallet()
		m, _ := myWallet.MarshalJSON()
		io.WriteString(w,string(m[:]))
	default:
		w.WriteHeader(http.StatusBadRequest)
		log.Println("Error : Invalid HTTP Method")
	}
}

func (ws *WalletServer) CreateTransaction(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		decoder := json.NewDecoder(req.Body)
		var t wallet.TransactionRequest
		err := decoder.Decode(&t)
		if err != nil {
			log.Printf("ERROR %v", err)
			// io.WriteString(w, string(utils.JsonStatus("fail")))
			io.WriteString(w,"fail")  //////////
			return
		}
		if !t.Validate() {
			log.Println("ERROR: missing field(s)---")
			io.WriteString(w,"fail")      /////////
			return
		}

		/* our sender privat key data will come as a string which is a hex that is 64 bytes string cannot be processed in the back end
		   sender public key caontains both x, y hence 64 + 64 
		   the public and private key have to be converted in a way that golang can understand */

		publicKey := utils.PublicKeyFromString(*t.SenderPublicKey)
		privateKey := utils.PrivateKeyFromString(*t.SenderPrivateKey, publicKey)
		value, err := strconv.ParseFloat(*t.Value, 32)
		if err != nil {
			log.Println("ERROR: parse error")
			io.WriteString(w,"fail") ////////////
			return
		}
	
		value32 := float32(value)
		fmt.Println(publicKey)   // converted string into a data format that can be handeled by go
		fmt.Println(privateKey)
		fmt.Printf("%.1f\n", value32)

		w.Header().Add("Content-Type", "application/json")

	default:
		w.WriteHeader(http.StatusBadRequest)
		log.Println("Error: Invalid HTTP Method")
	}
}

func (ws *WalletServer) Run() {
	http.HandleFunc("/", ws.Index)
	http.HandleFunc("/wallet", ws.Wallet)
	http.HandleFunc("/transaction", ws.CreateTransaction)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+strconv.Itoa(int(ws.Port())), nil))
}
