package middleware

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// 端口状态常量
const (
	PortStatusOpen    = "open"
	PortStatusClosed  = "closed"
	PortStatusTimeout = "timeout"
	PortStatusError   = "error"
)

// 端口扫描结果
type PortScanResult struct {
	Port   int    `json:"port"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// 端口扫描请求
type PortScanRequest struct {
	Host      string `json:"host" binding:"required"`
	Ports     string `json:"ports"`
	ScanAll   bool   `json:"scanAll"`
	Timeout   int    `json:"timeout"`
	BatchSize int    `json:"batchSize"`
}

// 端口扫描响应
type PortScanResponse struct {
	Host         string          `json:"host"`
	TotalScanned int             `json:"totalScanned"`
	OpenPorts    []int           `json:"openPorts"`
	ClosedPorts  []int           `json:"closedPorts"`
	TimeoutPorts []int           `json:"timeoutPorts"`
	ErrorPorts   []int           `json:"errorPorts"`
	Results      []PortScanResult `json:"results"`
	Duration     string          `json:"duration"`
	StartTime    string          `json:"startTime"`
	EndTime      string          `json:"endTime"`
}

// 检测单个端口
func checkPort(host string, port int, timeout time.Duration) PortScanResult {
	result := PortScanResult{
		Port:   port,
		Status: PortStatusClosed,
	}

	address := fmt.Sprintf("%s:%d", host, port)
	fmt.Printf("[PORT-SCAN] 正在检测端口: %s\n", address)
	
	conn, err := net.DialTimeout("tcp", address, timeout)

	if err != nil {
		fmt.Printf("[PORT-SCAN] 端口 %d 连接失败: %v\n", port, err)
		if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "i/o timeout") {
			result.Status = PortStatusTimeout
			result.Error = "连接超时"
		} else if strings.Contains(err.Error(), "refused") || strings.Contains(err.Error(), "connection refused") {
			result.Status = PortStatusClosed
		} else if strings.Contains(err.Error(), "unreachable") || strings.Contains(err.Error(), "no route") {
			result.Status = PortStatusError
			result.Error = "网络不可达"
		} else {
			result.Status = PortStatusError
			result.Error = err.Error()
		}
		return result
	}

	defer conn.Close()
	
	// 尝试写入一个字节来验证连接是否真的可用
	conn.SetWriteDeadline(time.Now().Add(1 * time.Second))
	_, writeErr := conn.Write([]byte{0x00})
	
	if writeErr != nil {
		fmt.Printf("[PORT-SCAN] 端口 %d 连接建立但写入失败: %v\n", port, writeErr)
		// 即使写入失败，如果能建立连接，通常还是认为端口开放
	} else {
		fmt.Printf("[PORT-SCAN] 端口 %d 连接并写入成功！\n", port)
	}
	
	result.Status = PortStatusOpen
	return result
}

// 解析端口列表
func parsePorts(portsStr string) ([]int, error) {
	if portsStr == "" {
		return []int{}, nil
	}

	var ports []int
	parts := strings.Split(portsStr, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// 检查是否是范围 (例如 "1000-2000")
		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("无效的端口范围: %s", part)
			}

			start, err := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
			if err != nil {
				return nil, fmt.Errorf("无效的起始端口: %s", rangeParts[0])
			}

			end, err := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
			if err != nil {
				return nil, fmt.Errorf("无效的结束端口: %s", rangeParts[1])
			}

			if start < 1 || start > 65535 || end < 1 || end > 65535 || start > end {
				return nil, fmt.Errorf("端口范围无效: %d-%d", start, end)
			}

			for i := start; i <= end; i++ {
				ports = append(ports, i)
			}
		} else {
			// 单个端口
			port, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("无效的端口号: %s", part)
			}

			if port < 1 || port > 65535 {
				return nil, fmt.Errorf("端口号超出范围 (1-65535): %d", port)
			}

			ports = append(ports, port)
		}
	}

	return ports, nil
}

// 批量扫描端口
func batchScanPorts(host string, ports []int, timeout time.Duration, batchSize int) []PortScanResult {
	if batchSize <= 0 {
		batchSize = 100 // 默认批次大小
	}

	var results []PortScanResult
	var mutex sync.Mutex
	var wg sync.WaitGroup

	// 分批处理
	for i := 0; i < len(ports); i += batchSize {
		end := i + batchSize
		if end > len(ports) {
			end = len(ports)
		}

		batch := ports[i:end]
		wg.Add(1)

		go func(batchPorts []int) {
			defer wg.Done()
			batchResults := make([]PortScanResult, len(batchPorts))

			var batchWg sync.WaitGroup
			for j, port := range batchPorts {
				batchWg.Add(1)
				go func(index int, portNum int) {
					defer batchWg.Done()
					batchResults[index] = checkPort(host, portNum, timeout)
				}(j, port)
			}
			batchWg.Wait()

			// 合并结果
			mutex.Lock()
			results = append(results, batchResults...)
			mutex.Unlock()
		}(batch)
	}

	wg.Wait()
	return results
}

// 获取所有端口
func getAllPorts() []int {
	ports := make([]int, 65535)
	for i := 0; i < 65535; i++ {
		ports[i] = i + 1
	}
	return ports
}

// 处理端口扫描请求
func HandlePortScan(c *gin.Context) {
	var req PortScanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"error": "无效的请求参数: " + err.Error(),
		})
		return
	}

	// 设置默认值
	if req.Timeout <= 0 {
		req.Timeout = 3000 // 默认3秒超时
	}
	if req.BatchSize <= 0 {
		req.BatchSize = 100 // 默认批次大小
	}

	// 记录开始时间
	startTime := time.Now()
	startTimeStr := startTime.Format("2006-01-02 15:04:05")

	fmt.Printf("[PORT-SCAN] [%d] 开始扫描主机: %s\n", startTime.UnixNano(), req.Host)

	var ports []int
	var err error

	// 根据请求类型确定要扫描的端口
	if req.ScanAll {
		fmt.Printf("[PORT-SCAN] [%d] 扫描所有端口 (1-65535)\n", startTime.UnixNano())
		ports = getAllPorts()
	} else {
		fmt.Printf("[PORT-SCAN] [%d] 扫描指定端口: %s\n", startTime.UnixNano(), req.Ports)
		ports, err = parsePorts(req.Ports)
		if err != nil {
			c.JSON(400, gin.H{
				"error": "端口解析错误: " + err.Error(),
			})
			return
		}
	}

	if len(ports) == 0 {
		c.JSON(400, gin.H{
			"error": "未指定要扫描的端口",
		})
		return
	}

	fmt.Printf("[PORT-SCAN] [%d] 共需扫描 %d 个端口，超时设置: %dms，批次大小: %d\n", 
		startTime.UnixNano(), len(ports), req.Timeout, req.BatchSize)

	// 执行端口扫描
	timeout := time.Duration(req.Timeout) * time.Millisecond
	results := batchScanPorts(req.Host, ports, timeout, req.BatchSize)

	// 处理结果
	endTime := time.Now()
	duration := endTime.Sub(startTime)
	endTimeStr := endTime.Format("2006-01-02 15:04:05")

	// 统计结果
	openPorts := make([]int, 0)
	closedPorts := make([]int, 0)
	timeoutPorts := make([]int, 0)
	errorPorts := make([]int, 0)
	for _, result := range results {
		switch result.Status {
		case PortStatusOpen:
			openPorts = append(openPorts, result.Port)
		case PortStatusClosed:
			closedPorts = append(closedPorts, result.Port)
		case PortStatusTimeout:
			timeoutPorts = append(timeoutPorts, result.Port)
		case PortStatusError:
			errorPorts = append(errorPorts, result.Port)
		}
	}

	fmt.Printf("[PORT-SCAN] [%d] 扫描完成，耗时: %s\n", startTime.UnixNano(), duration)
	fmt.Printf("[PORT-SCAN] [%d] 开放端口: %d 个, 关闭端口: %d 个, 超时端口: %d 个, 错误端口: %d 个\n", 
		startTime.UnixNano(), len(openPorts), len(closedPorts), len(timeoutPorts), len(errorPorts))

	response := PortScanResponse{
		Host:         req.Host,
		TotalScanned: len(ports),
		OpenPorts:    openPorts,
		ClosedPorts:  closedPorts,
		TimeoutPorts: timeoutPorts,
		ErrorPorts:   errorPorts,
		Results:      results,
		Duration:     duration.String(),
		StartTime:    startTimeStr,
		EndTime:      endTimeStr,
	}

	c.JSON(200, response)
}

// 注册端口扫描路由
func RegisterPortScanRoutes(r *gin.Engine) {
	r.POST("/port-scan", HandlePortScan)
}
