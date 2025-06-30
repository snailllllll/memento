package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

// 定义响应数据结构
type PicResponse struct {
	Status  string `json:"status"`
	Retcode int    `json:"retcode"`
	Data    struct {
		File   string `json:"file"`
		URL    string `json:"url"`
		Base64 string `json:"base64"`
	} `json:"data"`
}

// 保存Base64图片到本地文件
func SaveBase64ToFile(jsonStr, filename string) error {
	// 如果outputPath为空，设置为当前目录下的pics文件夹
	// todo: 存储路径从配置设置
	if filename == "" {
		filename = "image.png"
	}

	outputPath := filepath.Join(".", "pics", filename)

	// 解析JSON响应
	var resp PicResponse
	if err := json.Unmarshal([]byte(jsonStr), &resp); err != nil {
		return fmt.Errorf("解析JSON失败: %v", err)
	}

	// 检查Base64数据是否为空
	if resp.Data.Base64 == "" {
		return fmt.Errorf("Base64数据为空")
	}

	// 解码Base64数据
	decoded, err := base64.StdEncoding.DecodeString(resp.Data.Base64)
	if err != nil {
		return fmt.Errorf("Base64解码失败: %v", err)
	}

	// 确保输出目录存在
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	// 写入文件
	if err := ioutil.WriteFile(outputPath, decoded, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	fmt.Printf("图片已成功保存到: %s\n", outputPath)
	return nil
}

// DownloadImageFromURL downloads an image from the given URL and saves it to a file
func DownloadImageFromURL(url, filename string) error {
	// 替换 https协议到 http
	if url[:5] == "https" {
		url = url[:4] + url[5:]
	}
	fmt.Printf("开始下载图片: %s\n", url)
	// If filename is empty, use a default name
	if filename == "" {
		filename = "downloaded_image.png"
	}

	outputPath := filepath.Join(".", "pics", filename)

	// Make HTTP GET request
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP请求返回非200状态码: %s", resp.Status)
	}

	// Read response body
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应数据失败: %v", err)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	// Write to file
	if err := ioutil.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	fmt.Printf("图片已成功下载并保存到: %s\n", outputPath)
	return nil
}
