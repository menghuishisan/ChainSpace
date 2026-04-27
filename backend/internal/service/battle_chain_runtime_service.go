package service

import (
	"context"
	"fmt"
	"time"

	"github.com/chainspace/backend/internal/model"
	"github.com/chainspace/backend/internal/pkg/k8s"
	"github.com/chainspace/backend/internal/pkg/logger"
	"github.com/chainspace/backend/internal/repository"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

// BattleChainRuntimeService 负责对抗赛共享链运行时。
type BattleChainRuntimeService struct {
	contestRepo *repository.ContestRepository
	k8sClient   *k8s.Client
	provisioner *RuntimeProvisionService
}

func NewBattleChainRuntimeService(contestRepo *repository.ContestRepository, k8sClient *k8s.Client) *BattleChainRuntimeService {
	return &BattleChainRuntimeService{
		contestRepo: contestRepo,
		k8sClient:   k8sClient,
		provisioner: NewRuntimeProvisionService(k8sClient),
	}
}

func (s *BattleChainRuntimeService) StartSharedChain(ctx context.Context, round *model.AgentBattleRound) (string, error) {
	if s.k8sClient == nil {
		return "", fmt.Errorf("Kubernetes未启用，无法启动链环境")
	}

	contest, err := s.contestRepo.GetByID(ctx, round.ContestID)
	if err != nil {
		return s.startSharedChainWithDefaults(ctx, round)
	}

	chainSpec := contest.BattleOrchestration.SharedChain
	if chainSpec.Image == "" {
		chainSpec.Image = defaultBattleSharedChainImage(chainSpec.ChainType)
	}
	if chainSpec.BlockTime == 0 {
		chainSpec.BlockTime = 1
	}

	chainID := fmt.Sprintf("battle-%d-%d", round.ContestID, round.ID)
	rpcServiceName := fmt.Sprintf("%s-svc", chainID)
	rpcURL := fmt.Sprintf("http://%s.%s.svc.cluster.local:8545", rpcServiceName, s.k8sClient.Namespace())

	logger.Info("Starting shared battle chain",
		zap.Uint("roundID", round.ID),
		zap.String("chainID", chainID),
		zap.String("image", chainSpec.Image),
	)

	podCfg := buildBattleSharedChainPodConfig(chainID, chainSpec.Image, chainSpec.NetworkID, chainSpec.BlockTime)

	if err := s.provisioner.StartInstance(ctx, podCfg, 2*time.Minute); err != nil {
		return "", fmt.Errorf("wait for chain ready: %w", err)
	}

	return rpcURL, nil
}

func (s *BattleChainRuntimeService) startSharedChainWithDefaults(ctx context.Context, round *model.AgentBattleRound) (string, error) {
	chainID := fmt.Sprintf("battle-%d-%d", round.ContestID, round.ID)
	rpcServiceName := fmt.Sprintf("%s-svc", chainID)
	rpcURL := fmt.Sprintf("http://%s.%s.svc.cluster.local:8545", rpcServiceName, s.k8sClient.Namespace())

	logger.Info("Starting default shared chain", zap.Uint("roundID", round.ID))

	podCfg := buildBattleSharedChainPodConfig(chainID, defaultBattleSharedChainImage("ethereum"), int(31338+round.ID), 1)

	if err := s.provisioner.StartInstance(ctx, podCfg, 2*time.Minute); err != nil {
		return "", fmt.Errorf("wait for anvil ready: %w", err)
	}

	return rpcURL, nil
}

func buildBattleSharedChainPodConfig(envID, image string, networkID, blockTime int) *k8s.PodConfig {
	if networkID <= 0 {
		networkID = 31338
	}
	if blockTime <= 0 {
		blockTime = 1
	}
	if resolveImageAlias(image) == "geth" {
		return &k8s.PodConfig{
			EnvID:     envID,
			UserID:    0,
			SchoolID:  0,
			Image:     image,
			Command:   []string{"chainspace-geth-runtime"},
			CPU:       "1000m",
			Memory:    "2Gi",
			Storage:   "5Gi",
			Timeout:   5 * time.Minute,
			Ports:     []int32{8545, 8546, 30303},
			ProbePort: 8545,
			EnvVars: map[string]string{
				"CHAINSPACE_RUNTIME_KIND":     "service",
				"CHAINSPACE_GETH_MODE":        "single",
				"CHAINSPACE_GETH_CHAIN_ID":    fmt.Sprintf("%d", networkID),
				"CHAINSPACE_GETH_BLOCK_TIME":  fmt.Sprintf("%d", blockTime),
				"CHAINSPACE_GETH_HTTP_PORT":   "8545",
				"CHAINSPACE_GETH_WS_PORT":     "8546",
				"CHAINSPACE_GETH_P2P_PORT":    "30303",
				"CHAINSPACE_HEALTHCHECK_PORT": "8545",
			},
		}
	}

	return &k8s.PodConfig{
		EnvID:     envID,
		UserID:    0,
		SchoolID:  0,
		Image:     image,
		Command:   buildAnvilRuntimeCommand(anvilRuntimeOptions{ChainID: networkID, BlockTime: blockTime, Accounts: 20}),
		CPU:       "1000m",
		Memory:    "2Gi",
		Storage:   "5Gi",
		Timeout:   5 * time.Minute,
		Ports:     []int32{8545},
		ProbePort: 8545,
		EnvVars: map[string]string{
			"ANVIL_BLOCK_TIME":        fmt.Sprintf("%d", blockTime),
			"ANVIL_ACCOUNTS":          "20",
			"ANVIL_CHAIN_ID":          fmt.Sprintf("%d", networkID),
			"CHAINSPACE_RUNTIME_KIND": "service",
		},
	}
}

func (s *BattleChainRuntimeService) StopSharedChain(ctx context.Context, round *model.AgentBattleRound) error {
	if s.k8sClient == nil {
		return nil
	}

	chainID := fmt.Sprintf("battle-%d-%d", round.ContestID, round.ID)
	logger.Info("Stopping shared battle chain", zap.Uint("roundID", round.ID), zap.String("chainID", chainID))

	if err := s.provisioner.StopInstance(ctx, chainID); err != nil {
		return fmt.Errorf("delete chain pod: %w", err)
	}

	return nil
}

func (s *BattleChainRuntimeService) GetCurrentBlock(ctx context.Context, rpcURL string) (uint64, error) {
	if rpcURL == "" {
		return 0, nil
	}

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return 0, err
	}
	defer client.Close()

	return client.BlockNumber(ctx)
}
