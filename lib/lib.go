package lib

import (
	"fmt"

	"github.com/tensor-programming/golang-blockchain/blockchain"
	"github.com/tensor-programming/golang-blockchain/network"
	"github.com/tensor-programming/golang-blockchain/wallet"
)

func CreateWallet(nodeID string) string {
	wallets, _ := wallet.CreateWallets(nodeID)
	address := wallets.AddWallet()
	wallets.SaveFile(nodeID)
	return address
}

func ListAddresses(nodeID string) []string {
	wallets, _ := wallet.CreateWallets(nodeID)
	addresses := wallets.GetAllAddresses()
	return addresses
}

func GetBalance(nodeID string, address string) int {
	chain := blockchain.ContinueBlockChain(nodeID)
	UTXOSet := blockchain.UTXOSet{
		Blockchain: chain,
	}
	defer chain.Database.Close()

	balance := 0
	pubKeyHash := wallet.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := UTXOSet.FindUnspentTransactions(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}
	return balance
}

func Send(nodeID string, from string, to string, amount int, mineNow bool) (*blockchain.Transaction, error) {
	chain := blockchain.ContinueBlockChain(nodeID)
	UTXOSet := blockchain.UTXOSet{
		Blockchain: chain,
	}
	defer chain.Database.Close()

	wallets, err := wallet.CreateWallets(nodeID)
	if err != nil {
		return nil, err
	}
	wallet := wallets.GetWallet(from)

	tx := blockchain.NewTransaction(&wallet, to, amount, &UTXOSet)
	if mineNow {
		cbTx := blockchain.CoinbaseTx(from, "")
		txs := []*blockchain.Transaction{cbTx, tx}
		block := chain.MineBlock(txs)
		UTXOSet.Update(block)
	} else {
		network.SendTx(network.KnownNodes[0], tx)
		fmt.Println("send tx")
	}

	fmt.Println("Success!")
	return tx, nil
}

func ReindexUTXO(nodeID string) int {
	chain := blockchain.ContinueBlockChain(nodeID)
	defer chain.Database.Close()
	UTXOSet := blockchain.UTXOSet{
		Blockchain: chain,
	}
	UTXOSet.Reindex()

	count := UTXOSet.CountTransactions()
	return count
}
