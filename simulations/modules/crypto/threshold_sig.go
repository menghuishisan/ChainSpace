package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// ThresholdParty 门限签名参与方
type ThresholdParty struct {
	ID         string `json:"id"`          // 参与方ID
	Share      string `json:"share"`       // 密钥份额
	PublicKey  string `json:"public_key"`  // 公钥份额
	PartialSig string `json:"partial_sig"` // 部分签名
	HasSigned  bool   `json:"has_signed"`  // 是否已签名
}

// ThresholdSignature 门限签名
type ThresholdSignature struct {
	Message     string    `json:"message"`      // 消息
	Hash        string    `json:"hash"`         // 消息哈希
	PartialSigs []string  `json:"partial_sigs"` // 部分签名列表
	FinalSig    string    `json:"final_sig"`    // 最终签名
	SignerCount int       `json:"signer_count"` // 签名者数量
	IsComplete  bool      `json:"is_complete"`  // 是否完成
	Timestamp   time.Time `json:"timestamp"`    // 时间戳
}

// ThresholdSigSimulator 门限签名演示器
// 展示(t,n)门限签名方案，需要t个参与方签名才能生成有效签名
type ThresholdSigSimulator struct {
	*base.BaseSimulator
	parties      map[string]*ThresholdParty // 参与方
	partyList    []string                   // 参与方ID列表
	threshold    int                        // 门限值t
	totalParties int                        // 总参与方数n
	groupPubKey  *ecdsa.PublicKey           // 群公钥
	currentSig   *ThresholdSignature        // 当前签名
	history      []*ThresholdSignature      // 签名历史
}

// NewThresholdSigSimulator 创建门限签名演示器
func NewThresholdSigSimulator() *ThresholdSigSimulator {
	sim := &ThresholdSigSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"threshold_sig",
			"门限签名演示器",
			"展示(t,n)门限签名方案，需要t个参与方合作完成签名",
			"crypto",
			types.ComponentTool,
		),
		parties:   make(map[string]*ThresholdParty),
		partyList: make([]string, 0),
		history:   make([]*ThresholdSignature, 0),
	}

	sim.AddParam(types.Param{
		Key:         "threshold",
		Name:        "门限值",
		Description: "需要的最少签名数",
		Type:        types.ParamTypeInt,
		Default:     3,
		Min:         2,
		Max:         10,
	})
	sim.AddParam(types.Param{
		Key:         "total_parties",
		Name:        "参与方总数",
		Description: "密钥份额的持有者总数",
		Type:        types.ParamTypeInt,
		Default:     5,
		Min:         3,
		Max:         20,
	})

	return sim
}

// Init 初始化演示器
func (s *ThresholdSigSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	s.threshold = 3
	s.totalParties = 5

	if v, ok := config.Params["threshold"]; ok {
		if n, ok := v.(float64); ok {
			s.threshold = int(n)
		}
	}
	if v, ok := config.Params["total_parties"]; ok {
		if n, ok := v.(float64); ok {
			s.totalParties = int(n)
		}
	}

	// 确保threshold <= totalParties
	if s.threshold > s.totalParties {
		s.threshold = s.totalParties
	}

	s.generateKeyShares()
	s.updateState()
	return nil
}

// generateKeyShares 生成密钥份额
func (s *ThresholdSigSimulator) generateKeyShares() {
	// 生成群私钥
	groupPrivKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	s.groupPubKey = &groupPrivKey.PublicKey

	s.parties = make(map[string]*ThresholdParty)
	s.partyList = make([]string, 0, s.totalParties)

	// 简化的份额分配
	for i := 0; i < s.totalParties; i++ {
		partyID := fmt.Sprintf("party-%d", i+1)
		// 生成模拟份额
		shareKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		pubBytes := append(shareKey.PublicKey.X.Bytes(), shareKey.PublicKey.Y.Bytes()...)

		party := &ThresholdParty{
			ID:        partyID,
			Share:     hex.EncodeToString(shareKey.D.Bytes())[:32] + "...",
			PublicKey: hex.EncodeToString(pubBytes)[:32] + "...",
			HasSigned: false,
		}
		s.parties[partyID] = party
		s.partyList = append(s.partyList, partyID)
	}

	s.EmitEvent("key_shares_generated", "", "", map[string]interface{}{
		"threshold":     s.threshold,
		"total_parties": s.totalParties,
	})
}

// StartSigning 开始签名流程
func (s *ThresholdSigSimulator) StartSigning(message string) {
	hash := sha256.Sum256([]byte(message))

	s.currentSig = &ThresholdSignature{
		Message:     message,
		Hash:        hex.EncodeToString(hash[:]),
		PartialSigs: make([]string, 0),
		SignerCount: 0,
		IsComplete:  false,
		Timestamp:   time.Now(),
	}

	// 重置所有参与方状态
	for _, party := range s.parties {
		party.HasSigned = false
		party.PartialSig = ""
	}

	s.EmitEvent("signing_started", "", "", map[string]interface{}{
		"message":   message,
		"hash":      s.currentSig.Hash[:16] + "...",
		"threshold": s.threshold,
	})

	s.updateState()
}

// PartialSign 参与方生成部分签名
func (s *ThresholdSigSimulator) PartialSign(partyID string) error {
	if s.currentSig == nil {
		return fmt.Errorf("没有进行中的签名")
	}

	party := s.parties[partyID]
	if party == nil {
		return fmt.Errorf("参与方不存在: %s", partyID)
	}

	if party.HasSigned {
		return fmt.Errorf("参与方已签名: %s", partyID)
	}

	// 生成模拟的部分签名
	sigData := fmt.Sprintf("%s-%s-%d", s.currentSig.Hash, partyID, time.Now().UnixNano())
	sigHash := sha256.Sum256([]byte(sigData))
	partialSig := hex.EncodeToString(sigHash[:])

	party.PartialSig = partialSig
	party.HasSigned = true
	s.currentSig.PartialSigs = append(s.currentSig.PartialSigs, partialSig)
	s.currentSig.SignerCount++

	s.EmitEvent("partial_signed", types.NodeID(partyID), "", map[string]interface{}{
		"signer_count": s.currentSig.SignerCount,
		"threshold":    s.threshold,
		"partial_sig":  partialSig[:16] + "...",
	})

	// 检查是否达到门限
	if s.currentSig.SignerCount >= s.threshold {
		s.combineSigs()
	}

	s.updateState()
	return nil
}

// combineSigs 合并部分签名
func (s *ThresholdSigSimulator) combineSigs() {
	if s.currentSig == nil || s.currentSig.SignerCount < s.threshold {
		return
	}

	// 模拟签名聚合
	combined := ""
	for _, sig := range s.currentSig.PartialSigs {
		combined += sig
	}
	finalHash := sha256.Sum256([]byte(combined))
	s.currentSig.FinalSig = hex.EncodeToString(finalHash[:])
	s.currentSig.IsComplete = true

	s.history = append(s.history, s.currentSig)

	s.EmitEvent("signature_complete", "", "", map[string]interface{}{
		"final_sig":    s.currentSig.FinalSig[:16] + "...",
		"signer_count": s.currentSig.SignerCount,
	})
}

// VerifyThresholdSig 验证门限签名
func (s *ThresholdSigSimulator) VerifyThresholdSig(sig *ThresholdSignature) bool {
	if sig == nil || !sig.IsComplete {
		return false
	}

	// 模拟验证
	valid := len(sig.PartialSigs) >= s.threshold

	s.EmitEvent("threshold_sig_verified", "", "", map[string]interface{}{
		"valid":        valid,
		"signer_count": sig.SignerCount,
	})

	return valid
}

// updateState 更新状态
func (s *ThresholdSigSimulator) updateState() {
	s.SetGlobalData("threshold", s.threshold)
	s.SetGlobalData("total_parties", s.totalParties)
	s.SetGlobalData("history_count", len(s.history))

	partyList := make([]map[string]interface{}, 0)
	for _, p := range s.parties {
		partyList = append(partyList, map[string]interface{}{
			"id":         p.ID,
			"has_signed": p.HasSigned,
		})
	}
	s.SetGlobalData("parties", partyList)

	if s.currentSig != nil {
		s.SetGlobalData("current_signing", map[string]interface{}{
			"message":      s.currentSig.Message,
			"signer_count": s.currentSig.SignerCount,
			"is_complete":  s.currentSig.IsComplete,
		})
	}

	summary := fmt.Sprintf("当前门限为 %d/%d，已记录 %d 条门限签名历史。", s.threshold, s.totalParties, len(s.history))
	nextHint := "可以先启动一次签名，再让参与方逐个提交部分签名，观察何时达到门限。"
	setCryptoTeachingState(
		s.BaseSimulator,
		"crypto",
		"准备门限签名",
		summary,
		nextHint,
		0.5,
		map[string]interface{}{"threshold": s.threshold, "total_parties": s.totalParties, "history_count": len(s.history)},
	)
}

func (s *ThresholdSigSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "start_signing":
		message := "ChainSpace"
		if raw, ok := params["message"].(string); ok && raw != "" {
			message = raw
		}
		s.StartSigning(message)
		return cryptoActionResult("已启动一次门限签名流程。", map[string]interface{}{"message": message}, &types.ActionFeedback{
			Summary:     "新的门限签名流程已经开始，等待参与方提交部分签名。",
			NextHint:    "继续让不同参与方执行部分签名，观察何时达到门限并合并为最终签名。",
			EffectScope: "crypto",
			ResultState: map[string]interface{}{"message": message},
		}), nil
	case "partial_sign":
		partyID := ""
		for id, party := range s.parties {
			if !party.HasSigned {
				partyID = id
				break
			}
		}
		if raw, ok := params["party_id"].(string); ok && raw != "" {
			partyID = raw
		}
		if partyID == "" {
			return nil, fmt.Errorf("no party available for partial signing")
		}
		if err := s.PartialSign(partyID); err != nil {
			return nil, err
		}
		return cryptoActionResult("已提交一份部分签名。", map[string]interface{}{"party_id": partyID}, &types.ActionFeedback{
			Summary:     "新的部分签名已加入当前门限签名流程。",
			NextHint:    "继续增加签名者，直到达到门限并观察最终签名是否完成。",
			EffectScope: "crypto",
			ResultState: map[string]interface{}{"party_id": partyID, "threshold": s.threshold},
		}), nil
	default:
		return nil, fmt.Errorf("unsupported threshold signature action: %s", action)
	}
}

// ThresholdSigFactory 门限签名演示器工厂
type ThresholdSigFactory struct{}

func (f *ThresholdSigFactory) Create() engine.Simulator {
	return NewThresholdSigSimulator()
}

func (f *ThresholdSigFactory) GetDescription() types.Description {
	return NewThresholdSigSimulator().GetDescription()
}

func NewThresholdSigFactory() *ThresholdSigFactory {
	return &ThresholdSigFactory{}
}

var _ engine.SimulatorFactory = (*ThresholdSigFactory)(nil)
