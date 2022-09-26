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

func bcyAccounts() gin.H {
	var (
		wallets = []struct {
			Address         string
			Private         string
			Public          string
			Unconfirmed     float64
			Balance         float64
			BalanceSatoshis int64
			Transactions    string
		}{
			{
				Address: "C8psDNmwR88b6fapZhHhjDiT16em9Wdqso",
				Private: "8a0149f1662eeabc063f0f565b27ed15f11281c6bd07805f09ea328e01f95e6b",
				Public:  "03a2590f54fb8e55cee0053d4c9f48efd6eb3cc6b6277901409e5cba809cade6ce",
			},
			{
				Address: "C99p1B5FdAsZbjPcPfQ3vpdidUgJtPBjKy",
				Private: "fb54ca81541111d62d59d9cb48b583da4fe92a939f8a9bf40cf3f64e6aecb25b",
				Public:  "03a35c9f16e94b1efdd0f5eabd6610741c6688fc7b143586224c4a7b9625e2eaf0",
			},
			{
				Address: "CBcZZYsVbo1APihkkW8NWVfAwW4gSRxGv7",
				Private: "f61f4bbd3aa6028b83fbbe99b0d6651a82184880ae5d964878eca5ebff3609db",
				Public:  "02475adc7c7170a5a145a7d16725007595e6221f2e126feeb7cde92dded9dff53a",
			},
		}
	)

	for _, wallet := range wallets {
		wallet.Transactions = fmt.Sprintf("%s/%s/", "https://live.blockcypher.com/bcy/address", wallet.Address)
		addr, err := BcApi.GetAddr(wallet.Address, nil)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return gin.H{}
		}
		wallet.Balance = float64(addr.Balance.Int64()) / 100000000
		wallet.BalanceSatoshis = addr.Balance.Int64()
		wallet.Unconfirmed = float64(addr.UnconfirmedBalance.Int64()) / 100000000
	}

	return gin.H{
		"wallets": wallets,
	}
}
