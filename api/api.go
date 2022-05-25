package api

import (
	"errors"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/tensor-programming/golang-blockchain/blockchain"
	"github.com/tensor-programming/golang-blockchain/lib"
	"github.com/tensor-programming/golang-blockchain/wallet"
)

type Api struct{}

type CreateWallet struct {
	Name string `json:"name" binding:"required"`
}

func ErrorHandler(c *gin.Context) {
	c.Next()

	if len(c.Errors) > 0 {
		var errors []string
		// ignore EOF errors
		for _, e := range c.Errors {
			if e.Err == http.ErrBodyReadAfterClose {
				continue
			}
			errors = append(errors, e.Error())
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors,
		})
	}
}

func StartServer() {
	r := gin.Default()
	r.Use(ErrorHandler)

	r.POST("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.GET("/createwallet", func(c *gin.Context) {
		address := lib.CreateWallet(os.Getenv("NODE_ID"))
		c.JSON(http.StatusOK, gin.H{
			"address": address,
		})
	})

	r.GET("/listaddresses", func(c *gin.Context) {
		addresses := lib.ListAddresses(os.Getenv("NODE_ID"))
		c.JSON(http.StatusOK, gin.H{
			"addresses": addresses,
		})
	})

	r.GET("/printchain", func(c *gin.Context) {
		chain := blockchain.ContinueBlockChain(os.Getenv("NODE_ID"))
		defer chain.Database.Close()
		iter := chain.Iterator()

		var blocks []blockchain.Block

		for {
			block := iter.Next()
			blocks = append(blocks, *block)
			if len(block.PrevHash) == 0 {
				break
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"blocks": blocks,
		})
	})

	r.GET("/createblockchain", func(c *gin.Context) {
		address := c.Query("address")
		validateAddress(address, c)
		chain, err := blockchain.InitBlockChain(address, os.Getenv("NODE_ID"))

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Blockchain already exists",
			})
			return
		}

		UTXOSet := blockchain.UTXOSet{
			Blockchain: chain,
		}
		UTXOSet.Reindex()

		c.JSON(http.StatusOK, gin.H{
			"message": "Finished!",
		})
	})

	r.GET("/getbalance", func(c *gin.Context) {
		address := c.Query("address")
		validateAddress(address, c)
		balance := lib.GetBalance(os.Getenv("NODE_ID"), address)

		c.JSON(http.StatusOK, gin.H{
			"balance": balance,
		})
	})

	r.POST("/send", func(c *gin.Context) {
		type Send struct {
			From   string `json:"from" binding:"required"`
			To     string `json:"to" binding:"required"`
			Amount int    `json:"amount" binding:"required"`
			Mine   bool   `json:"mine"`
		}
		var send Send
		err := c.ShouldBindJSON(&send)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		if err := validator.New().Struct(send); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		from := send.From
		to := send.To
		amount := send.Amount
		mine := send.Mine

		validateAddress(from, c)
		validateAddress(to, c)

		tx, err := lib.Send(os.Getenv("NODE_ID"), from, to, amount, mine)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":     "Finished!",
			"transaction": tx,
		})
	})

	r.GET("/reindexutxo", func(c *gin.Context) {
		count := lib.ReindexUTXO(os.Getenv("NODE_ID"))
		c.JSON(http.StatusOK, gin.H{
			"message": "Finished!",
			"txCount": count,
		})
	})

	r.Run("localhost:8080")
}

func validateAddress(address string, c *gin.Context) {
	if len(address) < 5 {
		c.AbortWithError(
			http.StatusBadRequest,
			errors.New("address is not valid"),
		)
		return
	}
	if !wallet.ValidateAddress(address) {
		c.AbortWithError(
			http.StatusBadRequest,
			errors.New("address is not valid"),
		)
		return
	}
}
