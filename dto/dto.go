package dto

import (
	"time"
)

type ConfigDTO struct {
	Port       string `json:"Port"`
	Screenshot string `json:"Screenshot"`
}

type RequestCheckDTO struct {
	Index     int64             `json:"index"`
	URL       string            `json:"url"`
	Scenarios *[]ReqScenarioDTO `json:"scenario"`
	SSLCheck  bool              `json:"sslcheck"`
	Timeout   int64             `json:"timeout"`
}

/*
Types
1. Contents Check: Text
2. Click: Target, WaitTime
3. Text Input: Target, Text
4. Wait: WaitTime
*/
type ReqScenarioDTO struct {
	Index    int64   `json:"index"`
	Type     int     `json:"type"`
	Target   *string `json:"target"`
	Text     *string `json:"text"`
	WaitTime *int    `json:"waittime"`
}

type ResponseCheckDTO struct {
	Index     int64                  `json:"index"`
	URL       string                 `json:"url"`
	Scenarios *[]RespScenarioDTO     `json:"scenario"`
	SSL       *SSLDTO                `json:"ssl"`
	Status    int                    `json:"status"`
	ExecTime  int                    `json:"exectime"`
	Headers   map[string]interface{} `json:"headers"`
	FailURL   []StatusDTO            `json:"failURL"`
}

type RespScenarioDTO struct {
	Index      int64  `json:"index"`
	Syntax     bool   `json:"syntax"`
	Result     *bool  `json:"result"`
	ExecTime   int    `json:"exectime"`
	Screenshot string `json:"screenshot"`
}

type ResponseDTO struct {
	Status       int              `json:"status"`
	Message      string           `json:"message"`
	ResponseData ResponseCheckDTO `json:"responseData"`
}

type SSLDTO struct {
	Issuer     string    `json:"issuer"`
	DNSNames   []string  `json:"dnsnames"`
	IssueDate  time.Time `json:"issuedate"`
	ExpireDate time.Time `json:"expiredate"`
}

type StatusDTO struct {
	URL    string `json:"url"`
	Status int    `json:"status"`
}
