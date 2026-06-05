package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ============================================
// PD API 响应结构体
// ============================================

// PDStoreInfo 单个 TiKV store 信息
type PDStoreInfo struct {
	Store struct {
		ID            int64  `json:"id"`
		Address       string `json:"address"`
		StateName     string `json:"state_name"`
		Version       string `json:"version"`
		StatusAddress string `json:"status_address"`
	} `json:"store"`
	Status struct {
		Capacity    string `json:"capacity"`
		Available   string `json:"available"`
		UsedSize    string `json:"used_size"`
		RegionCount int    `json:"region_count"`
		LeaderCount int    `json:"leader_count"`
		Uptime      string `json:"uptime"`
	} `json:"status"`
}

// PDStoresResponse PD /stores API 响应
type PDStoresResponse struct {
	Count  int           `json:"count"`
	Stores []PDStoreInfo `json:"stores"`
}

// PDHealthItem 单个 PD 节点健康状态
type PDHealthItem struct {
	MemberID uint64 `json:"member_id"`
	Name     string `json:"name"`
	Health   bool   `json:"health"`
}

// ============================================
// 返回给前端的聚合响应
// ============================================

// ClusterResponse 集群状态聚合响应
type ClusterResponse struct {
	PDHealth      bool          `json:"pd_health"`
	PDNodeCount   int           `json:"pd_node_count"`
	TiKVStores    []PDStoreInfo `json:"tikv_stores"`
	StoreCount    int           `json:"store_count"`
	UpStoreCount  int           `json:"up_store_count"`
	RegionCount   int           `json:"region_count"`
	LeaderCount   int           `json:"leader_count"`
	TotalCapacity uint64        `json:"total_capacity"`
	TotalUsed     uint64        `json:"total_used"`
	Error         string        `json:"error,omitempty"`
}

// ============================================
// Handler
// ============================================

// GetClusterStatus 获取 TiDB 集群状态（管理员专用）
func GetClusterStatus(c *gin.Context) {
	pdAddr := os.Getenv("PD_ADDR")
	if pdAddr == "" {
		pdAddr = "http://pd:2379"
	}

	client := &http.Client{Timeout: 5 * time.Second}

	var (
		wg     sync.WaitGroup
		mu     sync.Mutex
		stores PDStoresResponse
		health []PDHealthItem
		errs   []string
	)

	// 并发请求 2 个 PD API 端点
	wg.Add(2)

	// 1. 获取 stores（包含容量、region、leader 等信息）
	go func() {
		defer wg.Done()
		resp, err := client.Get(pdAddr + "/pd/api/v1/stores")
		if err != nil {
			mu.Lock()
			errs = append(errs, "stores: "+err.Error())
			mu.Unlock()
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			mu.Lock()
			errs = append(errs, fmt.Sprintf("stores: HTTP %d", resp.StatusCode))
			mu.Unlock()
			return
		}
		var s PDStoresResponse
		if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
			mu.Lock()
			errs = append(errs, "stores parse: "+err.Error())
			mu.Unlock()
			return
		}
		mu.Lock()
		stores = s
		mu.Unlock()
	}()

	// 2. 获取 health
	go func() {
		defer wg.Done()
		resp, err := client.Get(pdAddr + "/pd/api/v1/health")
		if err != nil {
			mu.Lock()
			errs = append(errs, "health: "+err.Error())
			mu.Unlock()
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			mu.Lock()
			errs = append(errs, fmt.Sprintf("health: HTTP %d", resp.StatusCode))
			mu.Unlock()
			return
		}
		var h []PDHealthItem
		if err := json.NewDecoder(resp.Body).Decode(&h); err != nil {
			mu.Lock()
			errs = append(errs, "health parse: "+err.Error())
			mu.Unlock()
			return
		}
		mu.Lock()
		health = h
		mu.Unlock()
	}()

	wg.Wait()

	// 如果所有端点都失败了，返回 503
	if stores.Count == 0 && len(health) == 0 && len(errs) > 0 {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "无法连接 PD 集群: " + errs[0],
		})
		return
	}

	// 聚合数据（从 stores 中计算，不依赖 /stats 端点）
	response := ClusterResponse{
		StoreCount: stores.Count,
	}

	var totalCapBytes, totalUsedBytes uint64

	for _, s := range stores.Stores {
		response.TiKVStores = append(response.TiKVStores, s)
		if s.Store.StateName == "Up" {
			response.UpStoreCount++
		}
		response.RegionCount += s.Status.RegionCount
		response.LeaderCount += s.Status.LeaderCount
		totalCapBytes += parseCapacity(s.Status.Capacity)
		totalUsedBytes += parseCapacity(s.Status.UsedSize)
	}

	response.TotalCapacity = totalCapBytes
	response.TotalUsed = totalUsedBytes

	// PD 健康状态
	response.PDNodeCount = len(health)
	response.PDHealth = len(health) > 0
	for _, h := range health {
		if !h.Health {
			response.PDHealth = false
			break
		}
	}

	// 如果有部分失败，附加 error 信息
	if len(errs) > 0 {
		response.Error = "部分数据获取失败: " + joinStrings(errs, "; ")
	}

	c.JSON(http.StatusOK, response)
}

// parseCapacity 解析容量字符串 (如 "476.4GiB", "290.8MiB") 为字节数
func parseCapacity(str string) uint64 {
	if str == "" {
		return 0
	}
	var value float64
	var unit string
	if _, err := fmt.Sscanf(str, "%f%s", &value, &unit); err != nil {
		return 0
	}
	switch unit {
	case "TiB":
		return uint64(value * 1024 * 1024 * 1024 * 1024)
	case "GiB":
		return uint64(value * 1024 * 1024 * 1024)
	case "MiB":
		return uint64(value * 1024 * 1024)
	case "KiB":
		return uint64(value * 1024)
	default:
		return uint64(value)
	}
}

func joinStrings(ss []string, sep string) string {
	if len(ss) == 0 {
		return ""
	}
	result := ss[0]
	for i := 1; i < len(ss); i++ {
		result += sep + ss[i]
	}
	return result
}
