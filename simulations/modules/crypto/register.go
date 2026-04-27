package crypto

import "github.com/chainspace/simulations/pkg/engine"

// RegisterAll 注册所有密码学模块 (15个)
func RegisterAll(registry *engine.Registry) {
	// 基础密码学 (5个)
	registry.Register("hash", NewHashFactory())                    // 哈希算法演示器
	registry.Register("symmetric", NewSymmetricFactory())          // 对称加密演示器
	registry.Register("asymmetric", NewAsymmetricFactory())        // 非对称加密演示器
	registry.Register("elliptic_curve", NewEllipticCurveFactory()) // 椭圆曲线演示器
	registry.Register("signature", NewSignatureFactory())          // 数字签名演示器

	// 高级密码学 (7个)
	registry.Register("zkp", NewZKPFactory())                    // 零知识证明演示器
	registry.Register("mpc", NewMPCFactory())                    // 安全多方计算演示器
	registry.Register("commitment", NewCommitmentFactory())      // 承诺方案演示器
	registry.Register("merkle_proof", NewMerkleProofFactory())   // Merkle证明演示器
	registry.Register("threshold_sig", NewThresholdSigFactory()) // 门限签名演示器
	registry.Register("ring_sig", NewRingSigFactory())           // 环签名演示器
	registry.Register("bls", NewBLSFactory())                    // BLS签名聚合演示器

	// 工具类 (3个)
	registry.Register("kdf", NewKDFFactory())               // 密钥派生函数演示器
	registry.Register("encoding", NewEncodingFactory())     // 编码演示器
	registry.Register("randomness", NewRandomnessFactory()) // 链上随机数演示器
}

// RegisterToEngine 注册到引擎 (15个)
func RegisterToEngine(eng *engine.Engine) {
	// 基础密码学 (5个)
	eng.Register("hash", NewHashFactory())
	eng.Register("symmetric", NewSymmetricFactory())
	eng.Register("asymmetric", NewAsymmetricFactory())
	eng.Register("elliptic_curve", NewEllipticCurveFactory())
	eng.Register("signature", NewSignatureFactory())

	// 高级密码学 (7个)
	eng.Register("zkp", NewZKPFactory())
	eng.Register("mpc", NewMPCFactory())
	eng.Register("commitment", NewCommitmentFactory())
	eng.Register("merkle_proof", NewMerkleProofFactory())
	eng.Register("threshold_sig", NewThresholdSigFactory())
	eng.Register("ring_sig", NewRingSigFactory())
	eng.Register("bls", NewBLSFactory())

	// 工具类 (3个)
	eng.Register("kdf", NewKDFFactory())
	eng.Register("encoding", NewEncodingFactory())
	eng.Register("randomness", NewRandomnessFactory())
}
