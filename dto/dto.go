package dto

import (
	"time"
)

type ConfigDTO struct {
	Port       string `json:"Port"`
	Screenshot string `json:"Screenshot"`
}

type RequestCheckDTO struct {
	Index     int64             `json:"Index"`
	URL       string            `json:"URL"`
	Scenarios *[]ReqScenarioDTO `json:"Scenario"`
	SSLCheck  bool              `json:"SSLCheck"`
	Timeout   int64             `json:"Timeout"`
}

/*
Types
1. Contents Check: Text
2. Click: Target, WaitTime
3. Text Input: Target, Text
4. Wait: WaitTime
*/
type ReqScenarioDTO struct {
	Index    int64   `json:"Index"`
	Type     int     `json:"Type"`
	Target   *string `json:"Target"`
	Text     *string `json:"Text"`
	WaitTime *int    `json:"WaitTime"`
}

type ResponseCheckDTO struct {
	Index     int64                  `json:"Index"`
	URL       string                 `json:"URL"`
	Scenarios *[]RespScenarioDTO     `json:"Scenario"`
	SSL       *SSLDTO                `json:"SSL"`
	Status    int                    `json:"Status"`
	ExecTime  int                    `json:"ExecTime"`
	Headers   map[string]interface{} `json:"Headers"`
	FailURL   []StatusDTO            `json:"FailURL"`
}

type RespScenarioDTO struct {
	Index      int64  `json:"Index"`
	Syntax     bool   `json:"Syntax"`
	Result     *bool  `json:"Result"`
	ExecTime   int    `json:"ExecTime"`
	Screenshot string `json:"Screenshot"`
}

type ResponseDTO struct {
	Status       int              `json:"Status"`
	Message      string           `json:"Message"`
	ResponseData ResponseCheckDTO `json:"ResponseData"`
}

type SSLDTO struct {
	Issuer     string    `json:"Issuer"`
	DNSNames   []string  `json:"DNSNames"`
	IssueDate  time.Time `json:"IssueDate"`
	ExpireDate time.Time `json:"ExpireDate"`
}

type StatusDTO struct {
	URL    string `json:"URL"`
	Status int    `json:"Status"`
}
