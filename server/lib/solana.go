package lib

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type Meta struct {
	PostBalances []int64 `json:"postBalances"`
	PreBalances  []int64 `json:"preBalances"`
}

type Message struct {
	AccountKeys []AccountKey `json:"accountKeys"`
}

type Transaction struct {
	Message Message `json:"message"`
}

type Result struct {
	Transaction Transaction `json:"transaction"`
	Meta        Meta        `json:"meta"`
}

type RPCResponse struct {
	Result Result `json:"result"`
}

type SimpleTransaction struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount int64  `json:"amount"`
}

type AccountKey struct {
	Pubkey string `json:"pubkey"`
	Signer bool   `json:"signer"`
}

func GetTransaction(signature string) (*SimpleTransaction, error) {
	// url := fmt.Sprintf("https://solana-devnet.g.alchemy.com/v2/%s", os.Getenv("ALCHEMY_API_KEY"))
	url := fmt.Sprintf("https://devnet.helius-rpc.com/?api-key=%s", os.Getenv("HELIUS_API_KEY"))

	requestBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getTransaction",
		"params":  []interface{}{signature, "jsonParsed"},
	}
	body, _ := json.Marshal(requestBody)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var rpcResponse RPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResponse); err != nil {
		return nil, err
	}
	log.Println(rpcResponse)

	from := rpcResponse.Result.Transaction.Message.AccountKeys[0].Pubkey
	to := rpcResponse.Result.Transaction.Message.AccountKeys[1].Pubkey
	amount := rpcResponse.Result.Meta.PreBalances[0] - rpcResponse.Result.Meta.PostBalances[0]
	return &SimpleTransaction{
		From:   from,
		To:     to,
		Amount: amount,
	}, nil
}

func VerifySignature(publicKeyHex string, message []byte, signatureHex string) bool {
	publicKey, err := hex.DecodeString(publicKeyHex)
	if err != nil {
		return false
	}

	signature, err := hex.DecodeString(signatureHex)
	if err != nil {
		return false
	}

	return ed25519.Verify(publicKey, message, signature)
}

// func main() {
//     // Example signature and public key
//     transactionSignature := "YOUR_TRANSACTION_SIGNATURE"
//     publicKeyHex := "YOUR_PUBLIC_KEY"

//     // Fetch transaction details
//     transaction, err := getTransaction(transactionSignature)
//     if err != nil {
//         fmt.Println("Error fetching transaction:", err)
//         return
//     }

//     // Verify the signature
//     message := []byte("...") // You need to create the message from transaction details
//     if verifySignature(publicKeyHex, message, transaction.Transaction.Signatures[0]) {
//         fmt.Println("Signature is valid!")
//     } else {
//         fmt.Println("Invalid signature.")
//     }

//     // Print transaction details
//     fmt.Printf("Transaction Slot: %d\n", transaction.Slot)
//     fmt.Printf("Transaction Details: %+v\n", transaction.Transaction)
// }
