package ssl

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Manager SSL证书管理器
type Manager struct {
	CertDir string
}

// CertInfo 证书信息
type CertInfo struct {
	Type        string    `json:"type"`        // "self-signed" 或 "letsencrypt"
	Domain      string    `json:"domain"`      // 域名
	ExpiryDate  time.Time `json:"expiry_date"` // 过期时间
	DaysLeft    int       `json:"days_left"`   // 剩余天数
	IsValid     bool      `json:"is_valid"`    // 是否有效
	CertPath    string    `json:"cert_path"`   // 证书文件路径
	KeyPath     string    `json:"key_path"`    // 私钥文件路径
}

// NewManager 创建SSL管理器
func NewManager() *Manager {
	return &Manager{
		CertDir: "/etc/server-panel",
	}
}

// GetCertInfo 获取当前证书信息
func (m *Manager) GetCertInfo() (*CertInfo, error) {
	certPath := filepath.Join(m.CertDir, "server.crt")
	keyPath := filepath.Join(m.CertDir, "server.key")

	// 检查证书文件是否存在
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		return &CertInfo{
			Type:    "none",
			IsValid: false,
		}, nil
	}

	// 读取证书文件
	certData, err := ioutil.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("读取证书文件失败: %v", err)
	}

	// 解析证书
	block, _ := pem.Decode(certData)
	if block == nil {
		return nil, fmt.Errorf("无法解析证书文件")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析证书失败: %v", err)
	}

	// 计算剩余天数
	daysLeft := int(time.Until(cert.NotAfter).Hours() / 24)
	isValid := time.Now().Before(cert.NotAfter)

	// 判断证书类型
	certType := "self-signed"
	if cert.Issuer.Organization != nil && len(cert.Issuer.Organization) > 0 {
		for _, org := range cert.Issuer.Organization {
			if strings.Contains(strings.ToLower(org), "let's encrypt") {
				certType = "letsencrypt"
				break
			}
		}
	}

	domain := "localhost"
	if len(cert.DNSNames) > 0 {
		domain = cert.DNSNames[0]
	} else if cert.Subject.CommonName != "" {
		domain = cert.Subject.CommonName
	}

	return &CertInfo{
		Type:       certType,
		Domain:     domain,
		ExpiryDate: cert.NotAfter,
		DaysLeft:   daysLeft,
		IsValid:    isValid,
		CertPath:   certPath,
		KeyPath:    keyPath,
	}, nil
}

// GenerateSelfSigned 生成自签名证书
func (m *Manager) GenerateSelfSigned(domain string) error {
	// 确保证书目录存在
	if err := os.MkdirAll(m.CertDir, 0755); err != nil {
		return fmt.Errorf("创建证书目录失败: %v", err)
	}

	// 生成私钥
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("生成私钥失败: %v", err)
	}

	// 创建证书模板
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"Server Panel"},
			Country:       []string{"CN"},
			Province:      []string{"Beijing"},
			Locality:      []string{"Beijing"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
			CommonName:    domain,
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour), // 1年有效期
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
	}

	// 添加域名和IP
	if domain != "" && domain != "localhost" {
		template.DNSNames = []string{domain, "localhost"}
	} else {
		template.DNSNames = []string{"localhost"}
	}

	// 获取服务器IP
	if serverIP := getServerIP(); serverIP != "" {
		if ip := net.ParseIP(serverIP); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		}
		template.DNSNames = append(template.DNSNames, serverIP)
	}

	// 生成证书
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("生成证书失败: %v", err)
	}

	// 保存证书文件
	certPath := filepath.Join(m.CertDir, "server.crt")
	certOut, err := os.Create(certPath)
	if err != nil {
		return fmt.Errorf("创建证书文件失败: %v", err)
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}); err != nil {
		return fmt.Errorf("写入证书文件失败: %v", err)
	}

	// 保存私钥文件
	keyPath := filepath.Join(m.CertDir, "server.key")
	keyOut, err := os.Create(keyPath)
	if err != nil {
		return fmt.Errorf("创建私钥文件失败: %v", err)
	}
	defer keyOut.Close()

	privateKeyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("序列化私钥失败: %v", err)
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyDER}); err != nil {
		return fmt.Errorf("写入私钥文件失败: %v", err)
	}

	// 设置文件权限
	os.Chmod(certPath, 0644)
	os.Chmod(keyPath, 0600)

	return nil
}

// RequestLetsEncrypt 申请Let's Encrypt证书
func (m *Manager) RequestLetsEncrypt(domain, email string) error {
	// 检查certbot是否安装
	if !m.isCertbotInstalled() {
		if err := m.installCertbot(); err != nil {
			return fmt.Errorf("安装certbot失败: %v", err)
		}
	}

	// 确保证书目录存在
	if err := os.MkdirAll(m.CertDir, 0755); err != nil {
		return fmt.Errorf("创建证书目录失败: %v", err)
	}

	// 使用certbot申请证书
	cmd := exec.Command("certbot", "certonly",
		"--standalone",
		"--non-interactive",
		"--agree-tos",
		"--email", email,
		"-d", domain,
		"--cert-path", filepath.Join(m.CertDir, "server.crt"),
		"--key-path", filepath.Join(m.CertDir, "server.key"),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("申请Let's Encrypt证书失败: %v\n输出: %s", err, string(output))
	}

	return nil
}

// RenewLetsEncrypt 续期Let's Encrypt证书
func (m *Manager) RenewLetsEncrypt() error {
	if !m.isCertbotInstalled() {
		return fmt.Errorf("certbot未安装")
	}

	cmd := exec.Command("certbot", "renew", "--quiet")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("续期证书失败: %v\n输出: %s", err, string(output))
	}

	return nil
}

// isCertbotInstalled 检查certbot是否已安装
func (m *Manager) isCertbotInstalled() bool {
	_, err := exec.LookPath("certbot")
	return err == nil
}

// installCertbot 安装certbot
func (m *Manager) installCertbot() error {
	// 检测系统类型并安装certbot
	if _, err := os.Stat("/etc/debian_version"); err == nil {
		// Debian/Ubuntu系统
		cmd := exec.Command("apt-get", "update")
		cmd.Run()
		cmd = exec.Command("apt-get", "install", "-y", "certbot")
		return cmd.Run()
	} else if _, err := os.Stat("/etc/redhat-release"); err == nil {
		// CentOS/RHEL系统
		cmd := exec.Command("yum", "install", "-y", "certbot")
		return cmd.Run()
	}

	return fmt.Errorf("不支持的系统类型")
}

// getServerIP 获取服务器IP地址
func getServerIP() string {
	cmd := exec.Command("hostname", "-I")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	
	ips := strings.Fields(string(output))
	if len(ips) > 0 {
		return ips[0]
	}
	
	return ""
}

// DeleteCertificate 删除证书文件
func (m *Manager) DeleteCertificate() error {
	certPath := filepath.Join(m.CertDir, "server.crt")
	keyPath := filepath.Join(m.CertDir, "server.key")

	// 删除证书文件
	if err := os.Remove(certPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除证书文件失败: %v", err)
	}

	// 删除私钥文件
	if err := os.Remove(keyPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除私钥文件失败: %v", err)
	}

	return nil
}
