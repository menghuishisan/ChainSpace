package service

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/chainspace/backend/internal/model"
	"github.com/ethereum/go-ethereum/crypto"
)

type runtimeBootConfig struct {
	Command   []string
	Args      []string
	EnvVars   map[string]string
	ProbePort int32
}

type gethRuntimeIdentity struct {
	InstanceKey string
	Role        string
	Host        string
	HTTPPort    int32
	WSPort      int32
	P2PPort     int32
	Signer      bool
	PrivateKey  string
	Address     string
	Enode       string
}

func buildRuntimeBootConfig(
	image string,
	exp *model.Experiment,
	env *model.ExperimentEnv,
	blueprint model.ExperimentBlueprint,
	runtime model.ExperimentRuntimeState,
	instance model.ExperimentRuntimeTarget,
) runtimeBootConfig {
	cfg := runtimeBootConfig{
		EnvVars: map[string]string{
			"CHAINSPACE_RUNTIME_KIND":      instance.Kind,
			"CHAINSPACE_INTERACTION_TOOLS": strings.Join(instance.InteractionTools, ","),
		},
	}

	alias := resolveImageAlias(image)
	switch instance.Kind {
	case "workspace":
		cfg.ProbePort = 7681
	case "simulation":
		cfg.ProbePort = 8080
	default:
		cfg.ProbePort = firstNonZeroPort(instance.Ports)
	}

	switch alias {
	case "geth":
		cfg.Command = []string{"chainspace-geth-runtime"}
		cfg.ProbePort = 8545
		applyGethRuntimeBootConfig(&cfg, exp, env, blueprint, runtime, instance)
	case "blockscout":
		cfg.ProbePort = preferredPort(instance.Ports, 4000)
	case "chainlink":
		cfg.ProbePort = preferredPort(instance.Ports, 6688, 6689)
	case "thegraph":
		cfg.ProbePort = preferredPort(instance.Ports, 8000, 8020, 8030, 8040)
	case "ipfs":
		cfg.ProbePort = preferredPort(instance.Ports, 5001, 8080, 4001)
	case "bitcoin":
		cfg.ProbePort = preferredPort(instance.Ports, 8332, 18443)
	case "solana":
		cfg.ProbePort = preferredPort(instance.Ports, 8899)
	case "substrate":
		cfg.ProbePort = preferredPort(instance.Ports, 9944, 9933)
	}

	if cfg.ProbePort > 0 {
		cfg.EnvVars["CHAINSPACE_HEALTHCHECK_PORT"] = strconv.Itoa(int(cfg.ProbePort))
	}

	return cfg
}

func applyGethRuntimeBootConfig(
	cfg *runtimeBootConfig,
	exp *model.Experiment,
	env *model.ExperimentEnv,
	blueprint model.ExperimentBlueprint,
	runtime model.ExperimentRuntimeState,
	instance model.ExperimentRuntimeTarget,
) {
	httpPort := preferredPortOrDefault(instance.Ports, 8545, 8545)
	wsPort := preferredPortOrDefault(instance.Ports, 8546, 8546)
	p2pPort := preferredPortOrDefault(instance.Ports, 30303, 30303)

	chainID := strings.TrimSpace(instance.EnvVars["CHAIN_ID"])
	if chainID == "" {
		chainID = "31337"
	}
	cfg.EnvVars["CHAINSPACE_GETH_CHAIN_ID"] = chainID
	cfg.EnvVars["CHAINSPACE_GETH_HTTP_PORT"] = strconv.Itoa(int(httpPort))
	cfg.EnvVars["CHAINSPACE_GETH_WS_PORT"] = strconv.Itoa(int(wsPort))
	cfg.EnvVars["CHAINSPACE_GETH_P2P_PORT"] = strconv.Itoa(int(p2pPort))
	cfg.EnvVars["CHAINSPACE_GETH_BLOCK_TIME"] = resolveGethBlockTime(instance.EnvVars)

	if blueprint.Mode == model.ExperimentModeSingle && instance.Kind == "service" {
		cfg.EnvVars["CHAINSPACE_GETH_MODE"] = "single"
		return
	}

	cfg.EnvVars["CHAINSPACE_GETH_MODE"] = "cluster"
	identities, signers, err := buildGethRuntimeIdentities(env.EnvID, blueprint, runtime)
	if err != nil {
		cfg.EnvVars["CHAINSPACE_GETH_MODE"] = "single"
		cfg.EnvVars["CHAINSPACE_GETH_BOOTSTRAP_ERROR"] = err.Error()
		return
	}

	current, ok := identities[instance.Key]
	if !ok {
		cfg.EnvVars["CHAINSPACE_GETH_MODE"] = "single"
		cfg.EnvVars["CHAINSPACE_GETH_BOOTSTRAP_ERROR"] = fmt.Sprintf("missing geth identity for %s", instance.Key)
		return
	}

	genesisJSON, err := buildCliqueGenesisJSON(chainID, signers, resolveGethBlockTime(instance.EnvVars))
	if err != nil {
		cfg.EnvVars["CHAINSPACE_GETH_MODE"] = "single"
		cfg.EnvVars["CHAINSPACE_GETH_BOOTSTRAP_ERROR"] = err.Error()
		return
	}

	staticNodes := make([]string, 0, len(identities)-1)
	bootnodes := make([]string, 0, len(identities)-1)
	for key, identity := range identities {
		if key == current.InstanceKey {
			continue
		}
		staticNodes = append(staticNodes, identity.Enode)
		if identity.Signer || strings.Contains(identity.Role, "validator") || strings.EqualFold(identity.Role, "rpc") {
			bootnodes = append(bootnodes, identity.Enode)
		}
	}
	if len(bootnodes) == 0 {
		bootnodes = append(bootnodes, staticNodes...)
	}

	cfg.EnvVars["CHAINSPACE_GETH_GENESIS_B64"] = base64.StdEncoding.EncodeToString(genesisJSON)
	if len(staticNodes) > 0 {
		staticRaw, _ := json.Marshal(staticNodes)
		cfg.EnvVars["CHAINSPACE_GETH_STATIC_NODES_B64"] = base64.StdEncoding.EncodeToString(staticRaw)
	}
	cfg.EnvVars["CHAINSPACE_GETH_BOOTNODES"] = strings.Join(bootnodes, ",")
	cfg.EnvVars["CHAINSPACE_GETH_MINER"] = strconv.FormatBool(current.Signer)
	if current.Signer {
		cfg.EnvVars["CHAINSPACE_GETH_SIGNER_PRIVATE_KEY"] = current.PrivateKey
		cfg.EnvVars["CHAINSPACE_GETH_SIGNER_ADDRESS"] = current.Address
	}

	if exp != nil {
		cfg.EnvVars["CHAINSPACE_GETH_EXPERIMENT_TYPE"] = exp.Type
	}
}

func buildGethRuntimeIdentities(
	envID string,
	blueprint model.ExperimentBlueprint,
	runtime model.ExperimentRuntimeState,
) (map[string]gethRuntimeIdentity, []gethRuntimeIdentity, error) {
	identities := make(map[string]gethRuntimeIdentity)
	signers := make([]gethRuntimeIdentity, 0)

	for _, instance := range runtime.Instances {
		image, role, ok := runtimeInstanceImageAndRole(blueprint, instance)
		if !ok || resolveImageAlias(image) != "geth" {
			continue
		}

		privateKey, address, nodeID, err := deriveGethIdentity(fmt.Sprintf("%s:%s", envID, instance.Key))
		if err != nil {
			return nil, nil, err
		}

		httpPort := preferredPortOrDefault(instance.Ports, 8545, 8545)
		wsPort := preferredPortOrDefault(instance.Ports, 8546, 8546)
		p2pPort := preferredPortOrDefault(instance.Ports, 30303, 30303)

		signer := strings.Contains(strings.ToLower(role), "validator")
		identity := gethRuntimeIdentity{
			InstanceKey: instance.Key,
			Role:        role,
			Host:        fmt.Sprintf("%s-svc", instance.PodName),
			HTTPPort:    httpPort,
			WSPort:      wsPort,
			P2PPort:     p2pPort,
			Signer:      signer,
			PrivateKey:  privateKey,
			Address:     address,
			Enode:       fmt.Sprintf("enode://%s@%s:%d", nodeID, fmt.Sprintf("%s-svc", instance.PodName), p2pPort),
		}
		identities[instance.Key] = identity
		if identity.Signer {
			signers = append(signers, identity)
		}
	}

	if len(identities) == 0 {
		return nil, nil, fmt.Errorf("no geth runtime instances available")
	}
	if len(signers) == 0 {
		for _, identity := range identities {
			identity.Signer = true
			identities[identity.InstanceKey] = identity
			signers = append(signers, identity)
			break
		}
	}

	return identities, signers, nil
}

func runtimeInstanceImageAndRole(
	blueprint model.ExperimentBlueprint,
	instance model.ExperimentRuntimeTarget,
) (string, string, bool) {
	switch instance.Kind {
	case "workspace":
		return blueprint.Workspace.Image, "workspace", true
	case "node":
		for _, node := range blueprint.Nodes {
			if node.Key == instance.Key {
				return node.Image, node.Role, true
			}
		}
	case "service":
		for _, serviceSpec := range blueprint.Services {
			if serviceSpec.Key == instance.Key {
				return serviceSpec.Image, serviceSpec.Role, true
			}
		}
	case "simulation":
		return resolveRuntimeImage("", "simulation", "visualization"), "simulation", true
	}
	return "", "", false
}

func deriveGethIdentity(seed string) (string, string, string, error) {
	for nonce := 0; nonce < 32; nonce += 1 {
		sum := sha256.Sum256([]byte(fmt.Sprintf("%s:%d", seed, nonce)))
		privateKey, err := crypto.ToECDSA(sum[:])
		if err != nil {
			continue
		}

		privateKeyBytes := crypto.FromECDSA(privateKey)
		publicKeyBytes := crypto.FromECDSAPub(&privateKey.PublicKey)
		address := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()
		nodeID := hex.EncodeToString(publicKeyBytes[1:])
		return hex.EncodeToString(privateKeyBytes), address, nodeID, nil
	}

	return "", "", "", fmt.Errorf("failed to derive deterministic geth identity for %s", seed)
}

func buildCliqueGenesisJSON(chainID string, signers []gethRuntimeIdentity, blockTime string) ([]byte, error) {
	chainIDValue, err := strconv.ParseInt(chainID, 10, 64)
	if err != nil || chainIDValue <= 0 {
		chainIDValue = 31337
	}
	periodValue, err := strconv.ParseInt(strings.TrimSpace(blockTime), 10, 64)
	if err != nil || periodValue <= 0 {
		periodValue = 2
	}

	extraData := strings.Repeat("0", 64)
	for _, signer := range signers {
		extraData += strings.TrimPrefix(strings.ToLower(signer.Address), "0x")
	}
	extraData += strings.Repeat("0", 130)

	alloc := map[string]map[string]string{}
	for _, signer := range signers {
		alloc[signer.Address] = map[string]string{
			"balance": "0x3635C9ADC5DEA00000",
		}
	}

	genesis := map[string]interface{}{
		"config": map[string]interface{}{
			"chainId":             chainIDValue,
			"homesteadBlock":      0,
			"eip150Block":         0,
			"eip155Block":         0,
			"eip158Block":         0,
			"byzantiumBlock":      0,
			"constantinopleBlock": 0,
			"petersburgBlock":     0,
			"istanbulBlock":       0,
			"berlinBlock":         0,
			"londonBlock":         0,
			"clique": map[string]interface{}{
				"period": periodValue,
				"epoch":  30000,
			},
		},
		"nonce":      "0x0",
		"timestamp":  "0x0",
		"extraData":  "0x" + extraData,
		"gasLimit":   "0x1fffffffffffff",
		"difficulty": "0x1",
		"mixHash":    "0x0000000000000000000000000000000000000000000000000000000000000000",
		"coinbase":   "0x0000000000000000000000000000000000000000",
		"alloc":      alloc,
	}

	return json.Marshal(genesis)
}

func resolveGethBlockTime(envVars map[string]string) string {
	if value := strings.TrimSpace(envVars["BLOCK_TIME"]); value != "" {
		return value
	}
	return "2"
}

func preferredPort(ports []int32, candidates ...int32) int32 {
	for _, candidate := range candidates {
		for _, port := range ports {
			if port == candidate {
				return candidate
			}
		}
	}
	return firstNonZeroPort(ports)
}

func matchingPort(ports []int32, candidates ...int32) int32 {
	for _, candidate := range candidates {
		for _, port := range ports {
			if port == candidate {
				return candidate
			}
		}
	}
	return 0
}

func preferredPortOrDefault(ports []int32, fallback int32, candidates ...int32) int32 {
	if port := matchingPort(ports, candidates...); port > 0 {
		return port
	}
	if fallback > 0 {
		return fallback
	}
	return firstNonZeroPort(ports)
}

func ensureRuntimePortsForImage(image string, ports []int32) []int32 {
	result := append([]int32{}, ports...)
	switch resolveImageAlias(image) {
	case "geth":
		result = ensurePortOnRuntimeInstance(result, 30303)
		result = ensurePortOnRuntimeInstance(result, 8546)
		result = ensurePortOnRuntimeInstance(result, 8545)
	case "chainlink":
		result = ensurePortOnRuntimeInstance(result, 6689)
		result = ensurePortOnRuntimeInstance(result, 6688)
	}
	return result
}

func resolveRuntimePortsForImage(image string, ports []int32, tools []string) []int32 {
	return ensureRuntimePortsForImage(image, normalizeRuntimePorts(ports, tools))
}

func firstNonZeroPort(ports []int32) int32 {
	for _, port := range ports {
		if port > 0 {
			return port
		}
	}
	return 0
}
