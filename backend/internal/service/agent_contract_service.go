package service

import (
	"context"
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/chainspace/backend/internal/model"
	"github.com/chainspace/backend/internal/pkg/k8s"
	"github.com/chainspace/backend/internal/pkg/logger"
	"github.com/chainspace/backend/internal/repository"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

// AgentContractService 负责对抗赛合约校验、编译与部署运行时。
type AgentContractService struct {
	contractRepo *repository.AgentContractRepository
	k8sClient    *k8s.Client
}

func NewAgentContractService(contractRepo *repository.AgentContractRepository, k8sClient *k8s.Client) *AgentContractService {
	return &AgentContractService{
		contractRepo: contractRepo,
		k8sClient:    k8sClient,
	}
}

func (s *AgentContractService) ValidateContractCode(code string) error {
	if len(code) < 50 {
		return fmt.Errorf("合约代码过短")
	}
	if !strings.Contains(code, "pragma solidity") {
		return fmt.Errorf("缺少Solidity版本声明")
	}
	if !strings.Contains(code, "contract ") {
		return fmt.Errorf("缺少合约定义")
	}

	dangerousPatterns := []string{"selfdestruct", "delegatecall", "assembly"}
	for _, pattern := range dangerousPatterns {
		if strings.Contains(strings.ToLower(code), pattern) {
			return fmt.Errorf("禁止使用 %s", pattern)
		}
	}

	contractNameRegex := regexp.MustCompile(`contract\s+(\w+)`)
	if !contractNameRegex.MatchString(code) {
		return fmt.Errorf("无效的合约名称")
	}

	return nil
}

func (s *AgentContractService) DeployContractAsync(ctx context.Context, contractID uint, rpcURL string) {
	if s.contractRepo == nil {
		logger.Error("Contract repository is not initialized", zap.Uint("contractID", contractID))
		return
	}
	if rpcURL == "" {
		logger.Error("Battle chain RPC URL is empty", zap.Uint("contractID", contractID))
		return
	}

	contract, err := s.contractRepo.GetByID(ctx, contractID)
	if err != nil {
		logger.Error("Failed to get contract for deployment", zap.Uint("contractID", contractID), zap.Error(err))
		return
	}

	contract.Status = model.ContractStatusDeploying
	s.contractRepo.Update(ctx, contract)

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		logger.Error("Failed to connect to chain", zap.String("rpcURL", rpcURL), zap.Error(err))
		s.updateContractStatus(ctx, contract, model.ContractStatusFailed, "连接链失败")
		return
	}
	defer client.Close()

	privateKey, err := crypto.HexToECDSA("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80")
	if err != nil {
		logger.Error("Failed to parse private key", zap.Error(err))
		s.updateContractStatus(ctx, contract, model.ContractStatusFailed, "私钥解析失败")
		return
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		logger.Error("Failed to get nonce", zap.Error(err))
		s.updateContractStatus(ctx, contract, model.ContractStatusFailed, "获取nonce失败")
		return
	}

	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		logger.Error("Failed to get gas price", zap.Error(err))
		s.updateContractStatus(ctx, contract, model.ContractStatusFailed, "获取gas价格失败")
		return
	}

	chainID, err := client.NetworkID(ctx)
	if err != nil {
		logger.Error("Failed to get chain ID", zap.Error(err))
		s.updateContractStatus(ctx, contract, model.ContractStatusFailed, "获取chainID失败")
		return
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		logger.Error("Failed to create transactor", zap.Error(err))
		s.updateContractStatus(ctx, contract, model.ContractStatusFailed, "创建签名器失败")
		return
	}
	auth.Nonce = new(big.Int).SetUint64(nonce)
	auth.Value = big.NewInt(0)
	auth.GasLimit = uint64(3000000)
	auth.GasPrice = gasPrice

	bytecode, err := s.CompileContract(contract.SourceCode)
	if err != nil {
		logger.Error("Failed to compile contract", zap.Error(err))
		s.updateContractStatus(ctx, contract, model.ContractStatusFailed, "编译失败: "+err.Error())
		return
	}

	contractAddress, tx, err := s.deployBytecode(ctx, client, auth, bytecode)
	if err != nil {
		logger.Error("Failed to deploy contract", zap.Error(err))
		s.updateContractStatus(ctx, contract, model.ContractStatusFailed, "部署失败: "+err.Error())
		return
	}

	if _, err = bind.WaitMined(ctx, client, tx); err != nil {
		logger.Error("Failed to wait for deployment", zap.Error(err))
		s.updateContractStatus(ctx, contract, model.ContractStatusFailed, "等待确认失败")
		return
	}

	now := time.Now()
	contract.ContractAddress = contractAddress.Hex()
	contract.Status = model.ContractStatusDeployed
	contract.DeployedAt = &now

	if err := s.contractRepo.Update(ctx, contract); err != nil {
		logger.Error("Failed to update contract after deployment", zap.Uint("contractID", contractID), zap.Error(err))
		return
	}

	logger.Info("Contract deployed", zap.Uint("contractID", contractID), zap.String("address", contractAddress.Hex()))
}

func (s *AgentContractService) DeployUpgradeAsync(ctx context.Context, contract *model.AgentContract, round *model.AgentBattleRound, bytecode []byte) {
	if contract == nil || round == nil {
		return
	}
	if round.ChainRPCURL == "" {
		s.updateContractStatus(ctx, contract, "upgrade_failed", "empty chain rpc url")
		return
	}

	client, err := ethclient.Dial(round.ChainRPCURL)
	if err != nil {
		s.updateContractStatus(ctx, contract, "upgrade_failed", err.Error())
		return
	}
	defer client.Close()

	privateKey, err := crypto.HexToECDSA("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80")
	if err != nil {
		s.updateContractStatus(ctx, contract, "upgrade_failed", err.Error())
		return
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		s.updateContractStatus(ctx, contract, "upgrade_failed", err.Error())
		return
	}

	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		s.updateContractStatus(ctx, contract, "upgrade_failed", err.Error())
		return
	}

	chainID, err := client.NetworkID(ctx)
	if err != nil {
		s.updateContractStatus(ctx, contract, "upgrade_failed", err.Error())
		return
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		s.updateContractStatus(ctx, contract, "upgrade_failed", err.Error())
		return
	}
	auth.Nonce = new(big.Int).SetUint64(nonce)
	auth.Value = big.NewInt(0)
	auth.GasLimit = uint64(3000000)
	auth.GasPrice = gasPrice

	newAddr, _, err := s.deployBytecode(ctx, client, auth, bytecode)
	if err != nil {
		s.updateContractStatus(ctx, contract, "upgrade_failed", err.Error())
		return
	}

	contract.ContractAddress = newAddr.Hex()
	contract.Status = "deployed"
	contract.Version++
	s.contractRepo.Update(ctx, contract)

	logger.Info("Contract upgraded", zap.Uint("contractID", contract.ID), zap.String("newAddress", newAddr.Hex()))
}

func (s *AgentContractService) updateContractStatus(ctx context.Context, contract *model.AgentContract, status, errMsg string) {
	contract.Status = status
	contract.ErrorMessage = errMsg
	s.contractRepo.Update(ctx, contract)
}

func (s *AgentContractService) CompileContract(sourceCode string) ([]byte, error) {
	if s.k8sClient == nil {
		return nil, fmt.Errorf("k8s client not initialized")
	}

	contractNameRegex := regexp.MustCompile(`contract\s+(\w+)`)
	matches := contractNameRegex.FindStringSubmatch(sourceCode)
	if len(matches) < 2 {
		return nil, fmt.Errorf("contract name not found")
	}
	contractName := matches[1]

	ctx := context.Background()
	compilePodName := fmt.Sprintf("compile-%d", time.Now().UnixNano())

	podCfg := &k8s.PodConfig{
		EnvID:   compilePodName,
		Image:   defaultContractCompileImage(),
		CPU:     "500m",
		Memory:  "1Gi",
		Storage: "1Gi",
		Timeout: 5 * time.Minute,
		Ports:   []int32{},
		EnvVars: map[string]string{},
	}

	_, err := s.k8sClient.CreatePod(ctx, podCfg)
	if err != nil {
		return nil, fmt.Errorf("create compile pod: %w", err)
	}
	defer s.k8sClient.DeletePod(ctx, compilePodName)

	if err := s.k8sClient.WaitForPodReady(ctx, compilePodName, 2*time.Minute); err != nil {
		return nil, fmt.Errorf("wait for compile pod: %w", err)
	}

	if _, err := s.k8sClient.ExecCommand(ctx, compilePodName, []string{"mkdir", "-p", "/workspace/src"}); err != nil {
		return nil, fmt.Errorf("create src dir: %w", err)
	}

	sourceBase64 := base64.StdEncoding.EncodeToString([]byte(sourceCode))
	writeCmd := []string{"sh", "-c", fmt.Sprintf("echo '%s' | base64 -d > /workspace/src/%s.sol", sourceBase64, contractName)}
	if _, err := s.k8sClient.ExecCommand(ctx, compilePodName, writeCmd); err != nil {
		return nil, fmt.Errorf("write contract: %w", err)
	}

	foundryConfig := `[profile.default]
src = "src"
out = "out"
libs = ["lib"]
solc_version = "0.8.20"
optimizer = true
optimizer_runs = 200`
	configBase64 := base64.StdEncoding.EncodeToString([]byte(foundryConfig))
	writeConfigCmd := []string{"sh", "-c", fmt.Sprintf("echo '%s' | base64 -d > /workspace/foundry.toml", configBase64)}
	if _, err := s.k8sClient.ExecCommand(ctx, compilePodName, writeConfigCmd); err != nil {
		return nil, fmt.Errorf("write foundry.toml: %w", err)
	}

	buildCmd := []string{"sh", "-c", "cd /workspace && forge build 2>&1"}
	buildOutput, err := s.k8sClient.ExecCommand(ctx, compilePodName, buildCmd)
	if err != nil {
		return nil, fmt.Errorf("forge build failed: %w, output: %s", err, buildOutput)
	}

	artifactPath := fmt.Sprintf("/workspace/out/%s.sol/%s.json", contractName, contractName)
	artifactData, err := s.k8sClient.ExecCommand(ctx, compilePodName, []string{"cat", artifactPath})
	if err != nil {
		return nil, fmt.Errorf("read artifact: %w", err)
	}

	var artifact struct {
		Bytecode struct {
			Object string `json:"object"`
		} `json:"bytecode"`
	}
	if err := json.Unmarshal([]byte(artifactData), &artifact); err != nil {
		return nil, fmt.Errorf("parse artifact: %w", err)
	}

	bytecodeHex := strings.TrimPrefix(artifact.Bytecode.Object, "0x")
	bytecode := common.Hex2Bytes(bytecodeHex)
	if len(bytecode) == 0 {
		return nil, fmt.Errorf("empty bytecode after compilation")
	}

	return bytecode, nil
}

func (s *AgentContractService) deployBytecode(ctx context.Context, client *ethclient.Client, auth *bind.TransactOpts, bytecode []byte) (common.Address, *types.Transaction, error) {
	tx := types.NewContractCreation(
		auth.Nonce.Uint64(),
		auth.Value,
		auth.GasLimit,
		auth.GasPrice,
		bytecode,
	)

	signedTx, err := auth.Signer(auth.From, tx)
	if err != nil {
		return common.Address{}, nil, fmt.Errorf("sign transaction: %w", err)
	}

	if err := client.SendTransaction(ctx, signedTx); err != nil {
		return common.Address{}, nil, fmt.Errorf("send transaction: %w", err)
	}

	contractAddress := crypto.CreateAddress(auth.From, auth.Nonce.Uint64())
	return contractAddress, signedTx, nil
}
