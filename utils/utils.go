package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/rpc"
)

func JsonFromObject(obj any) string {
	js, _ := json.MarshalIndent(obj, "", "  ")
	return string(js)
}

func RpcProgramAccountInfoToClientAccountInfo(v rpc.GetProgramAccountsAccount) (client.AccountInfo, error) {
	if v == (rpc.GetProgramAccountsAccount{}) {
		return client.AccountInfo{}, nil
	}

	data, ok := v.Data.([]any)
	if !ok {
		return client.AccountInfo{}, fmt.Errorf("failed to cast raw response to []interface{}")
	}
	if data[1] != string(rpc.GetAccountInfoConfigEncodingBase64) {
		return client.AccountInfo{}, fmt.Errorf("encoding mistmatch")
	}
	rawData, err := base64.StdEncoding.DecodeString(data[0].(string))
	if err != nil {
		return client.AccountInfo{}, fmt.Errorf("failed to base64 decode data")
	}
	return client.AccountInfo{
		Lamports:   v.Lamports,
		Owner:      v.Owner,
		Executable: v.Executable,
		RentEpoch:  v.RentEpoch,
		Data:       rawData,
	}, nil
}

func Reverse(arr []byte) []byte {
	for i, j := 0, len(arr)-1; i < j; i, j = i+1, j-1 {
		arr[i], arr[j] = arr[j], arr[i]
	}
	return arr
}

func BigIntFromBytes(bs []byte) *big.Int {
	return new(big.Int).SetBytes(Reverse(bs))
}

func BigRatFromBytes(bs []byte) *big.Rat {
	return new(big.Rat).SetInt(BigIntFromBytes(bs))
}

func RemoveDuplicateStr(strSlice []string) []string {
	allKeys := make(map[string]bool)
	list := []string{}
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}
