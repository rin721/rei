package utils

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

const (
	customEpochMs = int64(1704067200000)
	nodeBits      = int64(10)
	sequenceBits  = int64(12)
	maxNodeID     = -1 ^ (-1 << nodeBits)
	maxSequence   = -1 ^ (-1 << sequenceBits)
)

// IDGenerator 提供最小 Snowflake ID 生成能力。
type IDGenerator struct {
	mu            sync.Mutex
	nodeID        int64
	sequence      int64
	lastTimestamp int64
}

// NewIDGenerator 创建一个新的 Snowflake 生成器。
func NewIDGenerator(nodeID int64) (*IDGenerator, error) {
	if nodeID < 0 || nodeID > maxNodeID {
		return nil, fmt.Errorf("node id must be between 0 and %d", maxNodeID)
	}

	return &IDGenerator{nodeID: nodeID}, nil
}

// NextID 生成一个新的全局唯一整数 ID。
func (g *IDGenerator) NextID() (int64, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := currentTimestamp()
	if now < g.lastTimestamp {
		return 0, errors.New("clock moved backwards")
	}

	if now == g.lastTimestamp {
		g.sequence = (g.sequence + 1) & maxSequence
		if g.sequence == 0 {
			for now <= g.lastTimestamp {
				now = currentTimestamp()
			}
		}
	} else {
		g.sequence = 0
	}

	g.lastTimestamp = now

	return ((now - customEpochMs) << (nodeBits + sequenceBits)) | (g.nodeID << sequenceBits) | g.sequence, nil
}

// GetFreePort 返回一个当前可用的本地端口。
func GetFreePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	tcpAddr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		return 0, errors.New("listener address is not tcp")
	}

	return tcpAddr.Port, nil
}

// LocalIP 返回第一个非回环 IPv4 地址。
func LocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok || ipNet.IP.IsLoopback() {
			continue
		}
		if ipv4 := ipNet.IP.To4(); ipv4 != nil {
			return ipv4.String(), nil
		}
	}

	return "127.0.0.1", nil
}

func currentTimestamp() int64 {
	return time.Now().UnixMilli()
}
