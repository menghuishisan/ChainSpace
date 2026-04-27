package api

import (
	"fmt"
	"strconv"

	"github.com/chainspace/simulations/pkg/types"
	"github.com/gin-gonic/gin"
)

// handleGetMeta 获取当前模拟器元信息
func (s *Server) handleGetMeta(c *gin.Context) {
	desc := s.engine.GetDescription()
	if desc == nil {
		fail(c, 10001, "no active simulator")
		return
	}
	success(c, desc)
}

// handleListSimulators 列出所有可用模拟器
func (s *Server) handleListSimulators(c *gin.Context) {
	simulators := s.engine.ListSimulators()
	success(c, simulators)
}

// InitRequest 初始化请求
type InitRequest struct {
	Module    string                 `json:"module" binding:"required"`
	Params    map[string]interface{} `json:"params"`
	Mode      string                 `json:"mode"`
	NodeCount int                    `json:"node_count"`
}

// handleInit 初始化模拟器
func (s *Server) handleInit(c *gin.Context) {
	var req InitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 10001, err.Error())
		return
	}

	config := types.Config{
		Module:    req.Module,
		Params:    req.Params,
		Mode:      types.RunMode(req.Mode),
		NodeCount: req.NodeCount,
	}

	if err := s.engine.Init(config); err != nil {
		fail(c, 10002, err.Error())
		return
	}

	success(c, gin.H{"status": "initialized"})
}

// handleStart 启动模拟
func (s *Server) handleStart(c *gin.Context) {
	if err := s.engine.Start(c.Request.Context()); err != nil {
		fail(c, 10003, err.Error())
		return
	}
	success(c, gin.H{"status": "started"})
}

// handleStop 停止模拟
func (s *Server) handleStop(c *gin.Context) {
	if err := s.engine.Stop(); err != nil {
		fail(c, 10004, err.Error())
		return
	}
	success(c, gin.H{"status": "stopped"})
}

// handlePause 暂停模拟
func (s *Server) handlePause(c *gin.Context) {
	if err := s.engine.Pause(); err != nil {
		fail(c, 10005, err.Error())
		return
	}
	success(c, gin.H{"status": "paused"})
}

// handleResume 恢复模拟
func (s *Server) handleResume(c *gin.Context) {
	if err := s.engine.Resume(); err != nil {
		fail(c, 10006, err.Error())
		return
	}
	success(c, gin.H{"status": "resumed"})
}

// handleStep 单步执行
func (s *Server) handleStep(c *gin.Context) {
	state, err := s.engine.Step()
	if err != nil {
		fail(c, 10007, err.Error())
		return
	}
	success(c, state)
}

// handleReset 重置模拟
func (s *Server) handleReset(c *gin.Context) {
	if err := s.engine.Reset(); err != nil {
		fail(c, 10008, err.Error())
		return
	}
	success(c, gin.H{"status": "reset"})
}

// SwitchRequest 切换请求
type SwitchRequest struct {
	Module        string `json:"module" binding:"required"`
	PreserveState bool   `json:"preserve_state"`
}

// handleSwitch 切换模拟器
func (s *Server) handleSwitch(c *gin.Context) {
	var req SwitchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 10001, err.Error())
		return
	}

	if err := s.engine.Switch(req.Module, req.PreserveState); err != nil {
		fail(c, 10009, err.Error())
		return
	}

	success(c, gin.H{"status": "switched", "module": req.Module})
}

// handleGetState 获取状态
func (s *Server) handleGetState(c *gin.Context) {
	state := s.engine.GetState()
	if state == nil {
		fail(c, 10010, "no state available")
		return
	}
	success(c, state)
}

// handleGetEvents 获取事件
func (s *Server) handleGetEvents(c *gin.Context) {
	var since uint64
	if sinceStr := c.Query("since"); sinceStr != "" {
		if parsed, err := strconv.ParseUint(sinceStr, 10, 64); err == nil {
			since = parsed
		}
	}

	limit := 100
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	events := s.engine.GetEvents(since)
	if len(events) > limit {
		events = events[len(events)-limit:]
	}
	success(c, gin.H{
		"events": events,
		"count":  len(events),
		"since":  since,
	})
}

// handleGetParams 获取参数
func (s *Server) handleGetParams(c *gin.Context) {
	params := s.engine.GetParams()
	success(c, params)
}

// SetParamRequest 设置参数请求
type SetParamRequest struct {
	Value interface{} `json:"value" binding:"required"`
}

// handleSetParam 设置参数
func (s *Server) handleSetParam(c *gin.Context) {
	key := c.Param("key")
	var req SetParamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 10001, err.Error())
		return
	}

	if err := s.engine.SetParam(key, req.Value); err != nil {
		fail(c, 10011, err.Error())
		return
	}

	success(c, gin.H{"key": key, "value": req.Value})
}

// SetSpeedRequest 设置速度请求
type SetSpeedRequest struct {
	Speed float64 `json:"speed" binding:"required"`
}

// ActionRequest 模块动作请求。
// 用于触发模块自身暴露的教学动作，例如发起攻击、模拟交易或触发视图切换。
type ActionRequest struct {
	Action string                 `json:"action" binding:"required"`
	Params map[string]interface{} `json:"params"`
}

// handleSetSpeed 设置速度
func (s *Server) handleSetSpeed(c *gin.Context) {
	var req SetSpeedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 10001, err.Error())
		return
	}

	if err := s.engine.SetSpeed(req.Speed); err != nil {
		fail(c, 10012, err.Error())
		return
	}

	success(c, gin.H{"speed": req.Speed})
}

// handleExecuteAction 执行模块动作。
func (s *Server) handleExecuteAction(c *gin.Context) {
	var req ActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 10001, err.Error())
		return
	}

	result, err := s.engine.ExecuteAction(req.Action, req.Params)
	if err != nil {
		fail(c, 10012, err.Error())
		return
	}

	success(c, result)
}

// InjectFaultRequest 注入故障请求
type InjectFaultRequest struct {
	Type     string                 `json:"type" binding:"required"`
	Target   string                 `json:"target" binding:"required"`
	Params   map[string]interface{} `json:"params"`
	Duration uint64                 `json:"duration"`
}

// handleInjectFault 注入故障
func (s *Server) handleInjectFault(c *gin.Context) {
	var req InjectFaultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 10001, err.Error())
		return
	}

	fault := &types.Fault{
		Type:     types.FaultType(req.Type),
		Target:   types.NodeID(req.Target),
		Params:   req.Params,
		Duration: req.Duration,
		Active:   true,
	}

	if err := s.engine.InjectFault(fault); err != nil {
		fail(c, 10013, err.Error())
		return
	}

	success(c, gin.H{"fault_id": fault.ID})
}

// handleRemoveFault 移除故障
func (s *Server) handleRemoveFault(c *gin.Context) {
	faultID := c.Param("id")
	if err := s.engine.RemoveFault(faultID); err != nil {
		fail(c, 10014, err.Error())
		return
	}
	success(c, gin.H{"status": "removed"})
}

// handleClearFaults 清除所有故障
func (s *Server) handleClearFaults(c *gin.Context) {
	if err := s.engine.ClearFaults(); err != nil {
		fail(c, 10015, err.Error())
		return
	}
	success(c, gin.H{"status": "cleared"})
}

// InjectAttackRequest 注入攻击请求
type InjectAttackRequest struct {
	Type     string                 `json:"type" binding:"required"`
	Target   string                 `json:"target" binding:"required"`
	Params   map[string]interface{} `json:"params"`
	Duration uint64                 `json:"duration"`
}

// handleInjectAttack 注入攻击
func (s *Server) handleInjectAttack(c *gin.Context) {
	var req InjectAttackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 10001, err.Error())
		return
	}

	attack := &types.Attack{
		Type:     types.AttackType(req.Type),
		Target:   req.Target,
		Params:   req.Params,
		Duration: req.Duration,
		Active:   true,
	}

	if err := s.engine.InjectAttack(attack); err != nil {
		fail(c, 10016, err.Error())
		return
	}

	success(c, gin.H{"attack_id": attack.ID})
}

// handleRemoveAttack 移除攻击
func (s *Server) handleRemoveAttack(c *gin.Context) {
	attackID := c.Param("id")
	if err := s.engine.RemoveAttack(attackID); err != nil {
		fail(c, 10017, err.Error())
		return
	}
	success(c, gin.H{"status": "removed"})
}

// handleClearAttacks 清除所有攻击
func (s *Server) handleClearAttacks(c *gin.Context) {
	if err := s.engine.ClearAttacks(); err != nil {
		fail(c, 10018, err.Error())
		return
	}
	success(c, gin.H{"status": "cleared"})
}

// SaveSnapshotRequest 保存快照请求
type SaveSnapshotRequest struct {
	Name string `json:"name" binding:"required"`
}

// handleSaveSnapshot 保存快照
func (s *Server) handleSaveSnapshot(c *gin.Context) {
	var req SaveSnapshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 10001, err.Error())
		return
	}

	if err := s.engine.SaveSnapshot(req.Name); err != nil {
		fail(c, 10019, err.Error())
		return
	}

	success(c, gin.H{"status": "saved", "name": req.Name})
}

// handleListSnapshots 列出快照
func (s *Server) handleListSnapshots(c *gin.Context) {
	snapshots := s.engine.ListSnapshots()
	success(c, snapshots)
}

// LoadSnapshotRequest 加载快照请求
type LoadSnapshotRequest struct {
	Name string `json:"name" binding:"required"`
}

// handleLoadSnapshot 加载快照
func (s *Server) handleLoadSnapshot(c *gin.Context) {
	var req LoadSnapshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 10001, err.Error())
		return
	}

	if err := s.engine.LoadSnapshot(req.Name); err != nil {
		fail(c, 10020, err.Error())
		return
	}

	success(c, gin.H{"status": "loaded", "name": req.Name})
}

// handleDeleteSnapshot 删除快照
func (s *Server) handleDeleteSnapshot(c *gin.Context) {
	snapshotID := c.Param("id")
	if snapshotID == "" {
		fail(c, 10021, "snapshot id is required")
		return
	}

	if err := s.engine.DeleteSnapshot(snapshotID); err != nil {
		fail(c, 10022, fmt.Sprintf("failed to delete snapshot: %v", err))
		return
	}

	success(c, gin.H{"status": "deleted", "id": snapshotID})
}

// handleExportState 导出状态
func (s *Server) handleExportState(c *gin.Context) {
	data, err := s.engine.ExportState()
	if err != nil {
		fail(c, 10023, fmt.Sprintf("failed to export state: %v", err))
		return
	}
	success(c, gin.H{"state": data})
}

// ImportStateRequest 导入状态请求
type ImportStateRequest struct {
	State map[string]interface{} `json:"state" binding:"required"`
}

// handleImportState 导入状态
func (s *Server) handleImportState(c *gin.Context) {
	var req ImportStateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 10001, err.Error())
		return
	}

	if err := s.engine.ImportState(req.State); err != nil {
		fail(c, 10024, fmt.Sprintf("failed to import state: %v", err))
		return
	}

	success(c, gin.H{"status": "imported"})
}
