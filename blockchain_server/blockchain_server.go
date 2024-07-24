package main

import (
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/AarizZafar/goblockchain/block"
    "github.com/AarizZafar/goblockchain/wallet"
)

/* we do not want to create a complete block chain again when ever we create a new transaction
   once its made we want to tstore it in out cache 
 */
var cache map[string]*block.Blockchain = make(map[string]*block.Blockchain)

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
    bc, ok := cache["blockchain"]  // checking if we have the blockchain in our cache or not
    if !ok { // at the very begining the cache is empty 
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

func (bcs *BlockchainServer) Run() {
    http.HandleFunc("/", bcs.GetChain)
	/* 0.0.0.0 special address that is telling to listen on all available network interface, it means that the sever
	will accept connection from any IP address that the machine has including localhost 127.0.0.1 and any external IPs
	
	strconv.Itoa - converts the integer port number to its string representation */
    address := "0.0.0.0:" + strconv.Itoa(int(bcs.Port()))
    log.Fatal(http.ListenAndServe(address, nil))
}