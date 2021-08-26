package okex

import (
	"encoding/json"
	"fmt"
)

type Balance struct {
	AvailBal  string  `json:"availBal"`  //可用余额
	Bal       float64 `json:"bal"`       // 余额
	Ccy       string  `json:"ccy"`       //币种，如 BTC
	FrozenBal string  `json:"frozenBal"` //冻结（不可用）
}

type OKExAssetV5 struct {
	*OKEx
}

type CurrencyOKRes struct {
	Ccy         string `json:"ccy"`         //币种名称，如 BTC
	Name        string `json:"name"`        //币种中文名称，不显示则无对应名称
	Chain       string `json:"chain"`       //币种链信息 有的币种下有多个链，必须要做区分，如USDT下有USDT-ERC20，USDT-TRC20，USDT-Omni多个链
	CanDep      bool   `json:"canDep"`      //是否可充值，false表示不可链上充值，true表示可以链上充值
	CanWd       bool   `json:"canWd"`       //是否可提币，false表示不可链上提币，true表示可以链上提币
	CanInternal bool   `json:"canInternal"` //是否可内部转账，false表示不可内部转账，true表示可以内部转账
	MinWd       string `json:"minWd"`       //币种最小提币量
	MinFee      string `json:"minFee"`      //最小提币手续费数量
	MaxFee      string `json:"maxFee"`      //最大提币手续费数量 获取资金账户余额 获取资金账户所有资产列
}

//获取平台所有币种列表。并非所有币种都可被用于交易。
func (self *OKExAssetV5) GetCurrencies() ([]*CurrencyOKRes, error) {

	url := "/api/v5/asset/currencies"
	var res OKRes
	err := self.DoRequest("GET", url, "", &res)
	if err != nil {
		return nil, err
	}
	if res.Code != "0" {
		return nil, fmt.Errorf("res code: %s, res.Msg: %s", res.Code, res.Msg)
	}
	r := res.Data.([]interface{})
	re := make([]*CurrencyOKRes, 0)
	for _, item := range r {
		data, err := json.Marshal(item)
		if err != nil {
			return nil, err
		}
		var curren CurrencyOKRes
		err = json.Unmarshal(data, &curren)
		if err != nil {
			return nil, err
		}
		re = append(re, &curren)
	}
	return re, nil
}

type DepositAddressOKRes struct {
	Chain    string `json:"chain"`    //币种链信息 有的币种下有多个链，必须要做区分，如USDT下有USDT-ERC20，USDT-TRC20，USDT-Omni多个链
	CtAddr   string `json:"ctAddr"`   //合约地址后6位
	Ccy      string `json:"ccy"`      //币种，如BTC
	To       string `json:"to"`       //转入账户 1：币币 3：交割合约 6：资金账户 9：永续合约 12：期权 18：统一账户
	Selected bool   `json:"selected"` //该地址是否为页面选中的地址
	Addr     string `json:"addr"`     //充值地址
}

//获取各个币种的充值地址，包括曾使用过的老地址。
//限速： 6次/s
func (self *OKExAssetV5) GetDepositAddress(ccy string) ([]*DepositAddressOKRes, error) {

	url := "/api/v5/asset/deposit-address?ccy=" + ccy
	var res OKRes
	err := self.DoRequest("GET", url, "", &res)
	if err != nil {
		return nil, err
	}
	if res.Code != "0" {
		return nil, fmt.Errorf("res code: %s, res.Msg: %s", res.Code, res.Msg)
	}
	r := res.Data.([]interface{})
	re := make([]*DepositAddressOKRes, 0)
	for _, item := range r {
		data, err := json.Marshal(item)
		if err != nil {
			return nil, err
		}
		var curren DepositAddressOKRes
		err = json.Unmarshal(data, &curren)
		if err != nil {
			return nil, err
		}
		re = append(re, &curren)
	}
	return re, nil
}

// 查询余额
func (self *OKExAssetV5) GetBalances(ccy ...string) ([]*Balance, error) {
	if len(ccy) < 1 {
		return nil, fmt.Errorf("invalid ccy")
	}
	param := ccy[0]
	for i := 1; i < len(ccy); i++ {
		param += "," + ccy[i]
	}
	url := "/api/v5/asset/balances?ccy=" + param
	var res OKRes
	err := self.DoRequest("GET", url, "", &res)
	if err != nil {
		return nil, err
	}
	if res.Code != "0" {
		return nil, fmt.Errorf("res code: %s, res.Msg: %s", res.Code, res.Msg)
	}
	r := res.Data.([]interface{})
	re := make([]*Balance, 0)
	for _, item := range r {
		data, err := json.Marshal(item)
		if err != nil {
			return nil, err
		}
		var balance Balance
		err = json.Unmarshal(data, &balance)
		if err != nil {
			return nil, err
		}
		re = append(re, &balance)
	}
	return re, nil
}

type TransferParamV5 struct {
	Ccy  string `json:"ccy"`  //币种，如 USDT
	Amt  string `json:"amt"`  //划转数量
	Ty   string `json:"type"` //0：账户内划转 1：母账户转子账户 2：子账户转母账户 默认为0。
	From string `json:"from"` //转出账户 1：币币账户 3：交割合约 5：币币杠杆账户 6：资金账户 9：永续合约账户 12：期权合约 18：统一账户
	To   string `json:"to"`   //转入账户 1：币币账户 3：交割合约 5：币币杠杆账户 6：资金账户 9：永续合约账户 12：期权合约 18：统一账户
}
type TransferResponseV5 struct {
	TransId string `json:"transId"` //划转 ID
	Ccy     string `json:"ccy"`     //币种，如 USDT
	Amt     string `json:"amt"`     //划转数量
	From    string `json:"from"`    //转出账户 1：币币账户 3：交割合约 5：币币杠杆账户 6：资金账户 9：永续合约账户 12：期权合约 18：统一账户
	To      string `json:"to"`      //转入账户 1：币币账户 3：交割合约 5：币币杠杆账户 6：资金账户 9：永续合约账户 12：期权合约 18：统一账户
}

//资产划转
//支持母账户的资金账户划转到交易账户，母账户到子账户的资金账户和交易账户划转。
//不支持子账户和子账户之间直接划转。
func (self *OKExAssetV5) Transfer(Ccy string, Amt string, Ty string, From string, To string) (*TransferResponseV5, error) {

	url := "/api/v5/asset/transfer"
	tf := TransferParamV5{
		Ccy:  Ccy,
		Amt:  Amt,
		Ty:   Ty,
		From: From,
		To:   To,
	}
	jsonStr, _, _ := self.BuildRequestBody(tf)
	fmt.Println("jsonStr:", jsonStr)
	var res OKRes
	err := self.DoRequest("POST", url, jsonStr, &res)
	if err != nil {
		return nil, err
	}
	if res.Code != "0" {
		return nil, fmt.Errorf("res code: %s, res.Msg: %s", res.Code, res.Msg)
	}
	re, ok := res.Data.([]interface{})
	if !ok || re == nil || len(re) < 1 {
		panic(jsonStr)
	}
	data, err := json.Marshal(re[0])
	if err != nil {
		panic(err)
	}
	var tfR TransferResponseV5
	err = json.Unmarshal(data, &tfR)
	if err != nil {
		panic(err)
	}
	return &tfR, nil
}

type WithdrawalParam struct {
	Ccy    string `json:"ccy"`    //	币种，如 USDT
	Chain  string `json:"chain"`  //	链
	Amt    string `json:"amt"`    //	数量
	Dest   string `json:"dest"`   //	提币到 3：欧易OKEx 4：数字货币地址
	ToAddr string `json:"toAddr"` //	认证过的数字货币地址、邮箱或手机号。 某些数字货币地址格式为:地址+标签，如 ARDOR-7JF3-8F2E-QUWZ-CAN7F:123456
	Pwd    string `json:"pwd"`    //	交易密码
	Fee    string `json:"fee"`    //	网络手续费≥0，提币到数字货币地址所需网络手续费可通过获取币种列表接口查询
}

type WithdrawalRes struct {
	Ccy   string `json:"ccy"`   //	币种，如 USDT
	Chain string `json:"chain"` //	链
	Amt   string `json:"amt"`   //	数量
	WdId  string `json:"wdId"`  //	提币申请ID
}

//用户提币。
//限速： 6次/s
func (self *OKExAssetV5) Withdrawal(chain string, Ccy string, Amt string, Dest string, ToAddr string, Pwd string, Fee string) (*WithdrawalRes, error) {

	url := "/api/v5/asset/withdrawal"

	wp := &WithdrawalParam{
		Ccy:    Ccy,
		Chain:  chain,
		Amt:    Amt,
		Dest:   Dest,
		ToAddr: ToAddr,
		Pwd:    Pwd,
		Fee:    Fee,
	}
	jsonStr, _, _ := self.BuildRequestBody(wp)
	fmt.Println("jsonStr:", jsonStr)
	var res OKRes
	err := self.DoRequest("POST", url, jsonStr, &res)
	if err != nil {
		return nil, err
	}
	if res.Code != "0" {
		return nil, fmt.Errorf("res code: %s, res.Msg: %s", res.Code, res.Msg)
	}
	re, ok := res.Data.([]interface{})
	if !ok || re == nil || len(re) < 1 {
		panic(jsonStr)
	}
	data, err := json.Marshal(re[0])
	if err != nil {
		panic(err)
	}
	var tfR WithdrawalRes
	err = json.Unmarshal(data, &tfR)
	if err != nil {
		panic(err)
	}
	return &tfR, nil
}
