package sub

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Data struct {
	Port int    `json:"port"`
	IP   string `json:"ip"`
}
type HunterResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	RsData  RsData `json:"data"`
}

type RsData struct {
	Total int    `json:"total"`
	Arr   []Data `json:"arr"`
}

// 定义用于解析 JSON 的结构体
type QuakeResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    []Data `json:"data"`
}

type FofaResponse struct {
	Error   bool       `json:"error"`
	Results [][]string `json:"results"` // 每个结果是一个 IP 和端口的切片
	Errmsg  string     `json:"errmsg"`
}

func GetDataFromHunter(hunter HUNTER) (HunterResponse, error) {
	var response HunterResponse
	if hunter.Switch != "open" {
		return response, fmt.Errorf("---未开启hunter---")
	}
	fmt.Printf("***已开启hunter,将根据配置条件从hunter中获取%d条数据\n", hunter.ResultSize)
	end := hunter.ResultSize / 100

	for i := 1; i <= end; i++ {
		params := map[string]string{
			"api-key":   hunter.Key,
			"search":    base64.URLEncoding.EncodeToString([]byte(hunter.QueryString)),
			"page":      strconv.Itoa(i),
			"page_size": "100"}
		fmt.Printf("HUNTER:每页100条,正在查询第%v页\n", i)
		content, err := fetchContent(hunter.APIUrl, "GET", 60, params, nil, "")
		if err != nil {
			return response, fmt.Errorf("访问hunter异常%w", err)
		}
		tmpData := make([]Data, len(response.RsData.Arr))
		copy(tmpData, response.RsData.Arr)

		err = json.Unmarshal([]byte(content), &response)
		if err != nil {
			response.RsData.Arr = tmpData
			return response, fmt.Errorf("解析hunter返回内容异常:%w", err)
		}
		if response.Code != 200 {
			response.RsData.Arr = tmpData
			return response, fmt.Errorf("HUNTER:%s", response.Message)
		}

		if response.RsData.Total == 0 {
			response.RsData.Arr = tmpData
			return response, fmt.Errorf("HUNTER:未取到数据")
		}
		response.RsData.Arr = append(tmpData, response.RsData.Arr...)
		if len(response.RsData.Arr) >= response.RsData.Total {
			break
		}
		if end > 1 && i != end {
			time.Sleep(3 * time.Second) //防止hunter提示访问过快获取不到结果
		}
	}
	fmt.Println("+++hunter数据已取+++")
	return response, nil
}

func GetDataFromQuake(quake QUAKE) (QuakeResponse, error) {
	var response QuakeResponse
	if quake.Switch != "open" {
		return response, fmt.Errorf("---未开启quake---")
	}

	fmt.Printf("***已开启quake,将根据配置条件从quake中获取%d条数据***\n", quake.ResultSize)
	jsonCondition := "{\"query\": \"" + strings.Replace(quake.QueryString, `"`, `\"`, -1) + "\",\"start\": 0,\"size\": " + strconv.Itoa(quake.ResultSize) + ",\"include\":[\"ip\",\"port\"]}"
	headers := map[string]string{
		"X-QuakeToken": quake.Key,
		"Content-Type": "application/json"}
	content, err := fetchContent(quake.APIUrl, "POST", 60, nil, headers, jsonCondition)
	if err != nil {
		return response, fmt.Errorf("访问quake异常:%w", err)
	}

	err = json.Unmarshal([]byte(content), &response)
	if err != nil {
		return response, fmt.Errorf("解析quake返回内容异常:%w", err)
	}

	if response.Code != 0 {
		return response, fmt.Errorf("QUAKE:%s", response.Message)
	}
	return response, nil
}

// 从FOFA获取,结果为IP:PORT
func GetDataFromFofa(fofa FOFA) (FofaResponse, error) {
	var response FofaResponse
	if fofa.Switch != "open" {
		return response, fmt.Errorf("---未开启fofa---")

	}
	fmt.Printf("***已开启fofa,将根据配置条件从fofa中获取%d条数据***\n", fofa.ResultSize)

	params := map[string]string{
		"email":   fofa.Email,
		"key":     fofa.Key,
		"fields":  "ip,port",
		"qbase64": base64.URLEncoding.EncodeToString([]byte(fofa.QueryString)),
		"size":    strconv.Itoa(fofa.ResultSize)}
	content, err := fetchContent(fofa.APIUrl, "GET", 60, params, nil, "")
	if err != nil {
		return response, fmt.Errorf("访问fofa异常:%w", err)
	}

	// 解析 JSON 数据
	err = json.Unmarshal([]byte(content), &response)
	if err != nil {
		return response, fmt.Errorf("解析fofa返回内容异常:%w", err)
	}
	if response.Error {
		return response, fmt.Errorf("解析fofa返回内容异常:%s", response.Errmsg)
	}
	return response, nil

}

func fetchContent(baseURL string, method string, timeout int, urlParams map[string]string, headers map[string]string, jsonBody string) (string, error) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: time.Duration(timeout) * time.Second,
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	if urlParams != nil {
		q := u.Query()
		for key, value := range urlParams {
			q.Set(key, value)
		}
		u.RawQuery = q.Encode()
	}

	var req *http.Request
	if jsonBody != "" {
		req, err = http.NewRequest(method, u.String(), bytes.NewBufferString(jsonBody))
	} else {
		req, err = http.NewRequest(method, u.String(), nil)
	}

	if err != nil {
		return "", err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36 Edg/112.0.1722.17")
	if len(headers) != 0 {
		for key, value := range headers {
			req.Header.Add(key, value)
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func ConvertFofaResultsToData(results [][]string) []Data {
	var dataSlice []Data
	for _, row := range results {
		ip := row[0]
		portStr := row[1]
		port, _ := strconv.Atoi(portStr)
		dataSlice = append(dataSlice, Data{Port: port, IP: ip})
	}
	return dataSlice
}

func RemoveDuplicates(dataList []Data) []Data {
	uniqueMap := make(map[Data]struct{})

	for _, data := range dataList {
		uniqueMap[data] = struct{}{}
	}

	var uniqueList []Data
	for data := range uniqueMap {
		uniqueList = append(uniqueList, data)
	}

	return uniqueList
}
