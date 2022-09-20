package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/blockcypher/gobcy/v2"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var (
	BcToken  = "blockcypher-token"
	HttpAddr = "localhost:9091"

	BcApi *gobcy.API
)

func init() {
	if bcToken := os.Getenv("BC_TOKEN"); bcToken != "" {
		BcToken = bcToken
	}
	if httpAddr := os.Getenv("HTTP_ADDR"); httpAddr != "" {
		HttpAddr = httpAddr
	}

	BcApi = &gobcy.API{
		Token: BcToken,
		Coin:  "bcy",
		Chain: "test",
	}
}

func main() {
	router := gin.Default()
	router.Use(cors.Default())

	router.POST("/bitcoin/addresses/balances", getBalancesHandler)

	router.Run(HttpAddr)
}

type BalancesAnswerModel struct {
	Data    map[string]int64       `json:"data"`
	Context map[string]interface{} `json:"context"`
}

func getBalancesHandler(c *gin.Context) {
	var addresses []string
	postData, err := c.GetRawData()
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	if addressesString := strings.Split(string(postData), "="); len(addressesString) == 2 {
		addressesStringArr := strings.Split(addressesString[1], ",")
		addresses = append(addresses, addressesStringArr...)
	} else {
		c.AbortWithError(http.StatusBadRequest, errors.New("wrong data"))
		return
	}

	var (
		answer = BalancesAnswerModel{
			Data: make(map[string]int64),
			Context: map[string]interface{}{
				"code": 200,
			},
		}
	)
	for _, addr := range addresses {
		if balance := getBalance(addr); balance > 0 {
			answer.Data[addr] = balance
		}
	}
	c.JSON(http.StatusOK, answer)
}

func getBalance(hash string) int64 {
	balance, err := BcApi.GetAddrBal(hash, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return 0
	}
	return balance.FinalBalance.Int64()
}

// "C9LBdupQfLTtgsKDNRdeo6AroDMAeqoEqD"
