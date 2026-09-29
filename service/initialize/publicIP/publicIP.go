package publicIP

import (
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"sun-panel/global"
	"time"
)

var publicIPAPIs = []string{
	"https://4.ipw.cn",              // 国内高速纯文本 IPv4 接口 (IPW)
	"https://ip.3322.net",           // 国内经典快速纯文本接口 (PubYun/3322)
	"https://checkip.amazonaws.com", // 亚马逊 AWS 全球节点 (国内直连通常极快)
	"https://api.ipify.org",         // 国际通用接口 (Cloudflare/AWS)
	"https://icanhazip.com",         // 国际通用接口 (Cloudflare)
	"https://ifconfig.me/ip",        // 备用
	"https://api.ip.sb/ip",          // 备用
	"http://api.ipify.org",          // 备用 HTTP 降级
}

var httpClient = &http.Client{
	Timeout: 6 * time.Second,
}

// 获取并更新公网IP
func FetchPublicIP() (string, error) {
	for _, apiURL := range publicIPAPIs {
		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			continue
		}
		req.Header.Set("User-Agent", "curl/7.88.1")

		resp, err := httpClient.Do(req)
		if err != nil {
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			continue
		}

		ipStr := strings.TrimSpace(string(body))
		// 校验是否为合法IP地址
		parsedIP := net.ParseIP(ipStr)
		if parsedIP != nil {
			oldIP := global.GetServerPublicIP()
			if oldIP != ipStr {
				global.SetServerPublicIP(ipStr)
				if global.Logger != nil {
					global.Logger.Infof("[PublicIP] Server public IP updated from '%s' to '%s' (source: %s)", oldIP, ipStr, apiURL)
				}
			}
			return ipStr, nil
		}
	}

	err := errors.New("failed to retrieve public IP from all known sources")
	if global.Logger != nil {
		global.Logger.Warnf("[PublicIP] %v", err)
	}
	return "", err
}

// 初始化公网IP定时更新任务（例如每10分钟）
func InitPublicIP() {
	go func() {
		// 启动时立即获取一次
		ip, err := FetchPublicIP()
		if err != nil || ip == "" {
			// 如果启动时网络未就绪，5秒后重试一次
			time.Sleep(5 * time.Second)
			_, _ = FetchPublicIP()
		}

		// 每10分钟定时更新公网IP
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			_, _ = FetchPublicIP()
		}
	}()
}
