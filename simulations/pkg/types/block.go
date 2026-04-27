package types

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"
)

// Block 区块
type Block struct {
	Header       *BlockHeader  `json:"header"`
	Transactions []Transaction `json:"transactions"`
	Hash         Hash          `json:"hash"`
}

// BlockHeader 区块头
type BlockHeader struct {
	Version       uint32    `json:"version"`
	PrevBlockHash Hash      `json:"prev_block_hash"`
	MerkleRoot    Hash      `json:"merkle_root"`
	StateRoot     Hash      `json:"state_root"`
	Timestamp     time.Time `json:"timestamp"`
	Height        uint64    `json:"height"`
	Difficulty    uint64    `json:"difficulty"`
	Nonce         uint64    `json:"nonce"`
	Miner         Address   `json:"miner"`
	GasLimit      uint64    `json:"gas_limit"`
	GasUsed       uint64    `json:"gas_used"`
	ExtraData     []byte    `json:"extra_data"`
}

// Hash 哈希值
type Hash [32]byte

// EmptyHash 空哈希
var EmptyHash = Hash{}

// HashFromHex 从十六进制字符串创建Hash
func HashFromHex(s string) (Hash, error) {
	var h Hash
	b, err := hex.DecodeString(s)
	if err != nil {
		return h, err
	}
	copy(h[:], b)
	return h, nil
}

// String 返回十六进制字符串
func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

// MarshalJSON JSON序列化
func (h Hash) MarshalJSON() ([]byte, error) {
	return json.Marshal(h.String())
}

// UnmarshalJSON JSON反序列化
func (h *Hash) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	decoded, err := hex.DecodeString(s)
	if err != nil {
		return err
	}
	copy(h[:], decoded)
	return nil
}

// IsEmpty 是否为空哈希
func (h Hash) IsEmpty() bool {
	return h == EmptyHash
}

// CalculateHash 计算区块哈希
func (b *Block) CalculateHash() Hash {
	data, _ := json.Marshal(b.Header)
	hash := sha256.Sum256(data)
	return hash
}

// NewBlock 创建新区块
func NewBlock(prevHash Hash, height uint64, txs []Transaction, miner Address) *Block {
	header := &BlockHeader{
		Version:       1,
		PrevBlockHash: prevHash,
		Timestamp:     time.Now(),
		Height:        height,
		Miner:         miner,
		GasLimit:      30000000,
	}

	block := &Block{
		Header:       header,
		Transactions: txs,
	}

	// 计算Merkle根
	block.Header.MerkleRoot = CalculateMerkleRoot(txs)

	// 计算区块哈希
	block.Hash = block.CalculateHash()

	return block
}

// CalculateMerkleRoot 计算Merkle根
func CalculateMerkleRoot(txs []Transaction) Hash {
	if len(txs) == 0 {
		return EmptyHash
	}

	var hashes []Hash
	for _, tx := range txs {
		hashes = append(hashes, tx.Hash)
	}

	for len(hashes) > 1 {
		var newLevel []Hash
		for i := 0; i < len(hashes); i += 2 {
			var combined []byte
			combined = append(combined, hashes[i][:]...)
			if i+1 < len(hashes) {
				combined = append(combined, hashes[i+1][:]...)
			} else {
				combined = append(combined, hashes[i][:]...)
			}
			newLevel = append(newLevel, sha256.Sum256(combined))
		}
		hashes = newLevel
	}

	return hashes[0]
}

// GenesisBlock 创世区块
func GenesisBlock() *Block {
	header := &BlockHeader{
		Version:       1,
		PrevBlockHash: EmptyHash,
		Timestamp:     time.Unix(0, 0),
		Height:        0,
		Difficulty:    1,
		GasLimit:      30000000,
	}

	block := &Block{
		Header:       header,
		Transactions: nil,
	}
	block.Header.MerkleRoot = EmptyHash
	block.Hash = block.CalculateHash()

	return block
}

// Chain 链
type Chain struct {
	Blocks      []*Block        `json:"blocks"`
	BlockByHash map[Hash]*Block `json:"-"`
	Height      uint64          `json:"height"`
}

// NewChain 创建新链
func NewChain() *Chain {
	genesis := GenesisBlock()
	return &Chain{
		Blocks:      []*Block{genesis},
		BlockByHash: map[Hash]*Block{genesis.Hash: genesis},
		Height:      0,
	}
}

// AddBlock 添加区块
func (c *Chain) AddBlock(block *Block) error {
	c.Blocks = append(c.Blocks, block)
	c.BlockByHash[block.Hash] = block
	c.Height = block.Header.Height
	return nil
}

// GetBlock 获取区块
func (c *Chain) GetBlock(hash Hash) *Block {
	return c.BlockByHash[hash]
}

// GetBlockByHeight 按高度获取区块
func (c *Chain) GetBlockByHeight(height uint64) *Block {
	if height > c.Height {
		return nil
	}
	return c.Blocks[height]
}

// LatestBlock 获取最新区块
func (c *Chain) LatestBlock() *Block {
	if len(c.Blocks) == 0 {
		return nil
	}
	return c.Blocks[len(c.Blocks)-1]
}
