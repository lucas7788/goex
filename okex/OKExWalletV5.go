package okex

import (
	"encoding/json"
	"fmt"
)

type OKExWalletV5 struct {
	*OKEx
}

type AcctBalance struct {
	AvailBal  string  `json:"availBal"`  //可用余额
	CashBal   string `json:"cashBal"`   // 币种余额
	Ccy       string  `json:"ccy"`       //币种，如 BTC
	FrozenBal string  `json:"frozenBal"` //币种占用金额
	OrdFrozen string  `json:"ordFrozen"` //挂单冻结数量
}

func (ok *OKExWalletV5) GetAccountBalance(ccy ...string) ([]*AcctBalance, error) {
	if len(ccy) < 1 {
		return nil, fmt.Errorf("invalid ccy")
	}
	param := ccy[0]
	for i := 1; i < len(ccy); i++ {
		param += "," + ccy[i]
	}
	url := "/api/v5/account/balance?ccy=" + param
	var res OKRes
	err := ok.DoRequest("GET", url, "", &res)
	if err != nil {
		return nil, err
	}
	if res.Code != "0" {
		return nil, fmt.Errorf("res.Code:%s,res.Msg:%s", res.Code, res.Msg)
	}
	r := res.Data.([]interface{})
	accs := make([]*AcctBalance, 0)
	for _, itm := range r {
		i := itm.(map[string]interface{})
		arr := i["details"].([]interface{})
		for _, j := range arr {
			data, err := json.Marshal(j)
			if err != nil {
				panic(err)
			}
			var acc AcctBalance
			err = json.Unmarshal(data, &acc)
			if err != nil {
				panic(err)
			}
			accs = append(accs, &acc)
		}
	}
	return accs, nil
}
