package service

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/chainspace/backend/internal/model"
	"gopkg.in/yaml.v3"
)

type runtimeImageManifest struct {
	Registry  string                      `yaml:"registry"`
	Namespace string                      `yaml:"namespace"`
	Entries   map[string]imageManifestRef `yaml:",inline"`
}

type imageManifestRef struct {
	Version string `yaml:"version"`
}

var (
	runtimeImageManifestOnce sync.Once
	runtimeImageManifestData runtimeImageManifest
)

func loadRuntimeImageManifest() runtimeImageManifest {
	runtimeImageManifestOnce.Do(func() {
		runtimeImageManifestData = runtimeImageManifest{
			Registry:  "registry.chainspace.com",
			Namespace: "chainspace",
			Entries:   map[string]imageManifestRef{},
		}

		for _, candidate := range []string{
			filepath.Join("deploy", "images", "versions.yaml"),
			filepath.Join("..", "deploy", "images", "versions.yaml"),
		} {
			raw, err := os.ReadFile(candidate)
			if err != nil {
				continue
			}

			manifest := runtimeImageManifest{}
			if err := yaml.Unmarshal(raw, &manifest); err != nil {
				continue
			}
			if manifest.Registry == "" {
				manifest.Registry = runtimeImageManifestData.Registry
			}
			if manifest.Namespace == "" {
				manifest.Namespace = runtimeImageManifestData.Namespace
			}
			if manifest.Entries == nil {
				manifest.Entries = map[string]imageManifestRef{}
			}
			runtimeImageManifestData = manifest
			return
		}
	})

	return runtimeImageManifestData
}

func resolveImageFromManifest(name string) string {
	key := strings.TrimSpace(name)
	if key == "" {
		return ""
	}

	manifest := loadRuntimeImageManifest()
	entry, ok := manifest.Entries[key]
	if !ok || strings.TrimSpace(entry.Version) == "" {
		return ""
	}

	namespace := strings.Trim(manifest.Namespace, "/")
	if namespace == "" {
		namespace = "chainspace"
	}

	// 运行时镜像引用统一对齐项目文档和初始化数据：
	// 本地 Docker / 本地 K8s 直接使用 chainspace/<image>:latest。
	// versions.yaml 只承载镜像版本清单与构建元数据，不作为本地运行时拉取地址。
	return namespace + "/" + key + ":latest"
}

func resolvePlatformImage(name, fallback string) string {
	if resolved := resolveImageFromManifest(name); resolved != "" {
		return resolved
	}
	return fallback
}

func resolveImageAlias(hint string) string {
	value := strings.ToLower(strings.TrimSpace(hint))
	switch {
	case value == "", value == "workspace":
		return "eth-dev"
	case strings.Contains(value, "security"), strings.Contains(value, "slither"), strings.Contains(value, "mythril"):
		return "security"
	case strings.Contains(value, "crypto"):
		return "crypto"
	case strings.Contains(value, "simulation"), strings.Contains(value, "visualization"), strings.Contains(value, "simulator"):
		return "simulation"
	case strings.Contains(value, "chainlink"):
		return "chainlink"
	case strings.Contains(value, "blockscout"), strings.Contains(value, "explorer"):
		return "blockscout"
	case strings.Contains(value, "geth"):
		return "geth"
	case strings.Contains(value, "ipfs"):
		return "ipfs"
	case strings.Contains(value, "fabric"), strings.Contains(value, "peer"), strings.Contains(value, "orderer"), strings.Contains(value, "fabric-ca"):
		return "fabric"
	case strings.Contains(value, "fisco"), strings.Contains(value, "webase"):
		return "fisco"
	case strings.Contains(value, "chainmaker"):
		return "chainmaker"
	case strings.Contains(value, "bitcoin"), strings.Contains(value, "bitcoind"):
		return "bitcoin"
	case strings.Contains(value, "solana"), strings.Contains(value, "anchor"), strings.Contains(value, "validator"):
		return "solana"
	case strings.Contains(value, "substrate"), strings.Contains(value, "polkadot"):
		return "substrate"
	case strings.Contains(value, "cosmos"):
		return "cosmos"
	case strings.Contains(value, "move"), strings.Contains(value, "aptos"), strings.Contains(value, "sui"):
		return "move"
	case strings.Contains(value, "privacy"), strings.Contains(value, "zk"), strings.Contains(value, "circom"), strings.Contains(value, "snark"):
		return "privacy"
	case strings.Contains(value, "graph"):
		return "thegraph"
	case strings.Contains(value, "rollup"), strings.Contains(value, "optimism"), strings.Contains(value, "arbitrum"), strings.Contains(value, "l2"):
		return "l2-dev"
	case strings.Contains(value, "dapp"), strings.Contains(value, "frontend"), strings.Contains(value, "web"):
		return "dapp-dev"
	case strings.Contains(value, "python"):
		return "dev-python"
	case strings.Contains(value, "rust"):
		return "dev-rust"
	case strings.Contains(value, "go"):
		return "dev-go"
	case strings.Contains(value, "node"):
		return "dev-node"
	case strings.Contains(value, "fork"), strings.Contains(value, "anvil"), strings.Contains(value, "foundry"), strings.Contains(value, "hardhat"), strings.Contains(value, "eth"), strings.Contains(value, "evm"):
		return "eth-dev"
	default:
		return ""
	}
}

func resolveRuntimeImage(explicit string, hints ...string) string {
	if strings.TrimSpace(explicit) != "" {
		return explicit
	}
	for _, hint := range hints {
		if alias := resolveImageAlias(hint); alias != "" {
			return resolvePlatformImage(alias, "chainspace/"+alias+":latest")
		}
	}
	return resolvePlatformImage("eth-dev", "chainspace/eth-dev:latest")
}

func defaultExperimentWorkspaceImage(expType string) string {
	switch expType {
	case model.ExperimentTypeToolUsage, model.ExperimentTypeReverse, model.ExperimentTypeTroubleshoot:
		return resolveRuntimeImage("", "security")
	case model.ExperimentTypeVisualization:
		return resolveRuntimeImage("", "simulation")
	default:
		return resolveRuntimeImage("", "eth-dev")
	}
}

func defaultExperimentNodeImage(expType string, node model.ExperimentNodeBlueprint) string {
	return resolveRuntimeImage(node.Image, node.Key, node.Name, node.Role, expType)
}

func defaultExperimentServiceImage(serviceSpec model.ExperimentServiceBlueprint) string {
	return resolveRuntimeImage(serviceSpec.Image, serviceSpec.Key, serviceSpec.Name, serviceSpec.Role)
}

func defaultChallengeWorkspaceImage(runtimeProfile string, orch model.ChallengeOrchestration) string {
	return resolveRuntimeImage(orch.Workspace.Image, runtimeProfile, orch.Workspace.Template, orch.Mode)
}

func defaultChallengeServiceImage(serviceSpec model.ChallengeServiceSpec) string {
	return resolveRuntimeImage(serviceSpec.Image, serviceSpec.Key, serviceSpec.Purpose, serviceSpec.Description)
}

func defaultForkRuntimeImage() string {
	return resolveRuntimeImage("", "fork", "anvil")
}

func defaultBattleWorkspaceImage() string {
	return resolveRuntimeImage("", "eth-dev", "workspace")
}

func defaultBattleSharedChainImage(chainType string) string {
	return resolveRuntimeImage("", chainType, "anvil", "eth")
}

func defaultContractCompileImage() string {
	return resolveRuntimeImage("", "eth-dev", "foundry")
}
