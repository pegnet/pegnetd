(ConversionPricing on wiki)

# Calculating a Conversion

In Peget, the miners are the oracles that report the market prices for each given asset. The conversion price however the conversion price used is not just the miner reported rate. Instead a solution we are calling **ARC** (Average Rate Conversion) is used to find the conversion rate. This solution reduces the volatility of crypto assets on peget, which is necessary given the **0% slippage** and **unlimited liquidity** on its pegged assets.

To read about how ARC works, all the formulas are detailed in our docs [here](https://pegnet.org/docs/pdfdocs/calculatingconversions.pdf)

As a trader, to know what the effects of arc are, `pegnetd` has a command called [`get price`](https://github.com/pegnet/pegnetd/wiki/cli#get-price) (or for developers, the api call is [`get-pegnet-spreads`](https://github.com/pegnet/pegnetd/wiki/API#get-pegnet-spreads)) that will return the buy and sell prices of an asset or pair at a given height. The prices of the next block are not known but can be estimated by taking the prices of the current height. The closer the buy and sell price are together, the less volatile the market is, and the tighter the spread between the buy and sell will be. If the current market price is outside the current buy/sell price, then the next buy/sell price could have a larger spread. It should be known all spreads are reduced by 1% on both sides of the price on pegnet so that low volatile markets have a 0% spread.


---------

(API)

### `get-pegnet-spreads`:

Returns the pegnet conversion pricing and spread information. This can be used to find the buy/sell price for any given asset pairing. To calculate the buy/sell price, consult the Pricing section near the bottom of the [Calculating Conversions](https://pegnet.org/docs/pdfdocs/calculatingconversions.pdf) document.

#### Request:

```
curl -X POST --data-binary \
'{"jsonrpc": "2.0", "id": 0, "method":"get-pegnet-spreads",
"params":{"height":206920}}' \
-H 'content-type:text/plain;' http://localhost:8070/v1
```

#### Response:

```json
{
  "jsonrpc": "2.0",
  "result": {
    "PEG": {
      "marketrate": 0,
      "movingaverage": 0
    },
    "pADA": {
      "marketrate": 4980000,
      "movingaverage": 4968882
    },
    "pBNB": {
      "marketrate": 2718420000,
      "movingaverage": 2701532388
    },
    "pBRL": {
      "marketrate": 24270000,
      "movingaverage": 24527085
    },
    "pCAD": {
      "marketrate": 75000000,
      "movingaverage": 75072951
    },

   
    ...

    "pXAG": {
      "marketrate": 1741790000,
      "movingaverage": 1707190612
    },
    "pXAU": {
      "marketrate": 152671750000,
      "movingaverage": 149954583750
    },
    "pXBC": {
      "marketrate": 31654770000,
      "movingaverage": 30954527375
    },
    "pXBT": {
      "marketrate": 1040807850000,
      "movingaverage": 1016876596492
    },
    "pXLM": {
      "marketrate": 6870000,
      "movingaverage": 6672363
    },
    "pXMR": {
      "marketrate": 8215260000,
      "movingaverage": 8182495749
    },
    "pZEC": {
      "marketrate": 5107760000,
      "movingaverage": 5091501388
    }
  },
  "id": 0
}
```

<br/>




-------

(CLI)

## get price

To get the buy/sell conversion price of a given asset at a height. The buy/sell price depends on the market and moving price. [Why is the buy and sell price different than the market rate?](https://github.com/pegnet/pegnetd/wiki/ConversionPricing)

```
$ pegnetd get price 211500 pXBT
Price of 1.0 pXBT:
         Market Rate: 8448.27946241 pUSD
           Buy Price: 8928.02679651 pUSD
          Sell Price: 8448.27946241 pUSD
```

To get the buy/sell price for a trading pair:

```
$ pegnetd get price 211500 pFCT pXBT
Price of 1.0 pFCT:
         Market Rate: 0.00031773 pXBT
           Buy Price: 0.00034668 pXBT
          Sell Price: 0.00030065 pXBT

pUSD Prices
pFCT:
        Market Price: 2.68427803 pUSD
           Buy Price: 2.92892056 pUSD
          Sell Price: 2.68427803 pUSD
pXBT:
        Market Price: 8448.27946241 pUSD
           Buy Price: 8928.02679651 pUSD
          Sell Price: 8448.27946241 pUSD

```

