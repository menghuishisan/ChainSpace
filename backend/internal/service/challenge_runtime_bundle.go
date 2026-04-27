package service

import (
	"fmt"
	"hash/fnv"
	"sort"
	"strconv"
	"strings"

	"github.com/chainspace/backend/internal/model"
)

const (
	challengeRuntimeComponentRoleMain     = "main"
	challengeRuntimeComponentRolePostgres = "postgres"
	challengeRuntimeComponentRoleIPFS     = "ipfs"
)

type challengeRuntimeComponent struct {
	RuntimeKey    string
	Image         string
	Ports         []model.ChallengeServicePort
	Role          string
	StudentFacing bool
}

type challengeRuntimeBundle struct {
	Service       model.ChallengeServiceSpec
	Components    []challengeRuntimeComponent
	ToolProviders map[string]string
	PrimaryKey    string
}

func buildChallengeRuntimeBundles(orch model.ChallengeOrchestration) []challengeRuntimeBundle {
	bundles := make([]challengeRuntimeBundle, 0, len(orch.Services))
	for _, rawSpec := range orch.Services {
		serviceSpec := normalizeChallengeServiceSpec(rawSpec)
		if serviceSpec.Key == "" || serviceSpec.Key == "anvil_fork" || serviceSpec.Key == "fork" || serviceSpec.Key == "fork-node" {
			continue
		}

		bundle := challengeRuntimeBundle{
			Service:       serviceSpec,
			PrimaryKey:    serviceSpec.Key,
			ToolProviders: map[string]string{},
			Components: []challengeRuntimeComponent{{
				RuntimeKey:    serviceSpec.Key,
				Image:         defaultChallengeServiceImage(serviceSpec),
				Ports:         append([]model.ChallengeServicePort{}, serviceSpec.Ports...),
				Role:          challengeRuntimeComponentRoleMain,
				StudentFacing: true,
			}},
		}

		for _, portSpec := range serviceSpec.Ports {
			toolKey := strings.TrimSpace(portSpec.ExposeAs)
			if toolKey == "" {
				continue
			}
			if _, exists := bundle.ToolProviders[toolKey]; !exists {
				bundle.ToolProviders[toolKey] = bundle.PrimaryKey
			}
		}

		switch serviceSpec.Key {
		case "blockscout", "chainlink":
			bundle.Components = append(bundle.Components, challengeRuntimeComponent{
				RuntimeKey:    serviceSpec.Key + "-db",
				Image:         "postgres:16-alpine",
				Ports:         []model.ChallengeServicePort{{Name: "postgres", Port: 5432, Protocol: "tcp"}},
				Role:          challengeRuntimeComponentRolePostgres,
				StudentFacing: false,
			})
		case "thegraph":
			bundle.Components = append(bundle.Components,
				challengeRuntimeComponent{
					RuntimeKey:    serviceSpec.Key + "-db",
					Image:         "postgres:16-alpine",
					Ports:         []model.ChallengeServicePort{{Name: "postgres", Port: 5432, Protocol: "tcp"}},
					Role:          challengeRuntimeComponentRolePostgres,
					StudentFacing: false,
				},
				challengeRuntimeComponent{
					RuntimeKey: serviceSpec.Key + "-ipfs",
					Image:      resolveRuntimeImage("", "ipfs"),
					Ports: []model.ChallengeServicePort{
						{Name: "api", Port: 5001, Protocol: "http"},
						{Name: "gateway", Port: 8080, Protocol: "http"},
						{Name: "swarm", Port: 4001, Protocol: "tcp"},
					},
					Role:          challengeRuntimeComponentRoleIPFS,
					StudentFacing: false,
				},
			)
		}

		bundles = append(bundles, bundle)
	}
	return bundles
}

func challengeBundleByLogicalKey(bundles []challengeRuntimeBundle, serviceKey string) (challengeRuntimeBundle, bool) {
	for _, bundle := range bundles {
		if bundle.Service.Key == serviceKey {
			return bundle, true
		}
	}
	return challengeRuntimeBundle{}, false
}

func challengeBundleComponentByKey(bundle challengeRuntimeBundle, runtimeKey string) (challengeRuntimeComponent, bool) {
	for _, component := range bundle.Components {
		if component.RuntimeKey == runtimeKey {
			return component, true
		}
	}
	return challengeRuntimeComponent{}, false
}

func challengeBundleComponentEnvID(envID, runtimeKey string) string {
	return challengeServicePodName(envID, runtimeKey)
}

func challengeBundleComponentHost(namespace, envID, runtimeKey string) string {
	return fmt.Sprintf("%s-svc.%s.svc.cluster.local", challengeBundleComponentEnvID(envID, runtimeKey), namespace)
}

func challengeBundleComponentURL(namespace, envID, runtimeKey string, port int, scheme string) string {
	if port <= 0 {
		return ""
	}
	if strings.TrimSpace(scheme) == "" {
		scheme = "http"
	}
	return fmt.Sprintf("%s://%s:%d", scheme, challengeBundleComponentHost(namespace, envID, runtimeKey), port)
}

func challengeComponentPorts(component challengeRuntimeComponent) []int32 {
	seen := map[int32]struct{}{}
	ports := make([]int32, 0, len(component.Ports))
	for _, portSpec := range component.Ports {
		if portSpec.Port <= 0 {
			continue
		}
		port := int32(portSpec.Port)
		if _, exists := seen[port]; exists {
			continue
		}
		seen[port] = struct{}{}
		ports = append(ports, port)
	}
	return ensureRuntimePortsForImage(component.Image, ports)
}

func challengeComponentExposedTools(component challengeRuntimeComponent) []string {
	seen := map[string]struct{}{}
	tools := make([]string, 0, len(component.Ports))
	for _, portSpec := range component.Ports {
		toolKey := strings.TrimSpace(portSpec.ExposeAs)
		if toolKey == "" {
			continue
		}
		if _, exists := seen[toolKey]; exists {
			continue
		}
		seen[toolKey] = struct{}{}
		tools = append(tools, toolKey)
	}
	sort.Strings(tools)
	return tools
}

func challengePortByExposeAs(component challengeRuntimeComponent, exposeAs string, preferredProtocols ...string) int {
	for _, protocol := range preferredProtocols {
		for _, portSpec := range component.Ports {
			if portSpec.Port > 0 &&
				strings.EqualFold(strings.TrimSpace(portSpec.ExposeAs), strings.TrimSpace(exposeAs)) &&
				strings.EqualFold(strings.TrimSpace(portSpec.Protocol), strings.TrimSpace(protocol)) {
				return portSpec.Port
			}
		}
	}
	for _, portSpec := range component.Ports {
		if portSpec.Port > 0 && strings.EqualFold(strings.TrimSpace(portSpec.ExposeAs), strings.TrimSpace(exposeAs)) {
			return portSpec.Port
		}
	}
	return 0
}

func challengePortByName(component challengeRuntimeComponent, names ...string) int {
	for _, name := range names {
		for _, portSpec := range component.Ports {
			if portSpec.Port > 0 && strings.EqualFold(strings.TrimSpace(portSpec.Name), strings.TrimSpace(name)) {
				return portSpec.Port
			}
		}
	}
	return 0
}

func challengeWorkspaceProbePort(tools []string) int32 {
	for _, tool := range tools {
		switch strings.TrimSpace(tool) {
		case "terminal", "network":
			return 7681
		case "ide":
			return 8443
		}
	}
	return 0
}

func challengeStringEnvValue(env model.JSONMap, keys ...string) string {
	for _, key := range keys {
		value, ok := env[key]
		if !ok || value == nil {
			continue
		}
		text := strings.TrimSpace(fmt.Sprintf("%v", value))
		if text != "" && text != "<nil>" {
			return text
		}
	}
	return ""
}

func resolveChallengeChainID(orch model.ChallengeOrchestration, serviceSpec model.ChallengeServiceSpec, input ChallengePodInput) string {
	if value := challengeStringEnvValue(serviceSpec.Env, "CHAIN_ID", "CHAINSPACE_GETH_CHAIN_ID", "NETWORK_ID"); value != "" {
		return value
	}
	if orch.Fork.Enabled && orch.Fork.ChainID > 0 {
		return strconv.Itoa(orch.Fork.ChainID)
	}
	hasher := fnv.New32a()
	_, _ = hasher.Write([]byte(input.EnvID + ":" + serviceSpec.Key))
	return strconv.Itoa(30000 + int(hasher.Sum32()%10000))
}

func resolveChallengeBlockTime(serviceSpec model.ChallengeServiceSpec) string {
	if value := challengeStringEnvValue(serviceSpec.Env, "BLOCK_TIME", "CHAINSPACE_GETH_BLOCK_TIME"); value != "" {
		return value
	}
	return "2"
}

func resolveChallengeRPCHTTPURL(
	orch model.ChallengeOrchestration,
	input ChallengePodInput,
	bundles []challengeRuntimeBundle,
	namespace string,
	current model.ChallengeServiceSpec,
) string {
	if value := challengeStringEnvValue(current.Env, "CHAIN_RPC_URL", "RPC_URL", "ETH_RPC_URL", "ETHEREUM_JSONRPC_HTTP_URL"); value != "" {
		return value
	}
	if orch.Fork.Enabled {
		return fmt.Sprintf("http://fork-%s.%s.svc.cluster.local:8545", input.EnvID, namespace)
	}
	for _, bundle := range bundles {
		runtimeKey, ok := bundle.ToolProviders["rpc"]
		if !ok {
			continue
		}
		component, exists := challengeBundleComponentByKey(bundle, runtimeKey)
		if !exists {
			continue
		}
		port := challengePortByExposeAs(component, "rpc", "http")
		if port <= 0 {
			port = challengePortByName(component, "rpc")
		}
		if port <= 0 {
			continue
		}
		return challengeBundleComponentURL(namespace, input.EnvID, runtimeKey, port, "http")
	}
	return ""
}

func resolveChallengeRPCWSURL(
	orch model.ChallengeOrchestration,
	input ChallengePodInput,
	bundles []challengeRuntimeBundle,
	namespace string,
	current model.ChallengeServiceSpec,
) string {
	if value := challengeStringEnvValue(current.Env, "CHAIN_WS_URL", "WS_RPC_URL", "ETH_WS_URL", "ETHEREUM_JSONRPC_WS_URL"); value != "" {
		return value
	}
	if orch.Fork.Enabled {
		return fmt.Sprintf("ws://fork-%s.%s.svc.cluster.local:8546", input.EnvID, namespace)
	}
	for _, bundle := range bundles {
		runtimeKey, ok := bundle.ToolProviders["rpc"]
		if !ok {
			continue
		}
		component, exists := challengeBundleComponentByKey(bundle, runtimeKey)
		if !exists {
			continue
		}
		port := challengePortByExposeAs(component, "rpc", "ws")
		if port <= 0 {
			port = challengePortByName(component, "ws")
		}
		if port <= 0 {
			continue
		}
		return challengeBundleComponentURL(namespace, input.EnvID, runtimeKey, port, "ws")
	}
	return ""
}

func challengeRuntimeCommandForComponent(component challengeRuntimeComponent) []string {
	if resolveImageAlias(component.Image) == "geth" {
		return []string{"chainspace-geth-runtime"}
	}
	return nil
}

func buildChallengeRuntimeEnvVars(
	orch model.ChallengeOrchestration,
	input ChallengePodInput,
	bundles []challengeRuntimeBundle,
	bundle challengeRuntimeBundle,
	component challengeRuntimeComponent,
	namespace string,
) (map[string]string, []string, int32) {
	envVars := map[string]string{
		"CHALLENGE_MODE":              orch.Mode,
		"CHALLENGE_SERVICE_KEY":       bundle.Service.Key,
		"CHALLENGE_SERVICE_COMPONENT": component.RuntimeKey,
		"SERVICE_KEY":                 bundle.Service.Key,
		"SERVICE_PURPOSE":             bundle.Service.Purpose,
		"CHAINSPACE_RUNTIME_KIND":     "service",
	}

	for key, value := range bundle.Service.Env {
		if value == nil {
			continue
		}
		envVars[key] = fmt.Sprintf("%v", value)
	}

	if tools := challengeComponentExposedTools(component); len(tools) > 0 {
		envVars["CHAINSPACE_INTERACTION_TOOLS"] = strings.Join(tools, ",")
	}

	ports := challengeComponentPorts(component)
	probePort := firstNonZeroPort(ports)

	switch component.Role {
	case challengeRuntimeComponentRolePostgres:
		dbName := sanitizeChallengeKey(bundle.Service.Key)
		if dbName == "" {
			dbName = "chainspace"
		}
		password := dbName + "-chainspace"
		envVars["POSTGRES_DB"] = dbName
		envVars["POSTGRES_USER"] = dbName
		envVars["POSTGRES_PASSWORD"] = password
		probePort = 5432
	case challengeRuntimeComponentRoleIPFS:
		probePort = preferredPort(ports, 5001, 8080, 4001)
	default:
		switch resolveImageAlias(component.Image) {
		case "geth":
			httpPort := challengePortByExposeAs(component, "rpc", "http")
			if httpPort <= 0 {
				httpPort = challengePortByName(component, "rpc")
			}
			if httpPort <= 0 {
				httpPort = 8545
			}
			wsPort := challengePortByExposeAs(component, "rpc", "ws")
			if wsPort <= 0 {
				wsPort = challengePortByName(component, "ws")
			}
			if wsPort <= 0 {
				wsPort = 8546
			}
			p2pPort := challengePortByName(component, "p2p")
			if p2pPort <= 0 {
				p2pPort = 30303
			}
			envVars["CHAINSPACE_GETH_MODE"] = "single"
			envVars["CHAINSPACE_GETH_CHAIN_ID"] = resolveChallengeChainID(orch, bundle.Service, input)
			envVars["CHAINSPACE_GETH_BLOCK_TIME"] = resolveChallengeBlockTime(bundle.Service)
			envVars["CHAINSPACE_GETH_HTTP_PORT"] = strconv.Itoa(httpPort)
			envVars["CHAINSPACE_GETH_WS_PORT"] = strconv.Itoa(wsPort)
			envVars["CHAINSPACE_GETH_P2P_PORT"] = strconv.Itoa(p2pPort)
			probePort = int32(httpPort)
		case "blockscout":
			dbName := sanitizeChallengeKey(bundle.Service.Key)
			if dbName == "" {
				dbName = "blockscout"
			}
			dbPassword := dbName + "-chainspace"
			dbHost := challengeBundleComponentHost(namespace, input.EnvID, bundle.Service.Key+"-db")
			httpRPC := resolveChallengeRPCHTTPURL(orch, input, bundles, namespace, bundle.Service)
			wsRPC := resolveChallengeRPCWSURL(orch, input, bundles, namespace, bundle.Service)
			envVars["DATABASE_URL"] = fmt.Sprintf("postgresql://%s:%s@%s:5432/%s", dbName, dbPassword, dbHost, dbName)
			envVars["ECTO_USE_SSL"] = "false"
			envVars["SECRET_KEY_BASE"] = "chainspace-blockscout-secret-key-base"
			envVars["ETHEREUM_JSONRPC_VARIANT"] = "geth"
			envVars["PORT"] = strconv.Itoa(maxInt(challengePortByExposeAs(component, "explorer", "http"), 4000))
			if httpRPC != "" {
				envVars["ETHEREUM_JSONRPC_HTTP_URL"] = httpRPC
				envVars["ETHEREUM_JSONRPC_TRACE_URL"] = httpRPC
			}
			if wsRPC != "" {
				envVars["ETHEREUM_JSONRPC_WS_URL"] = wsRPC
			}
			envVars["NETWORK"] = "ChainSpace"
			envVars["SUBNETWORK"] = "Challenge"
			envVars["COIN"] = "ETH"
			probePort = int32(maxInt(challengePortByExposeAs(component, "explorer", "http"), 4000))
		case "chainlink":
			dbName := sanitizeChallengeKey(bundle.Service.Key)
			if dbName == "" {
				dbName = "chainlink"
			}
			dbPassword := dbName + "-chainspace"
			dbHost := challengeBundleComponentHost(namespace, input.EnvID, bundle.Service.Key+"-db")
			httpRPC := resolveChallengeRPCHTTPURL(orch, input, bundles, namespace, bundle.Service)
			wsRPC := resolveChallengeRPCWSURL(orch, input, bundles, namespace, bundle.Service)
			envVars["CHAINLINK_DATABASE_URL"] = fmt.Sprintf("postgresql://%s:%s@%s:5432/%s?sslmode=disable", dbName, dbPassword, dbHost, dbName)
			envVars["CHAINLINK_HTTP_PORT"] = strconv.Itoa(maxInt(challengePortByExposeAs(component, "api_debug", "http"), 6688))
			envVars["CHAINLINK_P2P_PORT"] = strconv.Itoa(maxInt(challengePortByName(component, "p2p"), 6689))
			envVars["CHAINLINK_EVM_CHAIN_ID"] = resolveChallengeChainID(orch, bundle.Service, input)
			envVars["CHAINLINK_API_EMAIL"] = "admin@chainspace.local"
			envVars["CHAINLINK_API_PASSWORD"] = "ChainspaceAdmin123!"
			envVars["CHAINLINK_KEYSTORE_PASSWORD"] = "ChainspaceKeystore123!"
			if httpRPC != "" {
				envVars["CHAINLINK_EVM_HTTP_URL"] = httpRPC
			}
			if wsRPC != "" {
				envVars["CHAINLINK_EVM_WS_URL"] = wsRPC
			}
			probePort = int32(maxInt(challengePortByExposeAs(component, "api_debug", "http"), 6688))
		case "thegraph":
			dbName := sanitizeChallengeKey(bundle.Service.Key)
			if dbName == "" {
				dbName = "graph"
			}
			dbPassword := dbName + "-chainspace"
			envVars["postgres_host"] = challengeBundleComponentHost(namespace, input.EnvID, bundle.Service.Key+"-db")
			envVars["postgres_port"] = "5432"
			envVars["postgres_user"] = dbName
			envVars["postgres_pass"] = dbPassword
			envVars["postgres_db"] = dbName
			envVars["ipfs"] = challengeBundleComponentURL(namespace, input.EnvID, bundle.Service.Key+"-ipfs", 5001, "http")
			if rpcURL := resolveChallengeRPCHTTPURL(orch, input, bundles, namespace, bundle.Service); rpcURL != "" {
				envVars["ethereum"] = "mainnet:" + rpcURL
			}
			envVars["GRAPH_LOG"] = "info"
			probePort = int32(maxInt(challengePortByExposeAs(component, "api_debug", "http"), 8000))
		case "ipfs":
			probePort = preferredPort(ports, 5001, 8080, 4001)
		case "bitcoin":
			probePort = preferredPort(ports, 8332, 18443)
		case "solana":
			probePort = preferredPort(ports, 8899)
		case "substrate":
			probePort = preferredPort(ports, 9944, 9933)
		}
	}

	if probePort > 0 {
		envVars["CHAINSPACE_HEALTHCHECK_PORT"] = strconv.Itoa(int(probePort))
	}
	return envVars, challengeRuntimeCommandForComponent(component), probePort
}

func maxInt(value, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}
