package main

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/big"
	"net/http"
	"regexp"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
)

var (
	// 0xa9c066859E4B1a227143DaA3dbdd3b3Ce0ae14b5
	EthPrivateBank = []byte{206, 30, 216, 11, 94, 171, 92, 14, 93, 35, 32, 129, 153, 0, 138, 130, 51, 197, 138, 192, 53, 98, 169, 124, 121, 128, 153, 23, 129, 223, 18, 97}
	// 0xE48a7F0d63D00b5c209CB663bac0ec3e1410f7b7s
	EthPrivateFaucet = []byte{98, 108, 5, 164, 87, 95, 210, 66, 49, 85, 57, 155, 193, 72, 84, 141, 143, 116, 113, 44, 10, 12, 17, 202, 125, 145, 171, 200, 3, 74, 210, 145}
)

type EthWallet struct {
	Address  string
	Private  []byte
	PrivateS string
	Public   []byte
	PublicS  string
}

func generateEthWallet() EthWallet {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		panic(err)
	}
	privateKeyBytes := crypto.FromECDSA(privateKey)

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		panic("error casting public key to ECDSA")
	}

	return EthWallet{
		Address:  crypto.PubkeyToAddress(*publicKeyECDSA).Hex(),
		Private:  crypto.FromECDSA(privateKey),
		PrivateS: hexutil.Encode(privateKeyBytes)[2:],
		Public:   crypto.FromECDSAPub(publicKeyECDSA),
		PublicS:  hexutil.Encode(crypto.FromECDSAPub(publicKeyECDSA))[4:],
	}
}

type EthBalance struct {
	Balance    int64
	BalanceEth float64
	Pending    int64
}

func getEthBalance(addr string) EthBalance {
	account := common.HexToAddress(addr)
	balance, err := AlchemyClient.BalanceAt(context.Background(), account, nil)
	if err != nil {
		panic(err)
	}

	fbalance := new(big.Float)
	fbalance.SetString(balance.String())
	ethValue, _ := new(big.Float).Quo(fbalance, big.NewFloat(math.Pow10(18))).Float64()
	pendingBalance, _ := AlchemyClient.PendingBalanceAt(context.Background(), account)

	return EthBalance{
		Balance:    balance.Int64(),
		BalanceEth: ethValue,
		Pending:    pendingBalance.Int64(),
	}
}

func getEthWallet(key []byte) EthWallet {
	privateKey, err := crypto.ToECDSA(key)
	if err != nil {
		panic(err)
	}

	privateKeyBytes := crypto.FromECDSA(privateKey)

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		panic("error casting public key to ECDSA")
	}

	return EthWallet{
		Address:  crypto.PubkeyToAddress(*publicKeyECDSA).Hex(),
		Private:  crypto.FromECDSA(privateKey),
		PrivateS: hexutil.Encode(privateKeyBytes)[2:],
		Public:   crypto.FromECDSAPub(publicKeyECDSA),
		PublicS:  hexutil.Encode(crypto.FromECDSAPub(publicKeyECDSA))[4:],
	}
}

type AlchemyGetAssetTransfersRequest struct {
	ID      int                    `json:"id"`
	JsonRPC string                 `json:"jsonrpc"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params"`
}

type AlchemyGetAssetTransfersResponse struct {
	ID      int    `json:"id"`
	JsonRPC string `json:"jsonrpc"`
	Result  struct {
		Transfers []AssetTransfer `json:"transfers"`
	} `json:"result"`
}

type AssetTransfer struct {
	BlockNum string      `json:"blockNum"`
	From     string      `json:"from"`
	To       string      `json:"to"`
	Value    interface{} `json:"value,omitempty"`
	Asset    string      `json:"asset"`
	Hash     string      `json:"hash"`
}

func getEthTransactions(addr string, in bool) []AssetTransfer {
	url := fmt.Sprintf("%s/%s", AlchemyAddr, AlchemyKey)

	payload := AlchemyGetAssetTransfersRequest{
		ID:      1,
		JsonRPC: "2.0",
		Method:  "alchemy_getAssetTransfers",
		Params: map[string]interface{}{
			"category": []string{"external"},
		},
	}
	if in {
		payload.Params["toAddress"] = addr
	} else {
		payload.Params["fromAddress"] = addr
	}

	jsPayload, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsPayload))

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	var answer AlchemyGetAssetTransfersResponse
	if err := json.Unmarshal(body, &answer); err != nil {
		panic(err)
	}

	return answer.Result.Transfers
}

func getEthBankAccount() gin.H {
	ethWallet := getEthWallet(EthPrivateBank)

	return gin.H{
		"wallet":       ethWallet,
		"balance":      getEthBalance(ethWallet.Address),
		"transactions": getEthTransactions(ethWallet.Address, true),
	}
}

func getEthFaucetAccount() gin.H {
	ethWallet := getEthWallet(EthPrivateFaucet)

	return gin.H{
		"wallet":       ethWallet,
		"balance":      getEthBalance(ethWallet.Address),
		"transactions": getEthTransactions(ethWallet.Address, false),
	}
}

type EthFaucetRequestModel struct {
	Addreess string  `json:"address"`
	Amount   float64 `json:"amount"`
}

func ethFaucetHandler(c *gin.Context) {
	var model EthFaucetRequestModel
	if err := c.BindJSON(&model); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	re := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")
	if !re.MatchString(model.Addreess) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "address not valid",
		})
		return
	}
	toAddress := common.HexToAddress(model.Addreess)

	faucetWallet := getEthWallet(EthPrivateFaucet)

	privateKey, err := crypto.ToECDSA(faucetWallet.Private)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	nonce, err := AlchemyClient.PendingNonceAt(context.Background(), common.HexToAddress(faucetWallet.Address))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	amount := model.Amount * 1000000000000000000
	value := big.NewInt(int64(amount)) // in wei (1 eth)
	gasLimit := uint64(21000)          // in units
	gasPrice, err := AlchemyClient.SuggestGasPrice(context.Background())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Make transaction
	var data []byte
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)

	// Chain ID
	chainID, err := AlchemyClient.NetworkID(context.Background())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Sign
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Send
	if err := AlchemyClient.SendTransaction(context.Background(), signedTx); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.Status(http.StatusOK)
}
