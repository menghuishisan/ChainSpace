package consensus

import "github.com/chainspace/simulations/pkg/engine"

// RegisterAll 注册所有共识算法模块 (10个)
func RegisterAll(registry *engine.Registry) {
	// 基础共识
	registry.Register("pow", NewPoWFactory())
	registry.Register("pos", NewPoSFactory())
	registry.Register("dpos", NewDPoSFactory())

	// BFT类共识
	registry.Register("pbft", NewPBFTFactory())
	registry.Register("raft", NewRaftFactory())
	registry.Register("hotstuff", NewHotStuffFactory())
	registry.Register("tendermint", NewTendermintFactory())

	// 高级共识
	registry.Register("dag", NewDAGFactory())
	registry.Register("vrf", NewVRFFactory())
	registry.Register("fork_choice", NewForkChoiceFactory())
}

// RegisterToEngine 注册到引擎 (10个)
func RegisterToEngine(eng *engine.Engine) {
	// 基础共识
	eng.Register("pow", NewPoWFactory())
	eng.Register("pos", NewPoSFactory())
	eng.Register("dpos", NewDPoSFactory())

	// BFT类共识
	eng.Register("pbft", NewPBFTFactory())
	eng.Register("raft", NewRaftFactory())
	eng.Register("hotstuff", NewHotStuffFactory())
	eng.Register("tendermint", NewTendermintFactory())

	// 高级共识
	eng.Register("dag", NewDAGFactory())
	eng.Register("vrf", NewVRFFactory())
	eng.Register("fork_choice", NewForkChoiceFactory())
}
