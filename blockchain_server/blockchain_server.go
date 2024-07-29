package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/AarizZafar/goblockchain/block"
	"github.com/AarizZafar/goblockchain/utils"
	"github.com/AarizZafar/goblockchain/wallet"
)

/*
we do not want to create a complete block chain again when ever we create a new transaction

	once its made we want to tstore it in out cache
*/
var cache map[string]*block.Blockchain = make(map[string]*block.Blockchain)

/* Creates a variable named cache that is a map (key value pair)
key - string
values (is a pointer)- block.Blockchain */

type BlockchainServer struct {
	port uint16
}

func NewBlockChainServer(port uint16) *BlockchainServer {
	return &BlockchainServer{port}
}

func (bcs *BlockchainServer) Port() uint16 {
	return bcs.port
}

func (bcs *BlockchainServer) GetBlockchain() *block.Blockchain {
	bc, ok := cache["blockchain"] // checking if we have the blockchain in our cache or not
	if !ok {                      // at the very begining the cache is empty
		/* when we dont have any we register the miners address */
		minersWallet := wallet.NewWallet()
		// when we generate a new block we will be registering the miners address
		// bcs.Port will be used to reserch the surrounding block chain servers and to be in sync with them
		bc = block.NewBlockchain(minersWallet.BlockChainAddress(), bcs.Port())
		// when we generate the block chain we will add it to the cache
		cache["blockchain"] = bc
		log.Printf("Private_key %v", minersWallet.PrivateKeyStr())
		log.Printf("Private_key %v", minersWallet.PublicKeyStr())
		log.Printf("Private_key %v", minersWallet.BlockChainAddress())
	}
	return bc
}

func (bcs *BlockchainServer) GetChain(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		w.Header().Add("Content-Type", "application/json")
		bc := bcs.GetBlockchain()
		m, _ := bc.MarshalJSON()
		// 	/* io - use this for writing simple, unformatted string, send plain string response */
		io.WriteString(w, string(m[:]))
	default:
		log.Printf("Error : Invalid HTTP Method")
	}
}

func (bcs *BlockchainServer) Transactions(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		w.Header().Add("Content-type", "application/json")
		bc := bcs.GetBlockchain()
		transaction := bc.TransactionPool()
		m, _ := json.Marshal(struct {
			Transaction []*block.Transaction `json:"transactions"`
			Length      int                  `json:"length"`
		}{
			Transaction: transaction,
			Length:      len(transaction),
		})
        io.WriteString(w,string(m[:]))

	case http.MethodPost:
		var t block.TransactionRequest
		decoder := json.NewDecoder(req.Body)
		err := decoder.Decode(&t)
		if err != nil {
			log.Printf("ERROR: %v", err)
			io.WriteString(w, "fail")
			return
		}
		if !t.Validate() {
			log.Println("ERROR: field(s)")
			io.WriteString(w, "fail")
			return
		}
		publicKey := utils.PublicKeyFromString(*t.SenderPublicKey)
		signature := utils.SignatureFromString(*t.Signature)

		bc := bcs.GetBlockchain()
		isCreated := bc.CreateTransaction(*t.SenderBlockchainAddress,
			*t.RecipientBlockchainAddress, *t.Value, publicKey, signature)

		w.Header().Add("Content-type", "application/json")
		// var m []byte
		if !isCreated {
			w.WriteHeader(http.StatusBadRequest)
			// m := "fail" //////////
		} else {
			w.WriteHeader(http.StatusCreated)
			// m := "Success" //////////////
		}
		io.WriteString(w, "fail/success")

	default:
		log.Println("ERROR: Invalid HTTP Method")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (bcs *BlockchainServer) Mine(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		bc := bcs.GetBlockchain()
		isMined := bc.Mining()

		var m []byte
		if !isMined {
			w.WriteHeader(http.StatusBadRequest)
			m = []byte("fail")
		} else {
			m = []byte("success")
		}
		w.Header().Add("content-Type","application/json")
		io.WriteString(w,string(m))
	default:
		log.Println("ERROR: Invalid HTTP Method")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (bcs *BlockchainServer) StartMine(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		bc := bcs.GetBlockchain()
		bc.StartMining()

		m := []byte("success")
		w.Header().Add("content-Type","application/json")
		io.WriteString(w,string(m))
	default:
		log.Println("ERROR: Invalid HTTP Method")
		w.WriteHeader(http.StatusBadRequest)
	}
}

// how much crypto does the use have calculation 
func (bcs *BlockchainServer) Amount(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		blockchainAddress := req.URL.Query().Get("blockchain_address")
		amount := bcs.GetBlockchain().CalculateTotalAmount(blockchainAddress)

		ar := &block.AmountResponse{Amount: amount}
		m, _ := ar.MarshalJSON()

		w.Header().Add("Content-Type" ,"application/json")
		io.WriteString(w,string(m[:]))
	default:
		log.Printf("ERROR: Invalid HTTP Method")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (bcs *BlockchainServer) Run() {
	http.HandleFunc("/", bcs.GetChain)
    http.HandleFunc("/transactions", bcs.Transactions)
    http.HandleFunc("/mine", bcs.Mine)
    http.HandleFunc("/mine/start", bcs.StartMine)
    http.HandleFunc("/amount", bcs.Amount)
	/* 0.0.0.0 special address that is telling to listen on all available network interface, it means that the sever
	will accept connection from any IP address that the machine has including localhost 127.0.0.1 and any external IPs

	strconv.Itoa - converts the integer port number to its string representation */
	address := "0.0.0.0:" + strconv.Itoa(int(bcs.Port()))
	log.Fatal(http.ListenAndServe(address, nil))
}
