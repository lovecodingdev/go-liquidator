package jupag

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"go-liquidator/utils"
	"io/ioutil"
	"net/http"

	"github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/common"
	"github.com/portto/solana-go-sdk/rpc"
	"github.com/portto/solana-go-sdk/types"
)

type CoinQuote struct {
	Data      []Route `json:"data"`
	TimeTaken string  `json:"timeTaken"`
}

type Route struct {
	InAmount              uint64       `json:"inAmount"`
	OutAmount             uint64       `json:"outAmount"`
	OutAmountWithSlippage uint64       `json:"outAmountWithSlippage"`
	PriceImpactPct        float64      `json:"priceImpactPct"`
	MarketInfos           []MarketInfo `json:"marketInfos"`
}

type MarketInfo struct {
	Id                 string  `json:"id"`
	Label              string  `json:"label"`
	InputMint          string  `json:"inputMint"`
	OutputMint         string  `json:"outputMint"`
	NotEnoughLiquidity bool    `json:"notEnoughLiquidity"`
	InAmount           uint64  `json:"inAmount"`
	OutAmount          uint64  `json:"outAmount"`
	PriceImpactPct     float64 `json:"priceImpactPct"`
	LpFee              Fee     `json:"lpFee"`
	PlatformFee        Fee     `json:"platformFee"`
}

type Fee struct {
	Amount uint64  `json:"amount"`
	Mint   string  `json:"mint"`
	Pct    float64 `json:"pct"`
}

type SwapTransaction struct {
	SetupTransaction   string `json:"setupTransaction"`
	SwapTransaction    string `json:"swapTransaction"`
	CleanupTransaction string `json:"cleanupTransaction"`
}

type SwapTransactionReq struct {
	Route         Route  `json:"route"`
	WrapUnwrapSOL bool   `json:"wrapUnwrapSOL"`
	FeeAccount    string `json:"feeAccount"`
	TokenLedger   string `json:"tokenLedger"`
	UserPublicKey string `json:"userPublicKey"`
}

func GetCoinQuote(inputMint string, outputMint string, amount uint64) (CoinQuote, error) {
	coinQuoteURL := fmt.Sprintf(
		"https://quote-api.jup.ag/v1/quote?inputMint=%s&outputMint=%s&amount=%d&slippage=0.2",
		inputMint, outputMint, amount,
	)
	// fmt.Println(coinQuoteURL)

	response, err := http.Get(coinQuoteURL)
	if err != nil {
		return CoinQuote{}, err
	}
	defer response.Body.Close()

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return CoinQuote{}, err
	}

	var responseObject CoinQuote
	json.Unmarshal(responseData, &responseObject)

	return responseObject, nil
}

func GetSwapTransaction(route Route, userPublicKey string) (SwapTransaction, error) {
	req := SwapTransactionReq{
		Route:         route,
		UserPublicKey: userPublicKey,
		WrapUnwrapSOL: false,
	}
	body, _ := json.Marshal(req)

	res, err := http.Post("https://quote-api.jup.ag/v1/swap", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return SwapTransaction{}, err
	}
	defer res.Body.Close()

	responseData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return SwapTransaction{}, err
	}

	var responseObject SwapTransaction
	err = json.Unmarshal(responseData, &responseObject)
	if err != nil {
		return SwapTransaction{}, err
	}

	return responseObject, nil
}

func Swap(
	inputMint string,
	outputMint string,
	amount uint64,
	wallet types.Account,
	c *client.Client,
) error {
	coinQuote, err := GetCoinQuote(inputMint, outputMint, amount)
	// fmt.Println(utils.JsonFromObject(coinQuote))
	if err != nil {
		return err
	}

	transactions, err := GetSwapTransaction(coinQuote.Data[0], wallet.PublicKey.ToBase58())
	// fmt.Println(utils.JsonFromObject(transactions))
	if err != nil {
		return err
	}

	txs := []string{transactions.SetupTransaction, transactions.SwapTransaction, transactions.CleanupTransaction}

	for _, rawTx := range txs {
		if rawTx == "" {
			continue
		}
		bytesTx, _ := base64.StdEncoding.DecodeString(rawTx)
		tx, _ := types.TransactionDeserialize(bytesTx)
		// fmt.Println(utils.JsonFromObject(tx))

		signedTx, _ := types.NewTransaction(types.NewTransactionParam{
			Signers: []types.Account{wallet},
			Message: tx.Message,
		})
		// fmt.Println(utils.JsonFromObject(signedTx))

		sig, err := c.SendTransactionWithConfig(context.TODO(), signedTx, client.SendTransactionConfig{
			SkipPreflight:       true,
			PreflightCommitment: rpc.CommitmentConfirmed,
		})
		if err != nil {
			return err
		}
		// fmt.Println(sig, err)

		err = utils.ConfirmTransaction(sig, c)
		if err != nil {
			return err
		}
	}
	fmt.Println("swaped")
	return nil
}

func SwapAllSolTo(
	outputMint string,
	wallet types.Account,
	c *client.Client,
) error {
	balance, err := c.GetBalanceWithConfig(context.TODO(), wallet.PublicKey.ToBase58(), rpc.GetBalanceConfig{
		Commitment: rpc.CommitmentConfirmed,
	})
	if err != nil {
		return err
	}
	if balance > 200_000_000 {
		err := Swap(
			"So11111111111111111111111111111111111111112",
			outputMint,
			balance-200_000_000,
			wallet,
			c,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func SwapSolFrom(
	inputMint string,
	wallet types.Account,
	c *client.Client,
) error {
	userTokenAccount, _, _ := common.FindAssociatedTokenAddress(wallet.PublicKey, common.PublicKeyFromString(inputMint))
	balance, _, err := c.GetTokenAccountBalanceWithConfig(context.TODO(), userTokenAccount.ToBase58(), rpc.GetTokenAccountBalanceConfig{
		Commitment: rpc.CommitmentConfirmed,
	})
	if err != nil {
		return err
	}
	err = Swap(
		inputMint,
		"So11111111111111111111111111111111111111112",
		balance,
		wallet,
		c,
	)
	if err != nil {
		return err
	}
	return nil
}

func SwapMax(
	inputMint string,
	outputMint string,
	wallet types.Account,
	c *client.Client,
) error {
	userTokenAccount, _, _ := common.FindAssociatedTokenAddress(wallet.PublicKey, common.PublicKeyFromString(inputMint))
	balance, _, err := c.GetTokenAccountBalanceWithConfig(context.TODO(), userTokenAccount.ToBase58(), rpc.GetTokenAccountBalanceConfig{
		Commitment: rpc.CommitmentConfirmed,
	})
	if err != nil {
		return err
	}
	if balance <= 0 {
		return fmt.Errorf("input balance must be greater than zero")
	}
	err = Swap(
		inputMint,
		outputMint,
		balance,
		wallet,
		c,
	)
	if err != nil {
		return err
	}
	return nil
}
