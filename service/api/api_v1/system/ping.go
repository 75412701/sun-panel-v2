package system

import (
	"net"
	"strings"
	"sun-panel/api/api_v1/common/apiReturn"
	"sun-panel/global"

	"github.com/gin-gonic/gin"
)

type Ping struct {
}

type PingReq struct {
	LanUrl string `form:"lanUrl" json:"lanUrl"`
	Url    string `form:"url" json:"url"`
	WanUrl string `form:"wanUrl" json:"wanUrl"`
}

// 获取请求客户端的真实 IP
func getClientIP(c *gin.Context) string {
	// 1. 优先读取 X-Forwarded-For 头部（针对反向代理）
	xForwardedFor := c.GetHeader("X-Forwarded-For")
	if xForwardedFor != "" {
		ips := strings.Split(xForwardedFor, ",")
		for _, ip := range ips {
			ip = strings.TrimSpace(ip)
			if ip != "" && net.ParseIP(ip) != nil {
				return ip
			}
		}
	}

	// 2. 检查 X-Real-IP
	xRealIP := strings.TrimSpace(c.GetHeader("X-Real-IP"))
	if xRealIP != "" && net.ParseIP(xRealIP) != nil {
		return xRealIP
	}

	// 3. Gin 自带客户端 IP 解析
	clientIP := strings.TrimSpace(c.ClientIP())
	if host, _, err := net.SplitHostPort(clientIP); err == nil {
		return host
	}
	return clientIP
}

// 判断客户端 IP 是否属于内网
func isLanIP(clientIPStr string, serverPublicIP string) bool {
	clientIPStr = strings.TrimSpace(clientIPStr)
	if clientIPStr == "" {
		return false
	}

	serverPublicIP = strings.TrimSpace(serverPublicIP)
	// 如果 Client_IP 等于当前缓存的 Server_Public_IP，则属于 NAT 回环，判定为内网
	if serverPublicIP != "" && clientIPStr == serverPublicIP {
		return true
	}

	ip := net.ParseIP(clientIPStr)
	if ip == nil {
		return false
	}

	// 本地环回地址 (127.0.0.1, ::1 等)
	if ip.IsLoopback() {
		return true
	}

	// 私有 IP 网段 (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, fc00::/7)
	if ip.IsPrivate() {
		return true
	}

	// 链路本地单播/多播 (169.254.0.0/16, fe80::/10 等)
	if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}

	return false
}

// Ping接口 - 用于网络连通性及内外网智能判定
func (p *Ping) Get(c *gin.Context) {
	var req PingReq
	if c.Request.Method == "POST" {
		_ = c.ShouldBind(&req)
	} else {
		_ = c.ShouldBindQuery(&req)
	}
	if req.Url == "" && req.WanUrl != "" {
		req.Url = req.WanUrl
	}

	clientIP := getClientIP(c)
	serverPublicIP := global.GetServerPublicIP()
	isLan := isLanIP(clientIP, serverPublicIP)

	targetUrl := ""
	if isLan {
		if req.LanUrl != "" {
			targetUrl = req.LanUrl
		} else {
			targetUrl = req.Url
		}
	} else {
		if req.Url != "" {
			targetUrl = req.Url
		} else {
			targetUrl = req.LanUrl
		}
	}

	apiReturn.SuccessData(c, gin.H{
		"isLan":          isLan,
		"targetUrl":      targetUrl,
		"clientIp":       clientIP,
		"serverPublicIp": serverPublicIP,
	})
}
