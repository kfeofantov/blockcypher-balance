package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/blockcypher/gobcy/v2"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var (
	BcToken     = "blockcypher-token"
	AlchemyAddr = "https://eth-goerli.g.alchemy.com/v2"
	AlchemyKey  = "alchemy-token"
	HttpAddr    = "localhost:9091"

	BcApi         *gobcy.API
	AlchemyClient *ethclient.Client
)

func init() {
	if bcToken := os.Getenv("BC_TOKEN"); bcToken != "" {
		BcToken = bcToken
	}
	if alchemyKey := os.Getenv("ALCHEMY_TOKEN"); alchemyKey != "" {
		AlchemyKey = alchemyKey
	}
	if httpAddr := os.Getenv("HTTP_ADDR"); httpAddr != "" {
		HttpAddr = httpAddr
	}

	BcApi = &gobcy.API{
		Token: BcToken,
		Coin:  "bcy",
		Chain: "test",
	}

	if client, err := ethclient.Dial(fmt.Sprintf("%s/%s", AlchemyAddr, AlchemyKey)); err != nil {
		panic(err)
	} else {
		AlchemyClient = client
	}
}

func main() {
	router := gin.Default()
	router.Use(cors.Default())

	router.POST("/bitcoin/addresses/balances", getBalancesHandler)
	router.POST("/bcy-faucet", bcyFaucetHandler)
	router.POST("/eth-faucet", ethFaucetHandler)

	// Load and render HTML pages
	router.LoadHTMLGlob("templates/*.html")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	router.GET("/bcy-faucet", func(c *gin.Context) {
		c.HTML(http.StatusOK, "bcy-faucet.html", nil)
	})
	router.GET("/bcy-bank", func(c *gin.Context) {
		c.HTML(http.StatusOK, "bcy-bank.html", bcyBankWalletData())
	})
	router.GET("/bcy-accounts", func(c *gin.Context) {
		c.HTML(http.StatusOK, "bcy-accounts.html", bcyAccounts())
	})

	router.GET("/eth-generate", func(c *gin.Context) {
		c.HTML(http.StatusOK, "eth-generate.html", generateEthWallet())
	})
	router.GET("/eth-bank", func(c *gin.Context) {
		c.HTML(http.StatusOK, "eth-bank.html", getEthBankAccount())
	})
	router.GET("/eth-faucet", func(c *gin.Context) {
		c.HTML(http.StatusOK, "eth-faucet.html", getEthFaucetAccount())
	})

	// Serve http server
	router.Run(HttpAddr)
}

type BalancesAnswerModel struct {
	Data    map[string]int64       `json:"data"`
	Context map[string]interface{} `json:"context"`
}

func getBalancesHandler(c *gin.Context) {
	var (
		addressesString string
		addresses       []string
	)

	switch c.Request.Header.Get("Content-Type") {
	case "application/json":
		var addressesData = struct {
			Addresses string `json:"addresses"`
		}{}
		if err := c.BindJSON(&addressesData); err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		addressesString = addressesData.Addresses
	default:
		postData, err := c.GetRawData()
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		if addressesStringArr := strings.Split(string(postData), "="); len(addressesStringArr) == 2 {
			addressesString = addressesStringArr[1]
		} else {
			c.AbortWithError(http.StatusBadRequest, errors.New("wrong data"))
			return
		}
	}

	// Split addresses
	addressesStringArr := strings.Split(addressesString, ",")
	addresses = append(addresses, addressesStringArr...)

	// Make answer
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
	balance, err := BcApi.GetAddrBal(strings.TrimSpace(hash), nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return 0
	}
	return balance.FinalBalance.Int64()
}
