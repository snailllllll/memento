package sms

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"snail.local/snailllllll/utils"
)

// SIM卡类型
type SIMType int

const (
	SIM1   SIMType = 1
	SIM2   SIMType = 2
	SIMALL SIMType = 0
)

// 短信类型
type SMSType int

const (
	SMSAll      SMSType = 0
	SMSReceived SMSType = 1
	SMSSent     SMSType = 2
	SMSDraft    SMSType = 3
)

// Client SMS客户端
type Client struct {
	baseURL    string
	secret     string
	encryption string
	publicKey  string
	privateKey string
	sm4Key     string
	timeout    time.Duration
	httpClient *http.Client
}

// NewClient 创建新的SMS客户端
func NewClient() *Client {
	return &Client{
		baseURL:    utils.GetConfig("SMS_BASE_URL", "http://localhost:5000"),
		secret:     utils.GetConfig("SMS_SECRET", ""),
		encryption: utils.GetConfig("SMS_ENCRYPTION", "plain"),
		publicKey:  utils.GetConfig("SMS_PUBLIC_KEY", ""),
		privateKey: utils.GetConfig("SMS_PRIVATE_KEY", ""),
		sm4Key:     utils.GetConfig("SMS_SM4_KEY", ""),
		timeout:    time.Duration(parseInt(utils.GetConfig("SMS_TIMEOUT", "30"))) * time.Second,
		httpClient: &http.Client{},
	}
}

// parseInt 解析字符串为整数
func parseInt(s string) int {
	if s == "" {
		return 0
	}
	if val, err := strconv.Atoi(s); err == nil {
		return val
	}
	return 0
}

// Request 请求结构
type Request struct {
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
	Sign      string      `json:"sign,omitempty"`
}

// Response 响应结构
type Response struct {
	Code      int         `json:"code"`
	Msg       string      `json:"msg"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
	Sign      string      `json:"sign,omitempty"`
}

// ConfigResponse 配置查询响应
type ConfigResponse struct {
	EnableAPIBatteryQuery bool               `json:"enable_api_battery_query"`
	EnableAPICallQuery    bool               `json:"enable_api_call_query"`
	EnableAPIClone        bool               `json:"enable_api_clone"`
	EnableAPIContactQuery bool               `json:"enable_api_contact_query"`
	EnableAPISMSQuery     bool               `json:"enable_api_sms_query"`
	EnableAPISMSSend      bool               `json:"enable_api_sms_send"`
	EnableAPIWOL          bool               `json:"enable_api_wol"`
	ExtraDeviceMark       string             `json:"extra_device_mark"`
	ExtraSIM1             string             `json:"extra_sim1"`
	ExtraSIM2             string             `json:"extra_sim2"`
	SIMInfoList           map[string]SIMInfo `json:"sim_info_list"`
}

// SIMInfo SIM卡信息
type SIMInfo struct {
	CarrierName  string `json:"carrier_name"`
	CountryISO   string `json:"country_iso"`
	ICCID        string `json:"icc_id"`
	Number       string `json:"number"`
	SIMSlotIndex int    `json:"sim_slot_index"`
	Subscription int    `json:"subscription_id"`
}

// SMSMessage 短信信息
type SMSMessage struct {
	Content string `json:"content"`
	Number  string `json:"number"`
	Name    string `json:"name"`
	Type    int    `json:"type"`
	Date    int64  `json:"date"`
	SIMID   int    `json:"sim_id"`
	SubID   int    `json:"sub_id"`
}

// CallRecord 通话记录
type CallRecord struct {
	DateLong int64  `json:"dateLong"`
	Number   string `json:"number"`
	Name     string `json:"name"`
	SIMID    int    `json:"sim_id"`
	Type     int    `json:"type"`
	Duration int    `json:"duration"`
}

// Contact 联系人
type Contact struct {
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
}

// BatteryInfo 电池信息
type BatteryInfo struct {
	Level       string `json:"level"`
	Scale       string `json:"scale"`
	Status      string `json:"status"`
	Health      string `json:"health"`
	Plugged     string `json:"plugged"`
	Voltage     string `json:"voltage,omitempty"`
	Temperature string `json:"temperature,omitempty"`
}

// LocationInfo 位置信息
type LocationInfo struct {
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Provider  string  `json:"provider"`
	Time      string  `json:"time"`
}

// WOLRequest 远程唤醒请求
type WOLRequest struct {
	MAC  string `json:"mac"`
	IP   string `json:"ip,omitempty"`
	Port int    `json:"port,omitempty"`
}

// SMSQueryRequest 短信查询请求
type SMSQueryRequest struct {
	Type     int    `json:"type"`
	PageNum  int    `json:"page_num"`
	PageSize int    `json:"page_size"`
	Keyword  string `json:"keyword,omitempty"`
}

// CallQueryRequest 通话查询请求
type CallQueryRequest struct {
	Type        int    `json:"type,omitempty"`
	PageNum     int    `json:"page_num"`
	PageSize    int    `json:"page_size"`
	PhoneNumber string `json:"phone_number,omitempty"`
}

// ContactQueryRequest 联系人查询请求
type ContactQueryRequest struct {
	PhoneNumber string `json:"phone_number,omitempty"`
	Name        string `json:"name,omitempty"`
}

// ClonePullRequest 克隆拉取请求
type ClonePullRequest struct {
	VersionCode int `json:"version_code"`
}

// ClonePushRequest 克隆推送请求
type ClonePushRequest map[string]interface{}

// generateSign 生成签名
func (c *Client) generateSign(timestamp int64) string {
	if c.secret == "" {
		return ""
	}

	message := fmt.Sprintf("%d\n%s", timestamp, c.secret)
	h := hmac.New(sha256.New, []byte(c.secret))
	h.Write([]byte(message))
	return url.QueryEscape(base64.StdEncoding.EncodeToString(h.Sum(nil)))
}

// doRequest 执行HTTP请求
func (c *Client) doRequest(method, endpoint string, data interface{}) (*Response, error) {
	timestamp := time.Now().UnixMilli()
	sign := c.generateSign(timestamp)

	request := &Request{
		Data:      data,
		Timestamp: timestamp,
		Sign:      sign,
	}

	var body []byte
	var err error

	// 根据加密方式处理数据
	switch c.encryption {
	case "plain":
		body, err = json.Marshal(request)
	case "rsa":
		// TODO: 实现RSA加密
		body, err = json.Marshal(request)
	case "sm4":
		// TODO: 实现SM4加密
		body, err = json.Marshal(request)
	default:
		body, err = json.Marshal(request)
	}

	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}

	url := fmt.Sprintf("%s%s", c.baseURL, endpoint)
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	var response Response
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("unmarshal response failed: %w", err)
	}

	if response.Code != 200 {
		return nil, fmt.Errorf("API error: %d - %s", response.Code, response.Msg)
	}

	return &response, nil
}

// QueryConfig 查询配置
func (c *Client) QueryConfig() (*ConfigResponse, error) {
	resp, err := c.doRequest("POST", "/config/query", map[string]interface{}{})
	if err != nil {
		return nil, err
	}

	var config ConfigResponse
	configData, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// ClonePull 从服务端拉取配置
func (c *Client) ClonePull(versionCode int) (map[string]interface{}, error) {
	req := ClonePullRequest{
		VersionCode: versionCode,
	}

	resp, err := c.doRequest("POST", "/clone/pull", req)
	if err != nil {
		return nil, err
	}

	result, ok := resp.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response data format")
	}

	return result, nil
}

// ClonePush 向服务端推送配置
func (c *Client) ClonePush(config ClonePushRequest) error {
	_, err := c.doRequest("POST", "/clone/push", config)
	return err
}

// SendSMS 发送短信
func (c *Client) SendSMS(simSlot int, phoneNumbers, content string) error {
	req := map[string]interface{}{
		"sim_slot":      simSlot,
		"phone_numbers": phoneNumbers,
		"msg_content":   content,
	}

	_, err := c.doRequest("POST", "/sms/send", req)
	return err
}

// QuerySMS 查询短信
func (c *Client) QuerySMS(smsType, pageNum, pageSize int, keyword string) ([]SMSMessage, error) {
	req := SMSQueryRequest{
		Type:     smsType,
		PageNum:  pageNum,
		PageSize: pageSize,
		Keyword:  keyword,
	}

	resp, err := c.doRequest("POST", "/sms/query", req)
	if err != nil {
		return nil, err
	}

	var messages []SMSMessage
	data, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &messages); err != nil {
		return nil, err
	}

	return messages, nil
}

// QueryCall 查询通话记录
func (c *Client) QueryCall(callType, pageNum, pageSize int, phoneNumber string) ([]CallRecord, error) {
	req := CallQueryRequest{
		Type:        callType,
		PageNum:     pageNum,
		PageSize:    pageSize,
		PhoneNumber: phoneNumber,
	}

	resp, err := c.doRequest("POST", "/call/query", req)
	if err != nil {
		return nil, err
	}

	var records []CallRecord
	data, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &records); err != nil {
		return nil, err
	}

	return records, nil
}

// QueryContact 查询联系人
func (c *Client) QueryContact(phoneNumber, name string) ([]Contact, error) {
	req := ContactQueryRequest{
		PhoneNumber: phoneNumber,
		Name:        name,
	}

	resp, err := c.doRequest("POST", "/contact/query", req)
	if err != nil {
		return nil, err
	}

	var contacts []Contact
	data, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &contacts); err != nil {
		return nil, err
	}

	return contacts, nil
}

// AddContact 添加联系人
func (c *Client) AddContact(phoneNumbers, name string) error {
	req := map[string]interface{}{
		"phone_number": phoneNumbers,
		"name":         name,
	}

	_, err := c.doRequest("POST", "/contact/add", req)
	return err
}

// QueryBattery 查询电池信息
func (c *Client) QueryBattery() (*BatteryInfo, error) {
	resp, err := c.doRequest("POST", "/battery/query", map[string]interface{}{})
	if err != nil {
		return nil, err
	}

	var battery BatteryInfo
	data, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &battery); err != nil {
		return nil, err
	}

	return &battery, nil
}

// QueryLocation 查询位置信息
func (c *Client) QueryLocation() (*LocationInfo, error) {
	resp, err := c.doRequest("POST", "/location/query", map[string]interface{}{})
	if err != nil {
		return nil, err
	}

	var location LocationInfo
	data, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &location); err != nil {
		return nil, err
	}

	return &location, nil
}

// SendWOL 发送远程唤醒
func (c *Client) SendWOL(mac, ip string, port int) error {
	req := WOLRequest{
		MAC:  mac,
		IP:   ip,
		Port: port,
	}

	_, err := c.doRequest("POST", "/wol/send", req)
	return err
}

// Example 使用示例
/*
package main

import (
	"fmt"
	"log"
	"os"

	"your_project/utils/sms"
)

func main() {
	// 设置环境变量
	os.Setenv("SMS_BASE_URL", "http://192.168.1.100:5000")
	os.Setenv("SMS_SECRET", "your-secret-key")

	// 创建客户端
	client := sms.NewClient()

	// 查询配置
	config, err := client.QueryConfig()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("配置: %+v\n", config)

	// 发送短信
	err = client.SendSMS(sms.SIM1, "13800138000", "测试短信")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("短信发送成功")
}
*/
