package blockchain

import "github.com/chainspace/simulations/pkg/engine"

// RegisterAll 注册所有区块链基础模块 (12个)
func RegisterAll(registry *engine.Registry) {
	// 数据结构
	registry.Register("block_structure", NewBlockStructureFactory())
	registry.Register("merkle_tree", NewMerkleTreeFactory())
	registry.Register("state_trie", NewStateTrieFactory())

	// 交易模型
	registry.Register("transaction", NewTransactionFactory())
	registry.Register("utxo", NewUTXOFactory())
	registry.Register("mempool", NewMempoolFactory())

	// 挖矿与同步
	registry.Register("mining", NewMiningFactory())
	registry.Register("difficulty", NewDifficultyFactory())
	registry.Register("chain_sync", NewChainSyncFactory())

	// 钱包与客户端
	registry.Register("wallet", NewWalletFactory())
	registry.Register("light_client", NewLightClientFactory())
	registry.Register("gas", NewGasFactory())
}

// RegisterToEngine 注册到引擎 (12个)
func RegisterToEngine(eng *engine.Engine) {
	eng.Register("block_structure", NewBlockStructureFactory())
	eng.Register("merkle_tree", NewMerkleTreeFactory())
	eng.Register("state_trie", NewStateTrieFactory())
	eng.Register("transaction", NewTransactionFactory())
	eng.Register("utxo", NewUTXOFactory())
	eng.Register("mempool", NewMempoolFactory())
	eng.Register("mining", NewMiningFactory())
	eng.Register("difficulty", NewDifficultyFactory())
	eng.Register("chain_sync", NewChainSyncFactory())
	eng.Register("wallet", NewWalletFactory())
	eng.Register("light_client", NewLightClientFactory())
	eng.Register("gas", NewGasFactory())
}
