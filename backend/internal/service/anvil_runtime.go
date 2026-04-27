package service

import (
	"fmt"
	"strings"
)

type anvilRuntimeOptions struct {
	ChainID     int
	BlockTime   int
	Accounts    int
	Balance     string
	ForkRPCURL  string
	BlockNumber uint64
}

func buildAnvilRuntimeCommand(options anvilRuntimeOptions) []string {
	args := []string{
		"anvil",
		"--host", "0.0.0.0",
		"--port", "8545",
	}

	if options.ChainID > 0 {
		args = append(args, "--chain-id", fmt.Sprintf("%d", options.ChainID))
	}
	if options.BlockTime > 0 {
		args = append(args, "--block-time", fmt.Sprintf("%d", options.BlockTime))
	}
	if options.Accounts > 0 {
		args = append(args, "--accounts", fmt.Sprintf("%d", options.Accounts))
	}
	if strings.TrimSpace(options.Balance) != "" {
		args = append(args, "--balance", strings.TrimSpace(options.Balance))
	}
	if strings.TrimSpace(options.ForkRPCURL) != "" {
		args = append(args, "--fork-url", strings.TrimSpace(options.ForkRPCURL))
	}
	if options.BlockNumber > 0 {
		args = append(args, "--fork-block-number", fmt.Sprintf("%d", options.BlockNumber))
	}

	command := "export PATH=\"/home/student/.foundry/bin:/root/.foundry/bin:$PATH\" && " + strings.Join(args, " ")
	return []string{"sh", "-lc", command}
}
