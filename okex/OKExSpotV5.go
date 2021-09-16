package okex

import (
	"fmt"
	"github.com/go-openapi/errors"
	. "github.com/lucas7788/goex"
	"sort"
	"strconv"
	"strings"
	"time"
)

type OKExSpotV5 struct {
	*OKEx
}

//OrdType 订单类型
//market：市价单
//limit：限价单
//post_only：只做maker单
//fok：全部成交或立即取消
//ioc：立即成交并取消剩余
//optimal_limit_ioc：市价委托立即成交并取消剩余（仅适用交割、永续）
type OrderParamV5 struct {
	InstId     string `json:"instId"` //产品ID
	TdMode     string `json:"tdMode"`
	Ccy        string `json:"ccy"`
	ClOrdId    string `json:"clOrdId"`
	Tag        string `json:"tag"`
	Side       string `json:"side"` //订单方向 buy：买 sell：卖
	PosSide    string `json:"posSide"`
	OrdType    string `json:"ordType"` //
	Sz         string `json:"sz"`      //委托数量
	Px         string `json:"px"`      //委托价格，
	ReduceOnly bool   `json:"reduceOnly"`
	TgtCcy     string `json:"tgtCcy"` //委托数量的类型 base_ccy：交易货币 ；quote_ccy：计价货币 仅适用于币币订单
}

type PlaceOrderResponseV5 struct {
	OrdId   string `json:"ordId"`
	ClOrdId string `json:"clOrdId"`
	Tag     string `json:"tag"`
	SCode   string `json:"sCode"`
	SMsg    string `json:"sMsg"`
}

func (ok *OKExSpotV5) PlaceOrder(ty string, ord *Order) (*Order, error) {
	urlPath := "/api/v5/trade/order"
	param := OrderParamV5{
		ClOrdId: GenerateOrderClientId(32),
		InstId:  ord.Currency.AdaptUsdToUsdt().ToUpper().ToSymbol("-"),
	}
	switch ord.Side {
	case BUY, SELL:
		param.Side = strings.ToLower(ord.Side.String())
		param.Px = FloatToString(ord.Price, 5)
		param.Sz = FloatToString(ord.Amount, 5)
	case SELL_MARKET:
		param.TdMode = "cash"
		param.Side = "sell"
		param.Sz = FloatToString(ord.Amount, 5)
	case BUY_MARKET:
		param.TdMode = "cash"
		param.Side = "buy"
		param.Sz = FloatToString(ord.Amount, 5)
	default:
		panic("not support")
	}

	switch ty {
	case "limit":
		param.OrdType = "limit"
		param.TdMode = "cash"
	case "market":
		param.OrdType = "market"
	case "post_only":
		param.OrdType = "post_only"
	case "fok":
		param.OrdType = "fok"
	case "ioc":
		param.OrdType = "ioc"
	}

	param.Sz = FloatToString(ord.Amount, 5)

	jsonStr, _, _ := ok.OKEx.BuildRequestBody(param)
	fmt.Println("jsonStr:", jsonStr)
	var response OKRes
	err := ok.OKEx.DoRequest("POST", urlPath, jsonStr, &response)
	if err != nil {
		return nil, err
	}

	if response.Code != "0" {
		fmt.Println("response.Data: %v", response.Data)
		return nil, errors.New(int32(ToInt(response.Code)), response.Msg)
	}

	res := response.Data.([]interface{})
	if len(res) == 0 {
		return nil, fmt.Errorf("take order failed")
	}
	r := res[0].(map[string]interface{})
	if r["sCode"] != "0" {
		return nil, fmt.Errorf("take order failed, erroCode: %s, errorMsg:%s", r["sCode"], r["sMsg"])
	}
	ord.Cid = r["clOrdId"].(string)
	ord.OrderID2 = r["ordId"].(string)
	return ord, nil
}

func (ok *OKExSpotV5) MarketBuy(amount, price string, currency CurrencyPair) (*Order, error) {
	return ok.PlaceOrder("market", &Order{
		Price:    ToFloat64(price),
		Amount:   ToFloat64(amount),
		Currency: currency,
		Side:     BUY_MARKET,
	})
}

func (ok *OKExSpotV5) MarketSell(amount, price string, currency CurrencyPair) (*Order, error) {
	return ok.PlaceOrder("market", &Order{
		Price:    ToFloat64(price),
		Amount:   ToFloat64(amount),
		Currency: currency,
		Side:     SELL_MARKET,
	})
}

func (ok *OKExSpotV5) LimitSell(amount, price string, currency CurrencyPair, opt ...LimitOrderOptionalParameter) (*Order, error) {
	ty := "limit"
	if len(opt) > 0 {
		ty = opt[0].String()
	}
	return ok.PlaceOrder(ty, &Order{
		Price:    ToFloat64(price),
		Amount:   ToFloat64(amount),
		Currency: currency,
		Side:     SELL,
	})
}

func (ok *OKExSpotV5) LimitBuy(amount, price string, currency CurrencyPair, opt ...LimitOrderOptionalParameter) (*Order, error) {
	ty := "limit"
	if len(opt) > 0 {
		ty = opt[0].String()
	}
	return ok.PlaceOrder(ty, &Order{
		Price:    ToFloat64(price),
		Amount:   ToFloat64(amount),
		Currency: currency,
		Side:     BUY,
	})
}

func (ok *OKExSpotV5) GetTicker(currency CurrencyPair) (*Ticker, error) {
	urlPath := fmt.Sprintf("/api/v5/market/ticker?instId=%s", currency.AdaptUsdToUsdt().ToSymbol("-"))
	var responses OKRes
	err := ok.OKEx.DoRequest("GET", urlPath, "", &responses)
	if err != nil {
		return nil, err
	}
	if responses.Code != "0" {
		return nil, fmt.Errorf("responses.Code: %s", responses.Code)
	}
	dat := responses.Data.([]interface{})
	if len(dat) < 1 {
		return nil, fmt.Errorf("no pair: %s", currency.String())
	}
	response := dat[0].(map[string]interface{})
	date, _ := time.Parse(time.RFC3339, response["ts"].(string))
	last, _ := strconv.ParseFloat(response["last"].(string), 64)
	high24h, _ := strconv.ParseFloat(response["high24h"].(string), 64)
	low24h, _ := strconv.ParseFloat(response["low24h"].(string), 64)
	sell, _ := strconv.ParseFloat(response["askPx"].(string), 64)
	buy, _ := strconv.ParseFloat(response["bidPx"].(string), 64)
	vol, _ := strconv.ParseFloat(response["volCcy24h"].(string), 64)
	return &Ticker{
		Pair: currency,
		Last: last,
		High: high24h,
		Low:  low24h,
		Sell: sell,
		Buy:  buy,
		Vol:  vol,
		Date: uint64(time.Duration(date.UnixNano() / int64(time.Millisecond)))}, nil
}

func (ok *OKExSpotV5) GetDepth(size int, currency CurrencyPair) (*Depth, error) {

	urlPath := fmt.Sprintf("/api/v5/market/books?instId=%s&sz=%d", currency.AdaptUsdToUsdt().ToSymbol("-"), size)

	//var response struct {
	//	Asks      [][]interface{} `json:"asks"`
	//	Bids      [][]interface{} `json:"bids"`
	//	Timestamp string          `json:"timestamp"`
	//}
	var response OKRes
	err := ok.OKEx.DoRequest("GET", urlPath, "", &response)
	if err != nil {
		return nil, err
	}
	if response.Code != "0" {
		return nil, fmt.Errorf("responses.Code: %s", response.Code)
	}
	dep := new(Depth)
	dep.Pair = currency
	r := response.Data.([]interface{})
	res := r[0].(map[string]interface{})
	dep.UTime, _ = time.Parse(time.RFC3339, res["ts"].(string))
	as := res["asks"].([]interface{})
	for _, itm := range as {
		i := itm.([]interface{})
		dep.AskList = append(dep.AskList, DepthRecord{
			Price:  ToFloat64(i[0].(string)),
			Amount: ToFloat64(i[1].(string)),
		})
	}
	bs := res["bids"].([]interface{})
	for _, itm := range bs {
		i := itm.([]interface{})
		dep.BidList = append(dep.BidList, DepthRecord{
			Price:  ToFloat64(i[0]),
			Amount: ToFloat64(i[1]),
		})
	}

	sort.Sort(sort.Reverse(dep.AskList))

	return dep, nil
}
