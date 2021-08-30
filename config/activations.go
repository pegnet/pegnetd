package config

import (
	"github.com/Factom-Asset-Tokens/factom"
	"github.com/pegnet/pegnetd/fat/fat2"
)

var (
	OPRChain         = factom.NewBytes32("a642a8674f46696cc47fdb6b65f9c87b2a19c5ea8123b3d2f0c13b6f33a9d5ef")
	SPRChain         = factom.NewBytes32("d5e395125335a21cef0ceca528168e87fe929fdac1f156870c1b1be6502448b4")
	TransactionChain = factom.NewBytes32("cffce0f409ebba4ed236d49d89c70e4bd1f1367d86402a3363366683265a242d")

	// Acivation Heights

	PegnetActivation    uint32 = 206421
	GradingV2Activation uint32 = 210330

	// TransactionConversionActivation indicates when tx/conversions go live on mainnet.
	// Target Activation Height is Oct 7, 2019 15 UTC
	TransactionConversionActivation uint32 = 213237

	// This is when PEG is priced by the market cap equation
	// Estimated to be Oct 14 2019, 15:00:00 UTC
	PEGPricingActivation uint32 = 214287

	// OneWaypFCTConversions makes pFCT a 1 way conversion. This means pFCT->pXXX,
	// but no asset can go into pFCT. AKA pXXX -/> pFCT.
	// The only way to aquire pFCT is to burn FCT. The burn command will remain.
	// Estimated to be Nov 25, 2019 17:47:00 UTC
	OneWaypFCTConversions uint32 = 220346

	// Once this is activated, a maximum amount of PEG of 5,000 can be
	// converted per block. At a future height, a dynamic bank should be used.
	// Estimated to be  Dec 9, 2019, 17:00 UTC
	PegnetConversionLimitActivation uint32 = 222270

	// This is when PEG price is determined by the exchange price
	// Estimated to be  Dec 9, 2019, 17:00 UTC
	PEGFreeFloatingPriceActivation uint32 = 222270

	// V4OPRUpdate indicates the activation of additional currencies and ecdsa keys.
	// Estimated to be  Feb 12, 2020, 18:00 UTC
	V4OPRUpdate uint32 = 231620

	// V20HeightActivation indicates the activation of PegNet 2.0.
	// Estimated to be  Aug 19th 2020 14:00 UTC
	V20HeightActivation uint32 = 258796

	// Activation height for developer rewards
	V20DevRewardsHeightActivation uint32 = 260118

	// SprSignatureActivation indicates the activation of SPR Signature.
	// Estimated to be  Aug 28th 2020
	SprSignatureActivation uint32 = 260118

	// OneWaypAssetsConversions makes some pAssets a 1 way conversion.
	// pDCR, pDGB, pDOGE, pHBAR, pONT, pRVN, pBAT, pALGO, pBIF, pETB, pKES, pNGN, pRWF, pTZS, pUGX
	// These pAssets have got small marketcap, and these will be disabled for conversion.
	// Estimated to be Dec 3th 2020
	OneWaySmallAssetsConversions uint32 = 274036

	// V202EnhanceActivation indicates the activation of PegNet 2.0.2.
	// Estimated to be  Dec 3th 2020
	V202EnhanceActivation uint32 = 274036

	// V204EnhanceActivation indicates the activation of PegNet 2.0.4.
	// Estimated to be  Mar 16th 2021
	V204EnhanceActivation uint32 = 288878

	// V204EnhanceActivation indicates the activation that burns remaining airdrop amount.
	// Estimated to be  April 16th 2021
	V204BurnMintedTokenActivation uint32 = 294206

	// PIP10AveragingActivation changes conversions to use the lesser of a rolling average and market price
	// for the source of a conversion, and the higher of the rolling average and the market price for the
	// target of a conversion
	//
	// Activation of 2.0.5
	PIP10AverageActivation uint32 = 295190

	// PIP18DelegateStakingActivation implements delegate staking by using PEG addresses.
	//	1. Balances of PEG for each address is more complicated.  It is the balance of PEG for the address (assuming it has not be delegated)
	//	2. We quit looking at the rich list, and just consider the top 100 submissions with the highest stake
	//	3. We pay out with the ratio of the total PEG staked. (Removes old top 100 PEG addresses staking reward and give staking opportunity to all PEG holders)
	PIP18DelegateStakingActivation uint32 = 313906
)

func SetAllActivations(act uint32) {
	PegnetActivation = act
	GradingV2Activation = act
	TransactionConversionActivation = act
	PEGPricingActivation = act
	OneWaypFCTConversions = act
	PegnetConversionLimitActivation = act
	PEGFreeFloatingPriceActivation = act
	fat2.Fat2RCDEActivation = act
	V4OPRUpdate = act
	V20HeightActivation = act
	V20DevRewardsHeightActivation = act
	OneWaySmallAssetsConversions = act
	SprSignatureActivation = act
	V202EnhanceActivation = act
	V204EnhanceActivation = act
	V204BurnMintedTokenActivation = act
	PIP10AverageActivation = act
	PIP18DelegateStakingActivation = act
}
