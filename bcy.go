package main

import (
	"fmt"
	"math/big"
	"net/http"

	"github.com/blockcypher/gobcy/v2"
	"github.com/gin-gonic/gin"
)

type BcyFaucetRequestModel struct {
	Addreess string  `json:"address"`
	Amount   float64 `json:"amount"`
}

func bcyFaucetHandler(c *gin.Context) {
	var model BcyFaucetRequestModel
	if err := c.BindJSON(&model); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	addr, err := BcApi.GetAddr(model.Addreess, nil)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Genetate random wallet
	genAddr, err := BcApi.GenAddrKeychain()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Faucet
	var btcAmount = int(model.Amount * 100000000)
	_, err = BcApi.Faucet(genAddr, btcAmount+50000)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// TX Sceleton
	txSkeleton, err := BcApi.NewTX(gobcy.TempNewTX(genAddr.Address, addr.Address, *big.NewInt(int64(btcAmount))), false)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Sign TX
	if err := txSkeleton.Sign([]string{genAddr.Private}); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Send TX
	if _, err := BcApi.SendTX(txSkeleton); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"from":   genAddr.Address,
		"to":     addr.Address,
		"amount": model.Amount,
	})
}

func bcyBankWalletData() gin.H {
	const (
		address    = "C16ZKtLMjo1e3KxxEaD6uewtdkYeojFQo1"
		privateKey = "c617d15889ba6b350da66a7a371beec20a99460760382cb9a0ac382382e56fd8"
		publicKey  = "02e07efec6e927ea007bd870969e93fef00c10780da5b6d75e2a0b9fbd2838736f"
	)

	addr, err := BcApi.GetAddr(address, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return gin.H{}
	}

	type Transaction struct {
		Addresses []string
		Status    bool
	}
	var (
		transactions = make([]Transaction, 0)
	)
	for _, txRef := range addr.TXRefs {
		tx, err := BcApi.GetTX(txRef.TXHash, nil)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return gin.H{}
		}
		transactions = append(transactions, Transaction{
			Addresses: tx.Addresses,
			Status:    tx.BlockHeight > 0,
		})
	}

	return gin.H{
		"wallet": map[string]interface{}{
			"address":         address,
			"private":         privateKey,
			"public":          publicKey,
			"unconfirmed":     float64(addr.UnconfirmedBalance.Int64()) / 100000000,
			"balance":         float64(addr.Balance.Int64()) / 100000000,
			"balanceSatoshis": addr.Balance.Int64(),
		},
		"transactions": transactions,
	}
}
