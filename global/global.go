package global

import (
	"math/big"
)

type Config struct {
  ProgramID string		`json:"programID"`
  Assets []Asset			`json:"assets"`
  Oracles Oracles			`json:"oracles"`
  Markets []Market		`json:"markets"`
}
type Asset struct {
  Name string					`json:"name"`
  Symbol string				`json:"symbol"`
  Decimals uint8			`json:"decimals"`
  MintAddress string	`json:"mintAddress"`
}
type Oracles struct {
  PythProgramID string					`json:"pythProgramID"`
  SwitchboardProgramID string		`json:"switchboardProgramID"`
  Assets []OracleAsset					`json:"assets"`
}
type OracleAsset struct {
  Asset string									`json:"asset"`
  PriceAddress string						`json:"priceAddress"`
  SwitchboardFeedAddress string	`json:"switchboardFeedAddress"`
}
type Market struct {
  Name string							`json:"name"`
  Address string					`json:"address"`
  AuthorityAddress string	`json:"authorityAddress"`
  Reserves []Reserve			`json:"reserves"`
}
type Reserve struct {
  Asset string												`json:"asset"`
  Address string											`json:"address"`
  CollateralMintAddress string				`json:"collateralMintAddress"`
  CollateralSupplyAddress string			`json:"collateralSupplyAddress"`
  LiquidityAddress string							`json:"liquidityAddress"`
  LiquidityFeeReceiverAddress string	`json:"liquidityFeeReceiverAddress"`
  UserSupplyCap uint8									`json:"userSupplyCap"`
}

type OracleToken struct {
	Symbol string
	ReserveAddress string
	MintAddress string
	Decimals *big.Rat
	Price *big.Rat
}

type WalletTokenData struct {
  Balance float64
  BalanceBase int64
  Symbol string
}