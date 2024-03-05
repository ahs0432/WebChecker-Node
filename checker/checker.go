package checker

import (
	"container/list"
	"context"
	"crypto/x509"
	"encoding/base64"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"false.kr/WebChecker-Node/dto"
	"false.kr/WebChecker-Node/files"
	"github.com/chromedp/cdproto/har"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type urlStatus struct {
	url    string
	status int
}

type hRequest har.Request
type hResponse har.Response

// https://github.com/chromedp/chromedp/issues/42
// process requests and return a structured data
func processRequest(r *network.EventRequestWillBeSent) *hRequest {
	req := hRequest{}
	req.Method = r.Request.Method
	req.URL = r.Request.URL
	req.HTTPVersion = ""
	req.Headers = []*har.NameValuePair{}
	for header := range r.Request.Headers {
		h := har.NameValuePair{}
		h.Name = header
		h.Value = r.Request.Headers[header].(string)
		req.Headers = append(req.Headers, &h)
	}
	req.Cookies = []*har.Cookie{}
	req.QueryString = []*har.NameValuePair{}
	u, err := url.Parse(req.URL)
	if err != nil {
		log.Printf("[E] Invalid URL data recived : %v", err)
	}

	for name := range u.Query() {
		if len(name) != 0 {
			values := u.Query()[name]
			for _, val := range values {
				req.QueryString = append(req.QueryString, &har.NameValuePair{
					Name:  name,
					Value: val,
				})
			}
		}
	}

	req.PostData = nil
	req.HeadersSize = 0
	req.BodySize = 0
	return &req
}

func processResponse(r *network.EventResponseReceived) *hResponse {
	res := hResponse{}
	res.Status = r.Response.Status
	res.StatusText = r.Response.StatusText
	res.HTTPVersion = r.Response.Protocol
	res.Cookies = nil
	res.Headers = []*har.NameValuePair{}
	for header := range r.Response.Headers {
		h := har.NameValuePair{}
		h.Name = header
		h.Value = r.Response.Headers[header].(string)
		res.Headers = append(res.Headers, &h)
	}
	res.Content = &har.Content{}
	res.Content.MimeType = r.Response.MimeType
	res.Content.Size = 0

	res.RedirectURL = ""
	res.HeadersSize = 0
	res.BodySize = 0
	return &res
}

func getCertificate(ctx context.Context, url string) (*dto.SSLDTO, error) {
	sslcert, err := network.GetCertificate(url).Do(ctx)
	if err != nil {
		log.Printf("["+url+"] ", "Error decoding string: %s ", err.Error())
		return &dto.SSLDTO{}, err
	}

	if len(sslcert) != 0 {
		sDec, err := base64.StdEncoding.DecodeString(sslcert[0])
		if err != nil {
			log.Printf("["+url+"] ", "Error decoding string: %s ", err.Error())
			return &dto.SSLDTO{}, err
		}

		if ssl, err := x509.ParseCertificate(sDec); err != nil {
			log.Printf("["+url+"] ", "Error decoding string: %s ", err.Error())
			return &dto.SSLDTO{}, err
		} else {
			sslData := &dto.SSLDTO{}
			sslData.Issuer, sslData.DNSNames, sslData.IssueDate, sslData.ExpireDate = ssl.Issuer.Organization[0]+" "+ssl.Issuer.CommonName, ssl.DNSNames, ssl.NotBefore, ssl.NotAfter

			log.Print("["+url+"] ", "SSL Issuer: ", ssl.Issuer)
			log.Print("["+url+"] ", "SSL DNS Names: ", ssl.DNSNames)
			log.Print("["+url+"] ", "SSL Issue Date: ", ssl.NotBefore)
			log.Print("["+url+"] ", "SSL Expire Date: ", ssl.NotAfter)

			return sslData, nil
		}
	} else {
		return &dto.SSLDTO{}, nil
	}
}

func findContents(ctx context.Context, contents string) (*bool, error) {
	var outerHTML string
	check := false
	if err := chromedp.OuterHTML("html", &outerHTML, chromedp.ByQuery).Do(ctx); err != nil {
		return &check, err
	}

	if strings.Contains(outerHTML, contents) {
		check = true
	}
	outerHTML = ""
	return &check, nil
}

func CheckPage(requestData *dto.RequestCheckDTO, baseURL string) dto.ResponseCheckDTO {
	responseData := dto.ResponseCheckDTO{}
	responseData.Index = requestData.Index
	responseData.URL = requestData.URL

	urlTmp := strings.Split(requestData.URL, ".")
	if len(strings.Split(urlTmp[len(urlTmp)-1], "/")) == 1 {
		requestData.URL += "/"
	}

	ctx, cancel := chromedp.NewContext(
		context.Background(),
	)
	defer cancel()

	if !requestData.SSLCheck {
		opts := append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.DisableGPU,
			chromedp.Flag("ignore-certificate-errors", "1"),
		)

		allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
		defer cancel()
		ctx, cancel = chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
		defer cancel()
	}

	// Timeout
	ctx, cancel = context.WithTimeout(ctx, time.Duration(requestData.Timeout)*time.Second)
	defer cancel()

	var ws sync.WaitGroup
	nRequests := list.New()
	nResponses := list.New()

	var statusCode int64
	var responseHeaders map[string]interface{}

	// Network HAR Data
	chromedp.ListenTarget(ctx, func(event interface{}) {
		switch v := event.(type) {
		case *network.EventRequestWillBeSent:
			ws.Add(1)
			go func(r *network.EventRequestWillBeSent) {
				req := processRequest(r)
				rm := map[network.RequestID]*hRequest{}
				rm[r.RequestID] = req
				nRequests.PushBack(rm)
				ws.Done()
			}(v)
		case *network.EventResponseReceived:
			// Response Status, Headers
			response := v.Response
			if response.URL == requestData.URL {
				statusCode = response.Status
				responseHeaders = response.Headers
			}

			ws.Add(1)
			go func(r *network.EventResponseReceived) {
				res := processResponse(r)
				rm := map[network.RequestID]*hResponse{}
				rm[r.RequestID] = res
				nResponses.PushBack(rm)
				ws.Done()
			}(v)
		}
	})

	start := time.Now()
	if err := chromedp.Run(
		ctx,
		network.Enable(),
		chromedp.EmulateViewport(1920, 1080),
		chromedp.Navigate(requestData.URL),
		//chromedp.WaitReady("body"),
		//chromedp.CaptureScreenshot(&buf),
		chromedp.ActionFunc(func(ctx context.Context) error {
			respScenarioData := []dto.RespScenarioDTO{{Index: 0, Syntax: true}}

			// URL Check Index 0
			var err error
			var buf []byte
			if err = chromedp.FullScreenshot(&buf, 50).Do(ctx); err != nil {
				buf = []byte{}
				return err
			}

			if err = os.WriteFile(files.Config.Screenshot+strconv.FormatInt(requestData.Index, 10)+"_0_screenshot.png", buf, 0o644); err != nil {
				buf = []byte{}
				log.Print("["+requestData.URL+"] ", err.Error())
			}

			respScenarioData[0].Screenshot = baseURL + "/images/" + strconv.FormatInt(requestData.Index, 10) + "_0"
			buf = []byte{}

			// Get SSL Certificate
			if requestData.SSLCheck && strings.Contains(requestData.URL, "https") {
				if responseData.SSL, err = getCertificate(ctx, requestData.URL); err != nil {
					return err
				}
			}

			endScenarioTime := time.Now()
			respScenarioData[0].ExecTime = int(endScenarioTime.Sub(start).Milliseconds())
			log.Print("["+requestData.URL+"] ", "Index ", 0, " Time Seconds(ms): ", respScenarioData[0].ExecTime)

			nowIndex := 0
			// Next Scenario Index 1~n
			if requestData.Scenarios != nil {
				for _, scenario := range *requestData.Scenarios {
					nowIndex++
					respScenarioData = append(respScenarioData, dto.RespScenarioDTO{Index: scenario.Index})
					startScenarioTime := time.Now()

					// Get Documents (Full Pages)
					if scenario.Type == 1 && scenario.Text != nil {
						respScenarioData[nowIndex].Syntax = true
						if respScenarioData[nowIndex].Result, err = findContents(ctx, *scenario.Text); err != nil {
							continue
						}
						log.Print("["+requestData.URL+"] ", "Request Contents Check: ", *scenario.Text)
					} else if scenario.Type == 2 && scenario.Target != nil && scenario.WaitTime != nil {
						respScenarioData[nowIndex].Syntax = true
						if err = chromedp.Click(*scenario.Target).Do(ctx); err != nil {
							continue
						}
						if *scenario.WaitTime > 0 {
							chromedp.Sleep(time.Second * time.Duration(*scenario.WaitTime)).Do(ctx)
						}
						log.Print("["+requestData.URL+"] ", "Request Click: ", *scenario.Target)
					} else if scenario.Type == 3 && scenario.Target != nil && scenario.Text != nil {
						respScenarioData[nowIndex].Syntax = true
						if err = chromedp.SendKeys(*scenario.Target, *scenario.Text).Do(ctx); err != nil {
							continue
						}
						log.Print("["+requestData.URL+"] ", "Request Text Input: ", *scenario.Text, " Target: ", *scenario.Target)
					} else if scenario.Type == 4 && scenario.WaitTime != nil {
						respScenarioData[nowIndex].Syntax = true
						if *scenario.WaitTime > 0 {
							chromedp.Sleep(time.Second * time.Duration(*scenario.WaitTime)).Do(ctx)
						}
					} else {
						respScenarioData[nowIndex].Syntax = false
						respScenarioData[nowIndex].Screenshot = ""
						break
					}

					if respScenarioData[nowIndex].Syntax && scenario.Type != 1 {
						if err = chromedp.FullScreenshot(&buf, 50).Do(ctx); err != nil {
							buf = []byte{}
							continue
						}

						if buf != nil {
							if err := os.WriteFile(files.Config.Screenshot+strconv.FormatInt(requestData.Index, 10)+"_"+strconv.FormatInt(scenario.Index, 10)+"_screenshot.png", buf, 0o644); err != nil {
								os.Remove(files.Config.Screenshot + strconv.FormatInt(requestData.Index, 10) + "_" + strconv.FormatInt(scenario.Index, 10) + "_screenshot.png")
								log.Print("["+requestData.URL+"] ", err.Error())
							}
							respScenarioData[nowIndex].Screenshot = baseURL + "/images/" + strconv.FormatInt(requestData.Index, 10) + "_" + strconv.FormatInt(scenario.Index, 10)
							buf = []byte{}
						}
					}
					endScenarioTime := time.Now()
					respScenarioData[nowIndex].ExecTime = int(endScenarioTime.Sub(startScenarioTime).Milliseconds())
					log.Print("["+requestData.URL+"] ", "Index ", scenario.Index, " Time Seconds(ms): ", respScenarioData[nowIndex].ExecTime)
				}
			}
			responseData.Scenarios = &respScenarioData

			return nil
		}),
	); err != nil {
		log.Print("["+requestData.URL+"] Error: ", err.Error())
	}

	end := time.Now()
	responseData.ExecTime = int(end.Sub(start).Milliseconds())
	log.Print("["+requestData.URL+"] ", "Total Time Seconds(ms): ", responseData.ExecTime)

	nStatus := make(map[network.RequestID]urlStatus)
	for e := nRequests.Front(); e != nil; e = e.Next() {
		r := e.Value.(map[network.RequestID]*hRequest)
		for k := range r {
			if thisStatus, ok := nStatus[k]; ok {
				thisStatus.url = r[k].URL
				nStatus[k] = thisStatus
			} else {
				thisStatus.url = r[k].URL
				thisStatus.status = 0
				nStatus[k] = thisStatus
			}
		}
	}

	for e := nResponses.Front(); e != nil; e = e.Next() {
		r := e.Value.(map[network.RequestID]*hResponse)
		for k := range r {
			if thisStatus, ok := nStatus[k]; ok {
				thisStatus.status = int(r[k].Status)
				nStatus[k] = thisStatus
			} else {
				thisStatus.url = ""
				thisStatus.status = int(r[k].Status)
				nStatus[k] = thisStatus
			}
		}
	}

	for s := range nStatus {
		if !(nStatus[s].status >= 200 && nStatus[s].status <= 399) {
			log.Print("["+requestData.URL+"] ", "URL: ", nStatus[s].url)
			log.Print("["+requestData.URL+"] ", "Status: ", nStatus[s].status)

			responseData.FailURL = append(responseData.FailURL, dto.StatusDTO{URL: nStatus[s].url, Status: nStatus[s].status})
		}
	}

	log.Print("["+requestData.URL+"] ", "Status Code: ", statusCode)
	responseData.Status = int(statusCode)
	log.Print("["+requestData.URL+"] ", "Headers: ", responseHeaders)
	responseData.Headers = responseHeaders
	ws.Wait()

	ctx.Done()
	return responseData
}
