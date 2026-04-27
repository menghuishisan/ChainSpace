package types

import (
	"encoding/json"
	"time"
)

// ComponentType 组件类型
type ComponentType string

const (
	ComponentTool    ComponentType = "tool"    // 工具类：输入→输出
	ComponentDemo    ComponentType = "demo"    // 演示类：展示静态结构
	ComponentProcess ComponentType = "process" // 过程类：模拟动态过程
	ComponentAttack  ComponentType = "attack"  // 攻击类：演示攻击流程
	ComponentDeFi    ComponentType = "defi"    // DeFi类：模拟金融机制
)

// Capability 组件能力
type Capability string

const (
	CapabilityParamPanel   Capability = "param_panel"   // 参数面板
	CapabilityTimeControl  Capability = "time_control"  // 时间控制
	CapabilityStateMonitor Capability = "state_monitor" // 状态监控
	CapabilityEventLog     Capability = "event_log"     // 事件日志
	CapabilitySnapshot     Capability = "snapshot"      // 快照保存
)

// SimulatorStatus 模拟器状态
type SimulatorStatus string

const (
	StatusIdle    SimulatorStatus = "idle"    // 空闲
	StatusRunning SimulatorStatus = "running" // 运行中
	StatusPaused  SimulatorStatus = "paused"  // 已暂停
	StatusStopped SimulatorStatus = "stopped" // 已停止
	StatusError   SimulatorStatus = "error"   // 错误
)

// NodeID 节点标识
type NodeID string

// Config 模拟器配置
type Config struct {
	Module     string                 `json:"module"`      // 模块名称
	Params     map[string]interface{} `json:"params"`      // 参数配置
	Mode       RunMode                `json:"mode"`        // 运行模式
	NodeCount  int                    `json:"node_count"`  // 节点数量
	NetworkCfg *NetworkConfig         `json:"network_cfg"` // 网络配置
}

// RunMode 运行模式
type RunMode string

const (
	ModeSimulated RunMode = "simulated" // 模拟模式：内存队列
	ModeReal      RunMode = "real"      // 真实模式：TCP/gRPC
)

// NetworkConfig 网络配置
type NetworkConfig struct {
	Latency    time.Duration `json:"latency"`     // 基础延迟
	Jitter     time.Duration `json:"jitter"`      // 延迟抖动
	PacketLoss float64       `json:"packet_loss"` // 丢包率 0-1
	Bandwidth  int64         `json:"bandwidth"`   // 带宽限制 bytes/s
	Partitions [][]NodeID    `json:"partitions"`  // 网络分区
}

// State 模拟器状态
type State struct {
	Tick       uint64                 `json:"tick"`        // 当前时刻
	Status     SimulatorStatus        `json:"status"`      // 运行状态
	Nodes      map[NodeID]*NodeState  `json:"nodes"`       // 节点状态
	GlobalData map[string]interface{} `json:"global_data"` // 全局数据
	UpdatedAt  time.Time              `json:"updated_at"`  // 更新时间
}

// NodeState 节点状态
type NodeState struct {
	ID          NodeID                 `json:"id"`
	Status      string                 `json:"status"`
	Data        map[string]interface{} `json:"data"`
	IsByzantine bool                   `json:"is_byzantine"`
}

// Event 事件
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Tick      uint64                 `json:"tick"`
	Timestamp time.Time              `json:"timestamp"`
	Source    NodeID                 `json:"source,omitempty"`
	Target    NodeID                 `json:"target,omitempty"`
	Data      map[string]interface{} `json:"data"`
}

// Param 参数定义
type Param struct {
	Key         string      `json:"key"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Type        ParamType   `json:"type"`
	Default     interface{} `json:"default"`
	Min         interface{} `json:"min,omitempty"`
	Max         interface{} `json:"max,omitempty"`
	Options     []Option    `json:"options,omitempty"`
	Value       interface{} `json:"value"`
}

// ParamType 参数类型
type ParamType string

const (
	ParamTypeInt    ParamType = "int"
	ParamTypeFloat  ParamType = "float"
	ParamTypeString ParamType = "string"
	ParamTypeBool   ParamType = "bool"
	ParamTypeSelect ParamType = "select"
	ParamTypeSlider ParamType = "slider"
)

// Option 选项
type Option struct {
	Label string      `json:"label"`
	Value interface{} `json:"value"`
}

// SnapshotInfo 快照信息
type SnapshotInfo struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Tick      uint64    `json:"tick"`
	CreatedAt time.Time `json:"created_at"`
	Size      int64     `json:"size"`
}

// Snapshot 快照数据
type Snapshot struct {
	Info  SnapshotInfo    `json:"info"`
	State json.RawMessage `json:"state"`
}

// Description 模块描述
type Description struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	Category     string        `json:"category"`
	Type         ComponentType `json:"type"`
	Capabilities []Capability  `json:"capabilities"`
	Params       []Param       `json:"params"`
	Version      string        `json:"version"`
}

// Message 节点间消息
type Message struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	From      NodeID          `json:"from"`
	To        NodeID          `json:"to"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp time.Time       `json:"timestamp"`
	Tick      uint64          `json:"tick"`
}

// Broadcast 广播消息
type Broadcast struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	From      NodeID          `json:"from"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp time.Time       `json:"timestamp"`
	Tick      uint64          `json:"tick"`
}
