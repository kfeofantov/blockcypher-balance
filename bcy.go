package main

import (
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
