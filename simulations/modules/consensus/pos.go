package consensus

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
	"github.com/google/uuid"
)

type PoSValidator struct {
	ID             types.NodeID `json:"id"`
	Stake          uint64       `json:"stake"`
	EffectiveStake uint64       `json:"effective_stake"`
	IsActive       bool         `json:"is_active"`
	IsSlashed      bool         `json:"is_slashed"`
	SlashReason    string       `json:"slash_reason,omitempty"`
	BlocksProposed int          `json:"blocks_proposed"`
	BlocksAttested int          `json:"blocks_attested"`
	Rewards        uint64       `json:"rewards"`
	Penalties      uint64       `json:"penalties"`
	LastActiveSlot uint64       `json:"last_active_slot"`
}

type PoSBlock struct {
	Hash         string           `json:"hash"`
	ParentHash   string           `json:"parent_hash"`
	Slot         uint64           `json:"slot"`
	Proposer     types.NodeID     `json:"proposer"`
	Attestations []PoSAttestation `json:"attestations"`
	Timestamp    time.Time        `json:"timestamp"`
	Finalized    bool             `json:"finalized"`
}

type PoSAttestation struct {
	Validator types.NodeID `json:"validator"`
	Slot      uint64       `json:"slot"`
	BlockHash string       `json:"block_hash"`
	Timestamp time.Time    `json:"timestamp"`
}

type PoSSimulator struct {
	*base.BaseSimulator
	mu              sync.RWMutex
	validators      map[types.NodeID]*PoSValidator
	validatorList   []types.NodeID
	chain           []*PoSBlock
	currentSlot     uint64
	currentEpoch    uint64
	slotsPerEpoch   uint64
	totalStake      uint64
	minStake        uint64
	slashingPenalty float64
	baseReward      uint64
	finalized       uint64
	justified       uint64
}

func NewPoSSimulator() *PoSSimulator {
	sim := &PoSSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"pos",
			"PoS 权益证明",
			"模拟 PoS 下的验证者选择、提议、证明、Slash 惩罚和最终确认过程。",
			"consensus",
			types.ComponentProcess,
		),
		validators:      make(map[types.NodeID]*PoSValidator),
		validatorList:   make([]types.NodeID, 0),
		chain:           make([]*PoSBlock, 0),
		slotsPerEpoch:   32,
		minStake:        32,
		slashingPenalty: 0.5,
		baseReward:      1,
	}

	sim.AddParam(types.Param{
		Key:         "validator_count",
		Name:        "验证者数量",
		Description: "参与共识的验证者数量。",
		Type:        types.ParamTypeInt,
		Default:     10,
		Min:         4,
		Max:         100,
	})
	sim.AddParam(types.Param{
		Key:         "slots_per_epoch",
		Name:        "每 Epoch 槽数",
		Description: "一个 Epoch 中包含的槽位数量。",
		Type:        types.ParamTypeInt,
		Default:     32,
		Min:         4,
		Max:         64,
	})
	sim.AddParam(types.Param{
		Key:         "min_stake",
		Name:        "最小质押",
		Description: "成为验证者所需的最小质押量。",
		Type:        types.ParamTypeInt,
		Default:     32,
		Min:         1,
		Max:         100,
	})

	sim.SetOnTick(sim.onTick)
	return sim
}

func (s *PoSSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	validatorCount := 10
	if v, ok := config.Params["validator_count"]; ok {
		if n, ok := v.(float64); ok {
			validatorCount = int(n)
		}
	}
	if v, ok := config.Params["slots_per_epoch"]; ok {
		if n, ok := v.(float64); ok {
			s.slotsPerEpoch = uint64(n)
		}
	}
	if v, ok := config.Params["min_stake"]; ok {
		if n, ok := v.(float64); ok {
			s.minStake = uint64(n)
		}
	}

	s.validators = make(map[types.NodeID]*PoSValidator)
	s.validatorList = make([]types.NodeID, 0, validatorCount)
	s.chain = make([]*PoSBlock, 0)
	s.totalStake = 0
	s.currentSlot = 0
	s.currentEpoch = 0
	s.finalized = 0
	s.justified = 0

	for i := 0; i < validatorCount; i++ {
		nodeID := types.NodeID(fmt.Sprintf("validator-%d", i))
		stake := s.minStake + uint64(rand.Intn(100))
		validator := &PoSValidator{
			ID:             nodeID,
			Stake:          stake,
			EffectiveStake: stake,
			IsActive:       true,
		}
		s.validators[nodeID] = validator
		s.validatorList = append(s.validatorList, nodeID)
		s.totalStake += stake
	}

	genesis := &PoSBlock{
		Hash:      "genesis",
		Slot:      0,
		Finalized: true,
		Timestamp: time.Now(),
	}
	s.chain = []*PoSBlock{genesis}

	s.updateAllValidatorStates()
	s.updateGlobalState()
	return nil
}

func (s *PoSSimulator) onTick(tick uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.currentSlot++
	if s.currentSlot%s.slotsPerEpoch == 0 {
		s.currentEpoch++
		s.processEpochTransition()
	}

	proposer := s.selectProposer(s.currentSlot)
	if proposer != "" {
		s.proposeBlock(proposer, s.currentSlot)
	}

	s.collectAttestations(s.currentSlot)
	s.updateFinality()
	s.updateAllValidatorStates()
	s.updateGlobalState()
	return nil
}

func (s *PoSSimulator) selectProposer(slot uint64) types.NodeID {
	activeValidators := s.getActiveValidators()
	if len(activeValidators) == 0 {
		return ""
	}

	totalEffective := uint64(0)
	for _, validator := range activeValidators {
		totalEffective += validator.EffectiveStake
	}

	rand.Seed(int64(slot))
	target := rand.Uint64() % totalEffective
	cumulative := uint64(0)
	for _, validator := range activeValidators {
		cumulative += validator.EffectiveStake
		if target < cumulative {
			return validator.ID
		}
	}

	return activeValidators[0].ID
}

func (s *PoSSimulator) proposeBlock(proposerID types.NodeID, slot uint64) {
	validator := s.validators[proposerID]
	if validator == nil || !validator.IsActive {
		return
	}

	parentBlock := s.chain[len(s.chain)-1]
	block := &PoSBlock{
		Hash:         fmt.Sprintf("block-%d-%s", slot, uuid.New().String()[:8]),
		ParentHash:   parentBlock.Hash,
		Slot:         slot,
		Proposer:     proposerID,
		Attestations: make([]PoSAttestation, 0),
		Timestamp:    time.Now(),
	}

	s.chain = append(s.chain, block)
	validator.BlocksProposed++
	validator.Rewards += s.baseReward
	validator.LastActiveSlot = slot

	s.EmitEvent("block_proposed", proposerID, "", map[string]interface{}{
		"slot":       slot,
		"block_hash": block.Hash,
	})
}

func (s *PoSSimulator) collectAttestations(slot uint64) {
	if len(s.chain) < 2 {
		return
	}

	currentBlock := s.chain[len(s.chain)-1]
	committee := s.getCommittee(slot)
	for _, validatorID := range committee {
		validator := s.validators[validatorID]
		if validator == nil || !validator.IsActive {
			continue
		}
		if rand.Float64() < 0.95 {
			attestation := PoSAttestation{
				Validator: validatorID,
				Slot:      slot,
				BlockHash: currentBlock.Hash,
				Timestamp: time.Now(),
			}
			currentBlock.Attestations = append(currentBlock.Attestations, attestation)
			validator.BlocksAttested++
			validator.Rewards += s.baseReward / 4
		}
	}
}

func (s *PoSSimulator) getCommittee(slot uint64) []types.NodeID {
	activeValidators := s.getActiveValidators()
	if len(activeValidators) == 0 {
		return nil
	}

	committeeSize := len(activeValidators) / 4
	if committeeSize < 4 {
		committeeSize = len(activeValidators)
	}

	rand.Seed(int64(slot))
	rand.Shuffle(len(activeValidators), func(i, j int) {
		activeValidators[i], activeValidators[j] = activeValidators[j], activeValidators[i]
	})

	committee := make([]types.NodeID, committeeSize)
	for i := 0; i < committeeSize; i++ {
		committee[i] = activeValidators[i].ID
	}
	return committee
}

func (s *PoSSimulator) getActiveValidators() []*PoSValidator {
	var active []*PoSValidator
	for _, validator := range s.validators {
		if validator.IsActive && !validator.IsSlashed {
			active = append(active, validator)
		}
	}
	sort.Slice(active, func(i, j int) bool {
		return active[i].ID < active[j].ID
	})
	return active
}

func (s *PoSSimulator) processEpochTransition() {
	for _, validator := range s.validators {
		if validator.IsSlashed {
			continue
		}
		validator.EffectiveStake = validator.Stake
		if validator.Stake > 0 {
			validator.Rewards += s.baseReward / 10
		}
	}

	s.EmitEvent("epoch_transition", "", "", map[string]interface{}{
		"epoch":       s.currentEpoch,
		"total_stake": s.totalStake,
	})
}

func (s *PoSSimulator) updateFinality() {
	if len(s.chain) < 3 {
		return
	}

	for i := len(s.chain) - 3; i >= 0; i-- {
		block := s.chain[i]
		if block.Finalized {
			break
		}

		attestationCount := len(block.Attestations)
		activeCount := len(s.getActiveValidators())
		if float64(attestationCount) < float64(activeCount)*2.0/3.0 {
			continue
		}

		block.Finalized = true
		s.finalized = block.Slot
		s.EmitEvent("block_finalized", "", "", map[string]interface{}{
			"slot":       block.Slot,
			"block_hash": block.Hash,
		})
	}
}

func (s *PoSSimulator) SlashValidator(validatorID types.NodeID, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	validator := s.validators[validatorID]
	if validator == nil {
		return fmt.Errorf("validator not found: %s", validatorID)
	}

	penalty := uint64(float64(validator.Stake) * s.slashingPenalty)
	validator.Stake -= penalty
	validator.Penalties += penalty
	validator.IsSlashed = true
	validator.SlashReason = reason
	validator.IsActive = false
	s.totalStake -= penalty

	s.EmitEvent("validator_slashed", validatorID, "", map[string]interface{}{
		"reason":  reason,
		"penalty": penalty,
	})
	return nil
}

func (s *PoSSimulator) updateValidatorState(validatorID types.NodeID) {
	validator := s.validators[validatorID]
	if validator == nil {
		return
	}

	status := "active"
	if validator.IsSlashed {
		status = "slashed"
	} else if !validator.IsActive {
		status = "inactive"
	}

	s.SetNodeState(validatorID, &types.NodeState{
		ID:     validatorID,
		Status: status,
		Data: map[string]interface{}{
			"stake":           validator.Stake,
			"effective_stake": validator.EffectiveStake,
			"blocks_proposed": validator.BlocksProposed,
			"blocks_attested": validator.BlocksAttested,
			"rewards":         validator.Rewards,
			"penalties":       validator.Penalties,
		},
	})
}

func (s *PoSSimulator) updateAllValidatorStates() {
	for validatorID := range s.validators {
		s.updateValidatorState(validatorID)
	}
}

func (s *PoSSimulator) updateGlobalState() {
	proposer := s.selectProposer(s.currentSlot)
	s.SetGlobalData("current_slot", s.currentSlot)
	s.SetGlobalData("current_epoch", s.currentEpoch)
	s.SetGlobalData("chain_length", len(s.chain))
	s.SetGlobalData("chain_height", len(s.chain)-1)
	s.SetGlobalData("total_stake", s.totalStake)
	s.SetGlobalData("finalized_slot", s.finalized)
	s.SetGlobalData("active_validators", len(s.getActiveValidators()))
	s.SetGlobalData("proposer", proposer)
	s.SetGlobalData("current_actor", proposer)
	s.SetGlobalData("committee_size", len(s.getCommittee(s.currentSlot)))
	s.SetGlobalData("result_height", len(s.chain)-1)
	setConsensusTeachingState(
		s.BaseSimulator,
		"pos_committee",
		"当前 PoS 网络正在根据 slot 和 stake 选择提议者并推进最终确认。",
		"继续观察提议者切换、委员会规模和 finalized slot 的推进情况。",
		65,
		map[string]interface{}{
			"slot":           s.currentSlot,
			"epoch":          s.currentEpoch,
			"proposer":       proposer,
			"finalized_slot": s.finalized,
			"result_height":  len(s.chain) - 1,
		},
	)
}

type PoSFactory struct{}

func (f *PoSFactory) Create() engine.Simulator { return NewPoSSimulator() }

func (f *PoSFactory) GetDescription() types.Description {
	return NewPoSSimulator().GetDescription()
}

func NewPoSFactory() *PoSFactory { return &PoSFactory{} }

func (s *PoSSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "advance_epoch":
		s.mu.Lock()
		s.currentSlot = (s.currentEpoch+1)*s.slotsPerEpoch - 1
		s.processEpochTransition()
		s.updateAllValidatorStates()
		s.updateGlobalState()
		nextEpoch := s.currentEpoch + 1
		s.mu.Unlock()

		return consensusActionResult(
			"已推进到下一 Epoch",
			map[string]interface{}{
				"epoch": nextEpoch,
			},
			&types.ActionFeedback{
				Summary:     "PoS 已进入新的 epoch，验证者集合和提议者会随之刷新。",
				NextHint:    "继续观察当前提议者、活跃验证者数量和 finalized slot 的推进情况。",
				EffectScope: "consensus",
				ResultState: map[string]interface{}{"status": "epoch_advanced"},
			},
		), nil
	case "slash_validator":
		target := types.NodeID("validator-0")
		if raw, ok := params["target"].(string); ok && raw != "" {
			target = types.NodeID(raw)
		}
		reason := "double-sign"
		if raw, ok := params["reason"].(string); ok && raw != "" {
			reason = raw
		}

		if err := s.SlashValidator(target, reason); err != nil {
			return nil, err
		}

		s.mu.Lock()
		s.updateAllValidatorStates()
		s.updateGlobalState()
		s.mu.Unlock()

		return consensusActionResult(
			"已执行 Slash 惩罚",
			map[string]interface{}{
				"target": target,
				"reason": reason,
			},
			&types.ActionFeedback{
				Summary:     "目标验证者已被惩罚，权益和后续出块资格都会受到影响。",
				NextHint:    "继续观察委员会规模、当前提议者和最终确认结果是否随之变化。",
				EffectScope: "consensus",
				ResultState: map[string]interface{}{"status": "validator_slashed"},
			},
		), nil
	default:
		return nil, fmt.Errorf("unsupported pos action: %s", action)
	}
}

var _ engine.SimulatorFactory = (*PoSFactory)(nil)
