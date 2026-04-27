package crosschain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

type BridgeType string

const (
	BridgeLockMint   BridgeType = "lock_mint"
	BridgeBurnUnlock BridgeType = "burn_unlock"
	BridgeLiquidity  BridgeType = "liquidity"
	BridgeNative     BridgeType = "native"
)

type BridgeTxStatus string

const (
	BridgeTxPending    BridgeTxStatus = "pending"
	BridgeTxConfirmed  BridgeTxStatus = "confirmed"
	BridgeTxSigned     BridgeTxStatus = "signed"
	BridgeTxExecuted   BridgeTxStatus = "executed"
	BridgeTxCompleted  BridgeTxStatus = "completed"
	BridgeTxFailed     BridgeTxStatus = "failed"
	BridgeTxChallenged BridgeTxStatus = "challenged"
)

type BridgeTransaction struct {
	ID              string         `json:"id"`
	User            string         `json:"user"`
	SourceChain     string         `json:"source_chain"`
	DestChain       string         `json:"dest_chain"`
	Token           string         `json:"token"`
	WrappedToken    string         `json:"wrapped_token"`
	Amount          *big.Int       `json:"amount"`
	Fee             *big.Int       `json:"fee"`
	Status          BridgeTxStatus `json:"status"`
	SourceTxHash    string         `json:"source_tx_hash"`
	DestTxHash      string         `json:"dest_tx_hash"`
	Confirmations   int            `json:"confirmations"`
	RequiredConfirm int            `json:"required_confirm"`
	Signatures      []string       `json:"signatures"`
	RequiredSigs    int            `json:"required_sigs"`
	MerkleProof     []string       `json:"merkle_proof"`
	ChallengeEnd    time.Time      `json:"challenge_end"`
	CreatedAt       time.Time      `json:"created_at"`
	ConfirmedAt     time.Time      `json:"confirmed_at"`
	CompletedAt     time.Time      `json:"completed_at"`
}

type BridgePool struct {
	ID              string   `json:"id"`
	Token           string   `json:"token"`
	SourceChain     string   `json:"source_chain"`
	DestChain       string   `json:"dest_chain"`
	SourceLocked    *big.Int `json:"source_locked"`
	DestMinted      *big.Int `json:"dest_minted"`
	SourceLiquidity *big.Int `json:"source_liquidity"`
	DestLiquidity   *big.Int `json:"dest_liquidity"`
	FeePercent      float64  `json:"fee_percent"`
	TotalVolume     *big.Int `json:"total_volume"`
	TxCount         int      `json:"tx_count"`
}

type BridgeValidator struct {
	Address      string    `json:"address"`
	PublicKey    string    `json:"public_key"`
	Stake        *big.Int  `json:"stake"`
	VotingPower  float64   `json:"voting_power"`
	IsActive     bool      `json:"is_active"`
	SignedCount  int       `json:"signed_count"`
	MissedCount  int       `json:"missed_count"`
	SlashedAmt   *big.Int  `json:"slashed_amount"`
	JoinedAt     time.Time `json:"joined_at"`
	LastActiveAt time.Time `json:"last_active_at"`
}

type ChainConfig struct {
	ChainID        string        `json:"chain_id"`
	ChainName      string        `json:"chain_name"`
	ChainType      string        `json:"chain_type"`
	BlockTime      time.Duration `json:"block_time"`
	Finality       int           `json:"finality_blocks"`
	BridgeContract string        `json:"bridge_contract"`
	NativeToken    string        `json:"native_token"`
	GasPrice       *big.Int      `json:"gas_price"`
}

type BridgeSimulator struct {
	*base.BaseSimulator
	bridgeType      BridgeType
	securityModel   string
	transactions    map[string]*BridgeTransaction
	pools           map[string]*BridgePool
	validators      map[string]*BridgeValidator
	chains          map[string]*ChainConfig
	requiredSigs    int
	confirmations   int
	challengePeriod time.Duration
	totalBridged    *big.Int
	totalFees       *big.Int
}

func NewBridgeSimulator() *BridgeSimulator {
	sim := &BridgeSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"bridge",
			"跨链桥演示器",
			"演示锁定-铸造、流动性桥、多签验证等跨链资产转移机制。",
			"crosschain",
			types.ComponentProcess,
		),
		transactions: make(map[string]*BridgeTransaction),
		pools:        make(map[string]*BridgePool),
		validators:   make(map[string]*BridgeValidator),
		chains:       make(map[string]*ChainConfig),
		totalBridged: big.NewInt(0),
		totalFees:    big.NewInt(0),
	}

	sim.AddParam(types.Param{
		Key:         "bridge_type",
		Name:        "桥类型",
		Description: "跨链桥的工作模式。",
		Type:        types.ParamTypeSelect,
		Default:     "lock_mint",
		Options: []types.Option{
			{Label: "锁定-铸造", Value: "lock_mint"},
			{Label: "销毁-解锁", Value: "burn_unlock"},
			{Label: "流动性桥", Value: "liquidity"},
			{Label: "原生桥", Value: "native"},
		},
	})
	sim.AddParam(types.Param{
		Key:         "security_model",
		Name:        "安全模型",
		Description: "桥的安全验证机制。",
		Type:        types.ParamTypeSelect,
		Default:     "multisig",
		Options: []types.Option{
			{Label: "多签验证", Value: "multisig"},
			{Label: "乐观验证", Value: "optimistic"},
			{Label: "ZK 证明", Value: "zkproof"},
			{Label: "轻客户端", Value: "lightclient"},
		},
	})
	sim.AddParam(types.Param{
		Key:         "required_confirmations",
		Name:        "确认数",
		Description: "源链所需区块确认数。",
		Type:        types.ParamTypeInt,
		Default:     12,
		Min:         1,
		Max:         100,
	})
	sim.AddParam(types.Param{
		Key:         "required_signatures",
		Name:        "签名数",
		Description: "多签模式下所需验证者签名数。",
		Type:        types.ParamTypeInt,
		Default:     5,
		Min:         1,
		Max:         21,
	})

	return sim
}

func (s *BridgeSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	s.bridgeType = BridgeLockMint
	s.securityModel = "multisig"
	s.confirmations = 12
	s.requiredSigs = 5
	s.challengePeriod = 7 * 24 * time.Hour

	if v, ok := config.Params["bridge_type"]; ok {
		if t, ok := v.(string); ok {
			s.bridgeType = BridgeType(t)
		}
	}
	if v, ok := config.Params["security_model"]; ok {
		if m, ok := v.(string); ok {
			s.securityModel = m
		}
	}
	if v, ok := config.Params["required_confirmations"]; ok {
		if n, ok := v.(float64); ok {
			s.confirmations = int(n)
		}
	}
	if v, ok := config.Params["required_signatures"]; ok {
		if n, ok := v.(float64); ok {
			s.requiredSigs = int(n)
		}
	}

	s.transactions = make(map[string]*BridgeTransaction)
	s.pools = make(map[string]*BridgePool)
	s.validators = make(map[string]*BridgeValidator)
	s.chains = make(map[string]*ChainConfig)
	s.totalBridged = big.NewInt(0)
	s.totalFees = big.NewInt(0)

	s.initializeChains()
	s.initializePools()
	s.initializeValidators()
	s.updateState()
	return nil
}

func (s *BridgeSimulator) initializeChains() {
	s.chains = map[string]*ChainConfig{
		"ethereum": {
			ChainID: "1", ChainName: "Ethereum", ChainType: "EVM",
			BlockTime: 12 * time.Second, Finality: 12,
			BridgeContract: "0x1234567890abcdef", NativeToken: "ETH",
			GasPrice: big.NewInt(30000000000),
		},
		"polygon": {
			ChainID: "137", ChainName: "Polygon", ChainType: "EVM",
			BlockTime: 2 * time.Second, Finality: 256,
			BridgeContract: "0xabcdef1234567890", NativeToken: "MATIC",
			GasPrice: big.NewInt(100000000000),
		},
		"arbitrum": {
			ChainID: "42161", ChainName: "Arbitrum One", ChainType: "EVM-L2",
			BlockTime: 250 * time.Millisecond, Finality: 1,
			BridgeContract: "0x2345678901abcdef", NativeToken: "ETH",
			GasPrice: big.NewInt(100000000),
		},
		"optimism": {
			ChainID: "10", ChainName: "Optimism", ChainType: "EVM-L2",
			BlockTime: 2 * time.Second, Finality: 1,
			BridgeContract: "0x3456789012abcdef", NativeToken: "ETH",
			GasPrice: big.NewInt(1000000),
		},
		"bsc": {
			ChainID: "56", ChainName: "BNB Chain", ChainType: "EVM",
			BlockTime: 3 * time.Second, Finality: 15,
			BridgeContract: "0x4567890123abcdef", NativeToken: "BNB",
			GasPrice: big.NewInt(5000000000),
		},
	}
}

func (s *BridgeSimulator) initializePools() {
	pools := []struct {
		token, src, dst string
		locked, liq     int64
		fee             float64
	}{
		{"ETH", "ethereum", "polygon", 1000, 500, 0.1},
		{"ETH", "ethereum", "arbitrum", 2000, 800, 0.05},
		{"ETH", "ethereum", "optimism", 1500, 600, 0.05},
		{"USDC", "ethereum", "polygon", 5000000, 2000000, 0.05},
		{"USDC", "ethereum", "arbitrum", 8000000, 3000000, 0.03},
	}

	for _, p := range pools {
		key := fmt.Sprintf("%s-%s-%s", p.token, p.src, p.dst)
		locked := new(big.Int).Mul(big.NewInt(p.locked), big.NewInt(1e18))
		liq := new(big.Int).Mul(big.NewInt(p.liq), big.NewInt(1e18))
		s.pools[key] = &BridgePool{
			ID: key, Token: p.token, SourceChain: p.src, DestChain: p.dst,
			SourceLocked: locked, DestMinted: new(big.Int).Set(locked),
			SourceLiquidity: liq, DestLiquidity: liq,
			FeePercent: p.fee, TotalVolume: big.NewInt(0), TxCount: 0,
		}
	}
}

func (s *BridgeSimulator) initializeValidators() {
	validators := []struct {
		name  string
		stake int64
	}{
		{"Guardian-1", 100000},
		{"Guardian-2", 80000},
		{"Guardian-3", 75000},
		{"Guardian-4", 70000},
		{"Guardian-5", 65000},
		{"Guardian-6", 60000},
		{"Guardian-7", 55000},
	}

	totalStake := int64(0)
	for _, v := range validators {
		totalStake += v.stake
	}

	for _, v := range validators {
		stake := new(big.Int).Mul(big.NewInt(v.stake), big.NewInt(1e18))
		hash := sha256.Sum256([]byte(v.name))
		s.validators[v.name] = &BridgeValidator{
			Address:      fmt.Sprintf("0x%s", hex.EncodeToString(hash[:20])),
			PublicKey:    hex.EncodeToString(hash[:]),
			Stake:        stake,
			VotingPower:  float64(v.stake) / float64(totalStake),
			IsActive:     true,
			SlashedAmt:   big.NewInt(0),
			JoinedAt:     time.Now().Add(-30 * 24 * time.Hour),
			LastActiveAt: time.Now(),
		}
	}
}

func (s *BridgeSimulator) ExplainBridgeTypes() map[string]interface{} {
	return map[string]interface{}{
		"lock_mint": map[string]interface{}{
			"name":        "锁定-铸造模式",
			"description": "用户在源链锁定原生资产，在目标链铸造等值包装资产。",
		},
		"burn_unlock": map[string]interface{}{
			"name":        "销毁-解锁模式",
			"description": "用户在目标链销毁包装资产，在源链解锁原生资产。",
		},
		"liquidity": map[string]interface{}{
			"name":        "流动性桥模式",
			"description": "两端预置流动性池，桥按目标链即时兑付。",
		},
		"native": map[string]interface{}{
			"name":        "原生桥模式",
			"description": "依赖链原生跨链消息或轻客户端验证。",
		},
	}
}

func (s *BridgeSimulator) ExplainSecurityModels() map[string]interface{} {
	return map[string]interface{}{
		"multisig": map[string]interface{}{
			"name":      "多签验证",
			"threshold": fmt.Sprintf("%d/%d", s.requiredSigs, len(s.validators)),
		},
		"optimistic": map[string]interface{}{
			"name":             "乐观验证",
			"challenge_period": s.challengePeriod.String(),
		},
		"zkproof": map[string]interface{}{
			"name": "ZK 证明",
		},
		"lightclient": map[string]interface{}{
			"name": "轻客户端验证",
		},
	}
}

func (s *BridgeSimulator) InitiateBridge(user, sourceChain, destChain, token string, amount *big.Int) (*BridgeTransaction, error) {
	srcChain, ok := s.chains[sourceChain]
	if !ok {
		return nil, fmt.Errorf("source chain not found: %s", sourceChain)
	}
	if _, ok := s.chains[destChain]; !ok {
		return nil, fmt.Errorf("destination chain not found: %s", destChain)
	}

	poolKey := fmt.Sprintf("%s-%s-%s", token, sourceChain, destChain)
	pool := s.pools[poolKey]
	if pool == nil {
		pool = &BridgePool{
			ID: poolKey, Token: token, SourceChain: sourceChain, DestChain: destChain,
			SourceLocked: big.NewInt(0), DestMinted: big.NewInt(0),
			SourceLiquidity: big.NewInt(0), DestLiquidity: big.NewInt(0),
			FeePercent: 0.1, TotalVolume: big.NewInt(0), TxCount: 0,
		}
		s.pools[poolKey] = pool
	}

	fee := new(big.Int).Mul(amount, big.NewInt(int64(pool.FeePercent*1000)))
	fee.Div(fee, big.NewInt(100000))

	txData := fmt.Sprintf("%s-%s-%s-%d", user, sourceChain, destChain, time.Now().UnixNano())
	txHash := sha256.Sum256([]byte(txData))
	txID := fmt.Sprintf("bridge-%s", hex.EncodeToString(txHash[:8]))

	tx := &BridgeTransaction{
		ID:              txID,
		User:            user,
		SourceChain:     sourceChain,
		DestChain:       destChain,
		Token:           token,
		WrappedToken:    fmt.Sprintf("w%s", token),
		Amount:          amount,
		Fee:             fee,
		Status:          BridgeTxPending,
		SourceTxHash:    fmt.Sprintf("0x%s", hex.EncodeToString(txHash[:])),
		RequiredConfirm: srcChain.Finality,
		Signatures:      make([]string, 0),
		RequiredSigs:    s.requiredSigs,
		MerkleProof:     make([]string, 0),
		CreatedAt:       time.Now(),
	}
	if s.securityModel == "optimistic" {
		tx.ChallengeEnd = time.Now().Add(s.challengePeriod)
	}

	s.transactions[txID] = tx
	pool.SourceLocked.Add(pool.SourceLocked, amount)
	pool.TotalVolume.Add(pool.TotalVolume, amount)
	pool.TxCount++
	s.totalBridged.Add(s.totalBridged, amount)
	s.totalFees.Add(s.totalFees, fee)

	s.EmitEvent("bridge_initiated", "", "", map[string]interface{}{
		"tx_id": txID, "user": user, "source_chain": sourceChain, "dest_chain": destChain,
		"token": token, "amount": amount.String(), "fee": fee.String(),
		"required_confirmations": srcChain.Finality,
	})

	s.updateState()
	return tx, nil
}

func (s *BridgeSimulator) ConfirmTransaction(txID string, confirmations int) error {
	tx, ok := s.transactions[txID]
	if !ok {
		return fmt.Errorf("transaction not found: %s", txID)
	}
	if tx.Status != BridgeTxPending {
		return fmt.Errorf("transaction status is not pending: %s", tx.Status)
	}

	tx.Confirmations = confirmations
	if tx.Confirmations >= tx.RequiredConfirm {
		tx.Status = BridgeTxConfirmed
		tx.ConfirmedAt = time.Now()
		s.EmitEvent("bridge_confirmed", "", "", map[string]interface{}{
			"tx_id": txID, "confirmations": confirmations,
		})
	}

	s.updateState()
	return nil
}

func (s *BridgeSimulator) SignTransaction(txID, validatorName string) error {
	tx, ok := s.transactions[txID]
	if !ok {
		return fmt.Errorf("transaction not found: %s", txID)
	}
	if tx.Status != BridgeTxConfirmed && tx.Status != BridgeTxSigned {
		return fmt.Errorf("transaction is not ready for signing: %s", tx.Status)
	}

	validator, ok := s.validators[validatorName]
	if !ok {
		return fmt.Errorf("validator not found: %s", validatorName)
	}
	if !validator.IsActive {
		return fmt.Errorf("validator is inactive: %s", validatorName)
	}
	for _, sig := range tx.Signatures {
		if sig == validatorName {
			return fmt.Errorf("validator already signed: %s", validatorName)
		}
	}

	sigData := fmt.Sprintf("%s-%s-%d", txID, validatorName, time.Now().UnixNano())
	sigHash := sha256.Sum256([]byte(sigData))
	signature := hex.EncodeToString(sigHash[:])

	tx.Signatures = append(tx.Signatures, validatorName)
	validator.SignedCount++
	validator.LastActiveAt = time.Now()
	if len(tx.Signatures) >= tx.RequiredSigs {
		tx.Status = BridgeTxSigned
	}

	s.EmitEvent("bridge_signed", "", "", map[string]interface{}{
		"tx_id": txID, "validator": validatorName, "signature": signature[:16] + "...",
		"signatures": len(tx.Signatures), "required": tx.RequiredSigs,
	})

	s.updateState()
	return nil
}

func (s *BridgeSimulator) ExecuteBridge(txID string) error {
	tx, ok := s.transactions[txID]
	if !ok {
		return fmt.Errorf("transaction not found: %s", txID)
	}

	if s.securityModel == "multisig" && tx.Status != BridgeTxSigned {
		return fmt.Errorf("transaction signatures are insufficient: %d/%d", len(tx.Signatures), tx.RequiredSigs)
	}
	if s.securityModel == "optimistic" && time.Now().Before(tx.ChallengeEnd) {
		return fmt.Errorf("challenge period has not ended: %s remaining", time.Until(tx.ChallengeEnd))
	}

	poolKey := fmt.Sprintf("%s-%s-%s", tx.Token, tx.SourceChain, tx.DestChain)
	if pool := s.pools[poolKey]; pool != nil {
		netAmount := new(big.Int).Sub(tx.Amount, tx.Fee)
		pool.DestMinted.Add(pool.DestMinted, netAmount)
	}

	destData := fmt.Sprintf("%s-dest-%d", txID, time.Now().UnixNano())
	destHash := sha256.Sum256([]byte(destData))
	tx.DestTxHash = fmt.Sprintf("0x%s", hex.EncodeToString(destHash[:]))
	tx.Status = BridgeTxCompleted
	tx.CompletedAt = time.Now()

	s.EmitEvent("bridge_completed", "", "", map[string]interface{}{
		"tx_id": txID, "dest_tx_hash": tx.DestTxHash,
		"duration": tx.CompletedAt.Sub(tx.CreatedAt).String(),
		"net_amount": new(big.Int).Sub(tx.Amount, tx.Fee).String(),
	})

	s.updateState()
	return nil
}

func (s *BridgeSimulator) SimulateLockMintFlow(user string, amount *big.Int) map[string]interface{} {
	steps := make([]map[string]interface{}, 0)

	tx, _ := s.InitiateBridge(user, "ethereum", "polygon", "ETH", amount)
	steps = append(steps, map[string]interface{}{
		"step": 1, "action": "用户在 Ethereum 发起锁定交易",
		"tx_id": tx.ID, "status": string(tx.Status),
	})

	srcChain := s.chains["ethereum"]
	for i := 1; i <= srcChain.Finality; i++ {
		s.ConfirmTransaction(tx.ID, i)
	}
	steps = append(steps, map[string]interface{}{
		"step": 2, "action": fmt.Sprintf("等待 %d 个区块确认", srcChain.Finality),
		"confirmations": srcChain.Finality, "status": string(tx.Status),
	})

	validatorNames := make([]string, 0, len(s.validators))
	for name := range s.validators {
		validatorNames = append(validatorNames, name)
	}
	sort.Strings(validatorNames)

	for i := 0; i < s.requiredSigs && i < len(validatorNames); i++ {
		s.SignTransaction(tx.ID, validatorNames[i])
	}
	steps = append(steps, map[string]interface{}{
		"step": 3, "action": fmt.Sprintf("验证者多签确认（%d/%d）", s.requiredSigs, len(s.validators)),
		"signatures": len(tx.Signatures), "status": string(tx.Status),
	})

	_ = s.ExecuteBridge(tx.ID)
	steps = append(steps, map[string]interface{}{
		"step": 4, "action": "在 Polygon 链完成包装资产铸造",
		"dest_tx_hash": tx.DestTxHash, "status": string(tx.Status),
	})

	return map[string]interface{}{
		"tx_id":      tx.ID,
		"user":       user,
		"amount":     amount.String(),
		"fee":        tx.Fee.String(),
		"net_amount": new(big.Int).Sub(tx.Amount, tx.Fee).String(),
		"duration":   tx.CompletedAt.Sub(tx.CreatedAt).String(),
		"steps":      steps,
	}
}

func (s *BridgeSimulator) SimulateLiquidityBridgeFlow(user string, amount *big.Int) map[string]interface{} {
	steps := make([]map[string]interface{}, 0)
	poolKey := "ETH-ethereum-arbitrum"
	pool := s.pools[poolKey]
	if pool == nil {
		return map[string]interface{}{"error": "流动性池不存在"}
	}
	if pool.DestLiquidity.Cmp(amount) < 0 {
		return map[string]interface{}{"error": "流动性不足"}
	}

	steps = append(steps, map[string]interface{}{
		"step": 1, "action": "检查目标链流动性池",
		"available": pool.DestLiquidity.String(), "required": amount.String(),
	})

	fee := new(big.Int).Mul(amount, big.NewInt(int64(pool.FeePercent*1000)))
	fee.Div(fee, big.NewInt(100000))
	steps = append(steps, map[string]interface{}{
		"step": 2, "action": "用户在源链存入资产",
		"amount": amount.String(), "fee": fee.String(),
	})

	netAmount := new(big.Int).Sub(amount, fee)
	pool.SourceLiquidity.Add(pool.SourceLiquidity, amount)
	pool.DestLiquidity.Sub(pool.DestLiquidity, netAmount)

	steps = append(steps, map[string]interface{}{
		"step": 3, "action": "目标链立即释放等值资产",
		"net_amount": netAmount.String(), "instant": true,
	})
	steps = append(steps, map[string]interface{}{
		"step": 4, "action": "后台异步再平衡流动性池",
		"description": "LP 或桥接维护者负责后续的再平衡。",
	})

	return map[string]interface{}{
		"mode":       "liquidity_bridge",
		"pool":       poolKey,
		"amount":     amount.String(),
		"fee":        fee.String(),
		"net_amount": netAmount.String(),
		"instant":    true,
		"steps":      steps,
	}
}

func (s *BridgeSimulator) GetBridgeStats() map[string]interface{} {
	pendingCount := 0
	completedCount := 0
	for _, tx := range s.transactions {
		if tx.Status == BridgeTxPending || tx.Status == BridgeTxConfirmed || tx.Status == BridgeTxSigned {
			pendingCount++
		} else if tx.Status == BridgeTxCompleted {
			completedCount++
		}
	}

	totalLocked := big.NewInt(0)
	totalMinted := big.NewInt(0)
	for _, pool := range s.pools {
		totalLocked.Add(totalLocked, pool.SourceLocked)
		totalMinted.Add(totalMinted, pool.DestMinted)
	}

	activeValidators := 0
	for _, v := range s.validators {
		if v.IsActive {
			activeValidators++
		}
	}

	return map[string]interface{}{
		"bridge_type":       string(s.bridgeType),
		"security_model":    s.securityModel,
		"total_bridged":     s.totalBridged.String(),
		"total_fees":        s.totalFees.String(),
		"total_locked":      totalLocked.String(),
		"total_minted":      totalMinted.String(),
		"pending_txs":       pendingCount,
		"completed_txs":     completedCount,
		"pool_count":        len(s.pools),
		"validator_count":   len(s.validators),
		"active_validators": activeValidators,
		"required_sigs":     s.requiredSigs,
		"chain_count":       len(s.chains),
	}
}

func (s *BridgeSimulator) GetPoolInfo(poolID string) map[string]interface{} {
	pool, ok := s.pools[poolID]
	if !ok {
		return nil
	}

	return map[string]interface{}{
		"id":               pool.ID,
		"token":            pool.Token,
		"source_chain":     pool.SourceChain,
		"dest_chain":       pool.DestChain,
		"source_locked":    pool.SourceLocked.String(),
		"dest_minted":      pool.DestMinted.String(),
		"source_liquidity": pool.SourceLiquidity.String(),
		"dest_liquidity":   pool.DestLiquidity.String(),
		"fee_percent":      pool.FeePercent,
		"total_volume":     pool.TotalVolume.String(),
		"tx_count":         pool.TxCount,
	}
}

func (s *BridgeSimulator) GetTransactionInfo(txID string) map[string]interface{} {
	tx, ok := s.transactions[txID]
	if !ok {
		return nil
	}

	return map[string]interface{}{
		"id":               tx.ID,
		"user":             tx.User,
		"source_chain":     tx.SourceChain,
		"dest_chain":       tx.DestChain,
		"token":            tx.Token,
		"wrapped_token":    tx.WrappedToken,
		"amount":           tx.Amount.String(),
		"fee":              tx.Fee.String(),
		"status":           string(tx.Status),
		"source_tx_hash":   tx.SourceTxHash,
		"dest_tx_hash":     tx.DestTxHash,
		"confirmations":    tx.Confirmations,
		"required_confirm": tx.RequiredConfirm,
		"signatures":       len(tx.Signatures),
		"required_sigs":    tx.RequiredSigs,
		"created_at":       tx.CreatedAt,
		"completed_at":     tx.CompletedAt,
	}
}

func (s *BridgeSimulator) updateState() {
	s.SetGlobalData("bridge_type", string(s.bridgeType))
	s.SetGlobalData("security_model", s.securityModel)
	s.SetGlobalData("transaction_count", len(s.transactions))
	s.SetGlobalData("pool_count", len(s.pools))
	s.SetGlobalData("validator_count", len(s.validators))
	s.SetGlobalData("chain_count", len(s.chains))
	s.SetGlobalData("total_bridged", s.totalBridged.String())
	s.SetGlobalData("recent_transactions", s.buildRecentTransactions())
	s.SetGlobalData("bridge_pools", s.buildPoolSnapshots())
	s.SetGlobalData("validator_overview", s.buildValidatorOverview())
	s.SetGlobalData("chain_overview", s.buildChainOverview())

	stage := "idle"
	summary := "当前桥接流程处于待机状态，可以发起跨链流程观察源链、验证层与目标链的闭环。"
	nextHint := "建议先执行锁定-铸造流程，观察确认数、签名数和目标链执行如何逐步推进。"
	progress := 0.0
	recent := s.buildRecentTransactions()
	if len(recent) > 0 {
		tx := recent[0]
		stage = fmt.Sprintf("%v", tx["status"])
		summary = fmt.Sprintf(
			"最近一笔跨链请求当前状态为 %v，确认数 %v/%v，签名数 %v/%v。",
			tx["status"], tx["confirmations"], tx["required_confirm"], tx["signatures"], tx["required_sigs"],
		)
		nextHint = "继续观察是否顺利进入目标链执行，或在哪一步被攻击/故障打断。"
		progress = 75
		if tx["status"] == string(BridgeTxCompleted) {
			progress = 100
		}
	}

	setCrosschainTeachingState(
		s.BaseSimulator,
		"crosschain",
		stage,
		summary,
		nextHint,
		progress,
		map[string]interface{}{
			"bridge_type":       string(s.bridgeType),
			"security_model":    s.securityModel,
			"transaction_count": len(s.transactions),
			"total_bridged":     s.totalBridged.String(),
		},
	)
}

func (s *BridgeSimulator) buildRecentTransactions() []map[string]interface{} {
	transactions := make([]*BridgeTransaction, 0, len(s.transactions))
	for _, tx := range s.transactions {
		transactions = append(transactions, tx)
	}
	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].CreatedAt.After(transactions[j].CreatedAt)
	})

	limit := 5
	if len(transactions) < limit {
		limit = len(transactions)
	}

	items := make([]map[string]interface{}, 0, limit)
	for _, tx := range transactions[:limit] {
		challengeEnd := ""
		if !tx.ChallengeEnd.IsZero() {
			challengeEnd = tx.ChallengeEnd.Format(time.RFC3339)
		}
		confirmedAt := ""
		if !tx.ConfirmedAt.IsZero() {
			confirmedAt = tx.ConfirmedAt.Format(time.RFC3339)
		}
		completedAt := ""
		if !tx.CompletedAt.IsZero() {
			completedAt = tx.CompletedAt.Format(time.RFC3339)
		}

		items = append(items, map[string]interface{}{
			"id":                 tx.ID,
			"user":               tx.User,
			"source_chain":       tx.SourceChain,
			"dest_chain":         tx.DestChain,
			"token":              tx.Token,
			"wrapped_token":      tx.WrappedToken,
			"amount":             tx.Amount.String(),
			"fee":                tx.Fee.String(),
			"status":             string(tx.Status),
			"confirmations":      tx.Confirmations,
			"required_confirm":   tx.RequiredConfirm,
			"signatures":         len(tx.Signatures),
			"required_sigs":      tx.RequiredSigs,
			"source_tx_hash":     tx.SourceTxHash,
			"dest_tx_hash":       tx.DestTxHash,
			"challenge_end":      challengeEnd,
			"created_at":         tx.CreatedAt.Format(time.RFC3339),
			"confirmed_at":       confirmedAt,
			"completed_at":       completedAt,
			"merkle_proof_count": len(tx.MerkleProof),
		})
	}
	return items
}

func (s *BridgeSimulator) buildPoolSnapshots() []map[string]interface{} {
	keys := make([]string, 0, len(s.pools))
	for key := range s.pools {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	limit := 6
	if len(keys) < limit {
		limit = len(keys)
	}

	items := make([]map[string]interface{}, 0, limit)
	for _, key := range keys[:limit] {
		pool := s.pools[key]
		items = append(items, map[string]interface{}{
			"id":               pool.ID,
			"token":            pool.Token,
			"source_chain":     pool.SourceChain,
			"dest_chain":       pool.DestChain,
			"source_locked":    pool.SourceLocked.String(),
			"dest_minted":      pool.DestMinted.String(),
			"source_liquidity": pool.SourceLiquidity.String(),
			"dest_liquidity":   pool.DestLiquidity.String(),
			"fee_percent":      pool.FeePercent,
			"total_volume":     pool.TotalVolume.String(),
			"tx_count":         pool.TxCount,
		})
	}
	return items
}

func (s *BridgeSimulator) buildValidatorOverview() []map[string]interface{} {
	keys := make([]string, 0, len(s.validators))
	for key := range s.validators {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	limit := 7
	if len(keys) < limit {
		limit = len(keys)
	}

	items := make([]map[string]interface{}, 0, limit)
	for _, key := range keys[:limit] {
		validator := s.validators[key]
		items = append(items, map[string]interface{}{
			"name":           key,
			"address":        validator.Address,
			"stake":          validator.Stake.String(),
			"voting_power":   validator.VotingPower,
			"is_active":      validator.IsActive,
			"signed_count":   validator.SignedCount,
			"missed_count":   validator.MissedCount,
			"slashed_amount": validator.SlashedAmt.String(),
		})
	}
	return items
}

func (s *BridgeSimulator) buildChainOverview() []map[string]interface{} {
	keys := make([]string, 0, len(s.chains))
	for key := range s.chains {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	items := make([]map[string]interface{}, 0, len(keys))
	for _, key := range keys {
		chain := s.chains[key]
		items = append(items, map[string]interface{}{
			"chain_id":        chain.ChainID,
			"chain_name":      chain.ChainName,
			"chain_type":      chain.ChainType,
			"block_time":      chain.BlockTime.String(),
			"finality_blocks": chain.Finality,
			"bridge_contract": chain.BridgeContract,
			"native_token":    chain.NativeToken,
			"gas_price":       chain.GasPrice.String(),
		})
	}
	return items
}

// ExecuteAction 执行跨链桥教学动作。
func (s *BridgeSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "simulate_lock_mint":
		user := bridgeStringParam(params, "user", "student")
		amount := bridgeAmountParam(params, "amount", 1)
		result := s.SimulateLockMintFlow(user, amount)
		s.updateState()
		return crosschainActionResult(
			"已完成一次锁定-铸造跨链流程",
			result,
			&types.ActionFeedback{
				Summary:     "源链锁定、桥接验证和目标链铸造流程已完成。",
				NextHint:    "观察确认数、签名数和目标链执行结果是否形成完整闭环。",
				EffectScope: "crosschain",
				ResultState: map[string]interface{}{"status": "lock_mint_completed"},
			},
		), nil
	case "simulate_liquidity_bridge":
		user := bridgeStringParam(params, "user", "student")
		amount := bridgeAmountParam(params, "amount", 1)
		result := s.SimulateLiquidityBridgeFlow(user, amount)
		if errorMessage, ok := result["error"].(string); ok && errorMessage != "" {
			return nil, fmt.Errorf(errorMessage)
		}
		s.updateState()
		return crosschainActionResult(
			"已完成一次流动性桥跨链流程",
			result,
			&types.ActionFeedback{
				Summary:     "流动性桥已经完成资产转移与目标链到账。",
				NextHint:    "观察桥池余额、验证过程和目标链结果是否一致。",
				EffectScope: "crosschain",
				ResultState: map[string]interface{}{"status": "liquidity_bridge_completed"},
			},
		), nil
	case "reset_bridge":
		config := types.Config{
			Params: map[string]interface{}{
				"bridge_type":            string(s.bridgeType),
				"security_model":         s.securityModel,
				"required_confirmations": float64(s.confirmations),
				"required_signatures":    float64(s.requiredSigs),
			},
		}
		if err := s.Init(config); err != nil {
			return nil, err
		}
		return crosschainActionResult(
			"已重置跨链桥场景",
			nil,
			&types.ActionFeedback{
				Summary:     "跨链桥实验状态已恢复到初始配置。",
				NextHint:    "可以重新观察锁定-铸造或流动性桥的完整生命周期。",
				EffectScope: "crosschain",
				ResultState: map[string]interface{}{"status": "bridge_reset"},
			},
		), nil
	default:
		return nil, fmt.Errorf("unsupported bridge action: %s", action)
	}
}

func bridgeStringParam(params map[string]interface{}, key, fallback string) string {
	if params == nil {
		return fallback
	}
	if value, ok := params[key].(string); ok && value != "" {
		return value
	}
	return fallback
}

func bridgeAmountParam(params map[string]interface{}, key string, fallback int64) *big.Int {
	if params == nil {
		return big.NewInt(fallback)
	}
	if value, ok := params[key].(float64); ok && value > 0 {
		return big.NewInt(int64(value))
	}
	if value, ok := params[key].(int64); ok && value > 0 {
		return big.NewInt(value)
	}
	if value, ok := params[key].(int); ok && value > 0 {
		return big.NewInt(int64(value))
	}
	return big.NewInt(fallback)
}

type BridgeFactory struct{}

func (f *BridgeFactory) Create() engine.Simulator {
	return NewBridgeSimulator()
}

func (f *BridgeFactory) GetDescription() types.Description {
	return NewBridgeSimulator().GetDescription()
}

func NewBridgeFactory() *BridgeFactory {
	return &BridgeFactory{}
}

var _ engine.SimulatorFactory = (*BridgeFactory)(nil)
