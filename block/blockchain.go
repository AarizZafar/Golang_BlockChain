package block

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/AarizZafar/goblockchain/utils"
)

// 3 is the difficulty level that means we the nonce has to start with 3 zeroes
const (
	MINING_DIFFICULTY = 3
	MINING_SENDER     = "THE BLOCKCHAIN"
	MINING_REWARD     = 1.0
)

type Block struct {
	timestamp    int64
	nonce        int
	previousHash [32]byte
	transactions []*Transaction
}

func NewBlock(nonce int, previousHash [32]byte, transactions []*Transaction) *Block {
	/* Allocates memory for a new 'Block' struct and initializes its field to their zero value
	   ('0' for numeric type ' "" ' for string 'nil' for slices) */
	b := new(Block)
	b.timestamp = time.Now().UnixNano()
	b.nonce = nonce
	b.previousHash = previousHash
	b.transactions = transactions
	return b
}

func (b *Block) Print() {
	fmt.Printf("timestamp        %d\n", b.timestamp)
	fmt.Printf("nonce            %d\n", b.nonce)
	fmt.Printf("previous_hash    %x\n", b.previousHash)

	for _, t := range b.transactions {
		t.Print()
	}
}

func (b *Block) Hash() [32]byte {
	m, _ := json.Marshal(b) // Converting the Block instance 'b' into a JSON-encoded byte

	// fmt.Println(string(m))              // An empty object printed out it is due to the way Martial function works
	/* *****************************IMPORTANT******************************
	    in the below code when we had done 	block := &Block{nonce: 1}
		we only gave a int value to nonce and the rest of the value are set to nil.
		The reason why we get a null are 2 reason ->
		1) fields with zero value like ("",0,nil) for slices are typically omitted from the JSON output to reduce redundancy (including repetative or unnecessary information)
		2) But we had give a int value to nonce but we still get a {} the reason beeing is Field Visability
		   json.Marshal function only includes fields that are exposed i.e fields that start with an uppercase by default
		   all our fields are in lower case  nonce, previousHash, timestamp, transaction
	*/

	return sha256.Sum256([]byte(m)) // it takes a byte slice as input and returns a fixed-size array([32]byte) cintaining the SHA-256 hash
}

func (b *Block) MarshalJSON() ([]byte, error) {
	/*
	   over here we are ensuring that all the fields include those with zero value (before it was all nil),
	   we defined an anonymous struct with the same fields as the Block, but using the exported fields
	*/
	return json.Marshal(struct {
		Timestamp    int64          `json:"timestamp"`
		Nonce        int            `json:"nonce"`
		PreviousHash string         `json:"previous_hash"`
		Transactions []*Transaction `json:"transactions"`
	}{
		Timestamp:    b.timestamp,
		Nonce:        b.nonce,
		PreviousHash: fmt.Sprintf("%x", b.previousHash),
		Transactions: b.transactions,
	})
}

type Blockchain struct {
	transactionPool   []*Transaction // Holds pending transaction to be added to block
	chain             []*Block       // holds the blockchain as a list of Block pointers
	blockchainAddress string
	port              uint16
}

func NewBlockchain(blockchainAddress string, port uint16) *Blockchain {
	b := &Block{} // storing the block in temp when creating a new block we are possing the hash of this b block
	bc := new(Blockchain)
	bc.blockchainAddress = blockchainAddress
	/*
		in the initial stage we do not have any previous block thats why store 0 in nonce and
		we dont have a previous hash so we store Init hash
	*/
	bc.CreateBlock(0, b.Hash())
	bc.port = port
	return bc
}

func (bc *Blockchain) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Blocks []*Block `json:"chains"`
	} {
		Blocks : bc.chain,
	})
}

func (bc *Blockchain) CreateBlock(nonce int, previousHash [32]byte) *Block {
	b := NewBlock(nonce, previousHash, bc.transactionPool) // creates a new block using a helper function NewBlock
	bc.chain = append(bc.chain, b)                         // appends the new Block to the blockchain (chain)
	bc.transactionPool = []*Transaction{}                  // after the transaction is added to the block we are emptying the transaction pool
	return b                                               // returns a created block
}

// Creating a function to identify which block is the last block
func (bc *Blockchain) LastBlock() *Block {
	return bc.chain[len(bc.chain)-1]
}

func (bc *Blockchain) Print() {
	for i, block := range bc.chain {
		fmt.Printf("%s Chain %d %s\n", strings.Repeat("=", 25), i, strings.Repeat("=", 25))
		block.Print()
	}
	fmt.Printf("%s\n", strings.Repeat("*", 25))
}

// -----------------------------------------------------------------------------------------------
func (bc *Blockchain) AddTransaction(sender string, recipient string, value float32, senderPublicKey *ecdsa.PublicKey, s *utils.Signature) bool {
	t := NewTransaction(sender, recipient, value)

	if sender == MINING_SENDER {
		bc.transactionPool = append(bc.transactionPool, t)
		return true
	}

	if bc.VerifyTransactionSignature(senderPublicKey, s, t) {
		/*
			if bc.CalculateTotalAmount(sender) < value {
				log.Println("Error : Not enough balance in a wallet to send ")
				return false
			}
		*/
		bc.transactionPool = append(bc.transactionPool, t)
		return true
	} else {
		log.Println("Error : Verify Transaction")
	}
	return false
}

func (bd *Blockchain) VerifyTransactionSignature(senderPublicKey *ecdsa.PublicKey, s *utils.Signature, t *Transaction) bool {
	m, _ := json.Marshal(t)                              // The transaction is converted to bytes
	h := sha256.Sum256([]byte(m))                        // encoding it
	return ecdsa.Verify(senderPublicKey, h[:], s.R, s.S) // using the senders public key and verifying the transaction was it done by the sender or not
}

func (bc *Blockchain) CopyTransactionPool() []*Transaction {
	transactions := make([]*Transaction, 0)
	for _, t := range bc.transactionPool {
		transactions = append(transactions,
			NewTransaction(t.senderBlockchainAddress,
				t.recipientBlockchainAddress,
				t.value))
	}
	return transactions
}

// a hash that starts with 3 "0" will only be created and added to the block chain
func (bc *Blockchain) ValidProof(nonce int, previousHash [32]byte, transactions []*Transaction, difficulty int) bool {
	zeros := strings.Repeat("0", difficulty)
	guessBlock := Block{0, nonce, previousHash, transactions}
	guessHashStr := fmt.Sprintf("%x", guessBlock.Hash())
	return guessHashStr[:difficulty] == zeros // determining of the leading 3 values are zeroes or not
}

// see notion to understand better 
func (bc *Blockchain) ProofOfWork() int {
	transaction := bc.CopyTransactionPool()
	previousHash := bc.LastBlock().Hash()

	// the formula that is beeing used to calculate the nonce is (nonce + prev Hash + transaction)
	// the nonce will keep incrementill we get an proff that has 3 zeroes in the starting of it
	nonce := 0
	for !bc.ValidProof(nonce, previousHash, transaction, MINING_DIFFICULTY) {
		nonce += 1
	}
	return nonce
}

// creating a block and adding it to the chain 
func (bc *Blockchain) Mining() bool {
	bc.AddTransaction(MINING_SENDER, bc.blockchainAddress, MINING_REWARD, nil, nil)
	nonce := bc.ProofOfWork()
	previousHash := bc.LastBlock().Hash()
	bc.CreateBlock(nonce, previousHash)
	log.Println("action=mining status=success")
	return true
}

// checking how much coins does the send and the receiver have in total now 
func (bc *Blockchain) CalculateTotalAmount(blockchainAddress string) float32 {
	var totalAmount float32 = 0.0
	for _, b := range bc.chain {
		for _, t := range b.transactions {
			value := t.value
			if blockchainAddress == t.recipientBlockchainAddress {
				totalAmount += value
			}
			if blockchainAddress == t.senderBlockchainAddress {
				totalAmount -= value
			}
		}
	}
	return totalAmount
}

type Transaction struct {
	senderBlockchainAddress    string
	recipientBlockchainAddress string
	value                      float32
}

func NewTransaction(sender string, recipient string, value float32) *Transaction {
	return &Transaction{sender, recipient, value}
}

func (t *Transaction) Print() {
	fmt.Printf("%s\n", strings.Repeat("-", 40))
	fmt.Printf(" sender_blockchain_address       %s\n", t.senderBlockchainAddress)
	fmt.Printf(" recipient_blockchain_addresss   %s\n", t.recipientBlockchainAddress)
	fmt.Printf("value                            %.1f\n", t.value)
}

func (t *Transaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Sender    string  `json:"sender_blockchain_address"`
		Recipient string  `json:"recipient_blockchain_address"`
		Value     float32 `json:"value"`
	}{
		Sender:    t.senderBlockchainAddress,
		Recipient: t.recipientBlockchainAddress,
		Value:     t.value,
	})
}
