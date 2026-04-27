package crypto

import (
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// =============================================================================
// 椭圆曲线数据结构
// =============================================================================

// ECPoint 椭圆曲线上的点
type ECPoint struct {
	X    *big.Int `json:"-"`
	Y    *big.Int `json:"-"`
	XHex string   `json:"x"`
	YHex string   `json:"y"`
}

// ECScalar 标量值
type ECScalar struct {
	Value    *big.Int `json:"-"`
	ValueHex string   `json:"value"`
}

// ECOperation 椭圆曲线操作记录
type ECOperation struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"` // add/double/multiply/verify
	Input1    string    `json:"input1"`
	Input2    string    `json:"input2"`
	Result    string    `json:"result"`
	Timestamp time.Time `json:"timestamp"`
}

// =============================================================================
// EllipticCurveSimulator 椭圆曲线演示器
// =============================================================================

// EllipticCurveSimulator 椭圆曲线演示器
// 演示区块链中使用的椭圆曲线密码学
//
// 核心概念:
// 1. 椭圆曲线方程: y² = x³ + ax + b (mod p)
// 2. 点加法: P + Q = R
// 3. 点倍乘: k * P (标量乘法)
// 4. 离散对数问题: 已知P和Q=kP，求k是困难的
//
// secp256k1 (比特币/以太坊使用):
// - p = 2^256 - 2^32 - 977
// - a = 0, b = 7
// - G = 生成元
// - n = 曲线阶
//
// P-256 (NIST标准):
// - 更广泛的应用场景
// - TLS/SSL证书常用
type EllipticCurveSimulator struct {
	*base.BaseSimulator
	mu        sync.RWMutex
	curve     elliptic.Curve
	curveName string
	G         *ECPoint // 生成元
	N         *big.Int // 曲线阶
	points    map[string]*ECPoint
	history   []*ECOperation
}

// NewEllipticCurveSimulator 创建椭圆曲线演示器
func NewEllipticCurveSimulator() *EllipticCurveSimulator {
	sim := &EllipticCurveSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"elliptic_curve",
			"椭圆曲线演示器",
			"演示secp256k1/P-256曲线运算，包括点加法、标量乘法和ECDH",
			"crypto",
			types.ComponentTool,
		),
		points:  make(map[string]*ECPoint),
		history: make([]*ECOperation, 0),
	}

	sim.AddParam(types.Param{
		Key:         "curve",
		Name:        "曲线类型",
		Description: "选择椭圆曲线",
		Type:        types.ParamTypeSelect,
		Default:     "P-256",
		Options: []types.Option{
			{Label: "P-256 (NIST)", Value: "P-256"},
			{Label: "P-384", Value: "P-384"},
			{Label: "P-521", Value: "P-521"},
		},
	})

	return sim
}

// Init 初始化演示器
func (s *EllipticCurveSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	// 选择曲线
	s.curveName = "P-256"
	if v, ok := config.Params["curve"]; ok {
		if str, ok := v.(string); ok {
			s.curveName = str
		}
	}

	switch s.curveName {
	case "P-384":
		s.curve = elliptic.P384()
	case "P-521":
		s.curve = elliptic.P521()
	default:
		s.curve = elliptic.P256()
		s.curveName = "P-256"
	}

	// 设置曲线参数
	params := s.curve.Params()
	s.G = &ECPoint{
		X:    params.Gx,
		Y:    params.Gy,
		XHex: hex.EncodeToString(params.Gx.Bytes()),
		YHex: hex.EncodeToString(params.Gy.Bytes()),
	}
	s.N = params.N
	s.points["G"] = s.G

	s.updateState()
	return nil
}

// =============================================================================
// 椭圆曲线运算
// =============================================================================

// GeneratePoint 生成随机点 (k * G)
func (s *EllipticCurveSimulator) GeneratePoint(name string) (*ECPoint, *ECScalar, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 生成随机标量 k
	k, err := rand.Int(rand.Reader, s.N)
	if err != nil {
		return nil, nil, fmt.Errorf("生成随机数失败: %v", err)
	}

	// 计算 P = k * G
	x, y := s.curve.ScalarBaseMult(k.Bytes())

	point := &ECPoint{
		X:    x,
		Y:    y,
		XHex: hex.EncodeToString(x.Bytes()),
		YHex: hex.EncodeToString(y.Bytes()),
	}

	scalar := &ECScalar{
		Value:    k,
		ValueHex: hex.EncodeToString(k.Bytes()),
	}

	s.points[name] = point

	s.history = append(s.history, &ECOperation{
		ID:        fmt.Sprintf("op-%d", len(s.history)+1),
		Type:      "generate",
		Input1:    "k (random)",
		Input2:    "G",
		Result:    name,
		Timestamp: time.Now(),
	})

	s.EmitEvent("point_generated", "", "", map[string]interface{}{
		"name": name,
		"x":    point.XHex[:16] + "...",
		"y":    point.YHex[:16] + "...",
	})

	s.updateState()
	return point, scalar, nil
}

// PointAdd 点加法 P + Q = R
//
// 几何意义: 过P和Q的直线与曲线的第三个交点关于x轴的对称点
// 代数公式:
// λ = (y2 - y1) / (x2 - x1) mod p
// x3 = λ² - x1 - x2 mod p
// y3 = λ(x1 - x3) - y1 mod p
func (s *EllipticCurveSimulator) PointAdd(name1, name2, resultName string) (*ECPoint, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	p1 := s.points[name1]
	p2 := s.points[name2]

	if p1 == nil || p2 == nil {
		return nil, fmt.Errorf("点不存在")
	}

	// 使用曲线的Add方法
	x, y := s.curve.Add(p1.X, p1.Y, p2.X, p2.Y)

	result := &ECPoint{
		X:    x,
		Y:    y,
		XHex: hex.EncodeToString(x.Bytes()),
		YHex: hex.EncodeToString(y.Bytes()),
	}

	s.points[resultName] = result

	s.history = append(s.history, &ECOperation{
		ID:        fmt.Sprintf("op-%d", len(s.history)+1),
		Type:      "add",
		Input1:    name1,
		Input2:    name2,
		Result:    resultName,
		Timestamp: time.Now(),
	})

	s.EmitEvent("point_added", "", "", map[string]interface{}{
		"p1":     name1,
		"p2":     name2,
		"result": resultName,
	})

	s.updateState()
	return result, nil
}

// PointDouble 点倍乘 2P
//
// 几何意义: 过P点的切线与曲线的交点关于x轴的对称点
// 代数公式:
// λ = (3x1² + a) / (2y1) mod p
// x3 = λ² - 2x1 mod p
// y3 = λ(x1 - x3) - y1 mod p
func (s *EllipticCurveSimulator) PointDouble(name, resultName string) (*ECPoint, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	p := s.points[name]
	if p == nil {
		return nil, fmt.Errorf("点不存在: %s", name)
	}

	// 2P = P + P
	x, y := s.curve.Double(p.X, p.Y)

	result := &ECPoint{
		X:    x,
		Y:    y,
		XHex: hex.EncodeToString(x.Bytes()),
		YHex: hex.EncodeToString(y.Bytes()),
	}

	s.points[resultName] = result

	s.history = append(s.history, &ECOperation{
		ID:        fmt.Sprintf("op-%d", len(s.history)+1),
		Type:      "double",
		Input1:    name,
		Input2:    "",
		Result:    resultName,
		Timestamp: time.Now(),
	})

	s.EmitEvent("point_doubled", "", "", map[string]interface{}{
		"input":  name,
		"result": resultName,
	})

	s.updateState()
	return result, nil
}

// ScalarMult 标量乘法 k * P
//
// 使用快速幂算法 (double-and-add):
// 将k表示为二进制，从高位到低位:
// - 每一步先倍乘 (double)
// - 如果当前位是1，则加P (add)
func (s *EllipticCurveSimulator) ScalarMult(pointName string, kHex string, resultName string) (*ECPoint, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	p := s.points[pointName]
	if p == nil {
		return nil, fmt.Errorf("点不存在: %s", pointName)
	}

	// 解析标量k
	k := new(big.Int)
	kBytes, err := hex.DecodeString(kHex)
	if err != nil {
		k.SetString(kHex, 10)
	} else {
		k.SetBytes(kBytes)
	}
	k.Mod(k, s.N)

	// k * P
	x, y := s.curve.ScalarMult(p.X, p.Y, k.Bytes())

	result := &ECPoint{
		X:    x,
		Y:    y,
		XHex: hex.EncodeToString(x.Bytes()),
		YHex: hex.EncodeToString(y.Bytes()),
	}

	s.points[resultName] = result

	s.history = append(s.history, &ECOperation{
		ID:        fmt.Sprintf("op-%d", len(s.history)+1),
		Type:      "scalar_mult",
		Input1:    kHex[:min(16, len(kHex))] + "...",
		Input2:    pointName,
		Result:    resultName,
		Timestamp: time.Now(),
	})

	s.EmitEvent("scalar_mult", "", "", map[string]interface{}{
		"point":  pointName,
		"result": resultName,
	})

	s.updateState()
	return result, nil
}

// VerifyOnCurve 验证点是否在曲线上
func (s *EllipticCurveSimulator) VerifyOnCurve(name string) bool {
	s.mu.RLock()
	p := s.points[name]
	s.mu.RUnlock()

	if p == nil {
		return false
	}

	// 验证 y² = x³ + ax + b (mod p)
	valid := s.curve.IsOnCurve(p.X, p.Y)

	s.EmitEvent("point_verified", "", "", map[string]interface{}{
		"name":     name,
		"on_curve": valid,
	})

	return valid
}

// DemonstrateECDH 演示ECDH密钥交换
//
// 流程:
// 1. Alice生成私钥a，计算公钥A = a*G
// 2. Bob生成私钥b，计算公钥B = b*G
// 3. Alice计算共享密钥: S = a*B = a*b*G
// 4. Bob计算共享密钥: S = b*A = b*a*G
// 5. 两者得到相同的点S
func (s *EllipticCurveSimulator) DemonstrateECDH() map[string]interface{} {
	// Alice的密钥对
	alicePriv, _ := rand.Int(rand.Reader, s.N)
	alicePubX, alicePubY := s.curve.ScalarBaseMult(alicePriv.Bytes())

	// Bob的密钥对
	bobPriv, _ := rand.Int(rand.Reader, s.N)
	bobPubX, bobPubY := s.curve.ScalarBaseMult(bobPriv.Bytes())

	// Alice计算共享密钥: a * B
	sharedAliceX, sharedAliceY := s.curve.ScalarMult(bobPubX, bobPubY, alicePriv.Bytes())

	// Bob计算共享密钥: b * A
	sharedBobX, sharedBobY := s.curve.ScalarMult(alicePubX, alicePubY, bobPriv.Bytes())

	// 验证两者相等
	match := sharedAliceX.Cmp(sharedBobX) == 0 && sharedAliceY.Cmp(sharedBobY) == 0

	result := map[string]interface{}{
		"alice_public": map[string]string{
			"x": hex.EncodeToString(alicePubX.Bytes())[:16] + "...",
			"y": hex.EncodeToString(alicePubY.Bytes())[:16] + "...",
		},
		"bob_public": map[string]string{
			"x": hex.EncodeToString(bobPubX.Bytes())[:16] + "...",
			"y": hex.EncodeToString(bobPubY.Bytes())[:16] + "...",
		},
		"shared_secret": map[string]string{
			"x": hex.EncodeToString(sharedAliceX.Bytes())[:16] + "...",
			"y": hex.EncodeToString(sharedAliceY.Bytes())[:16] + "...",
		},
		"secrets_match": match,
	}

	s.EmitEvent("ecdh_demonstrated", "", "", result)

	return result
}

// GetCurveParams 获取曲线参数
func (s *EllipticCurveSimulator) GetCurveParams() map[string]string {
	params := s.curve.Params()
	return map[string]string{
		"name":    params.Name,
		"p":       params.P.Text(16),
		"n":       params.N.Text(16),
		"b":       params.B.Text(16),
		"gx":      params.Gx.Text(16),
		"gy":      params.Gy.Text(16),
		"bitsize": fmt.Sprintf("%d", params.BitSize),
	}
}

// updateState 更新状态
func (s *EllipticCurveSimulator) updateState() {
	pointList := make([]map[string]interface{}, 0)
	for name, p := range s.points {
		pointList = append(pointList, map[string]interface{}{
			"name": name,
			"x":    p.XHex[:min(16, len(p.XHex))] + "...",
			"y":    p.YHex[:min(16, len(p.YHex))] + "...",
		})
	}

	s.SetGlobalData("curve", s.curveName)
	s.SetGlobalData("points", pointList)
	s.SetGlobalData("point_count", len(s.points))
	s.SetGlobalData("operation_count", len(s.history))

	summary := fmt.Sprintf("当前曲线为 %s，已记录 %d 次椭圆曲线运算。", s.curveName, len(s.history))
	nextHint := "可以生成点、执行点加或标量乘，观察椭圆曲线群运算。"
	setCryptoTeachingState(
		s.BaseSimulator,
		"crypto",
		"准备曲线运算",
		summary,
		nextHint,
		0.45,
		map[string]interface{}{"curve": s.curveName, "point_count": len(s.points), "operation_count": len(s.history)},
	)
}

func (s *EllipticCurveSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "generate_point":
		name := fmt.Sprintf("P%d", len(s.points)+1)
		if raw, ok := params["name"].(string); ok && raw != "" {
			name = raw
		}
		point, scalar, err := s.GeneratePoint(name)
		if err != nil {
			return nil, err
		}
		return cryptoActionResult("已生成一个椭圆曲线点。", map[string]interface{}{"name": name, "point": point, "scalar": scalar}, &types.ActionFeedback{
			Summary:     "新的点已经通过随机标量乘以基点生成。",
			NextHint:    "继续执行点加法或标量乘法，观察曲线群运算如何组合。",
			EffectScope: "crypto",
			ResultState: map[string]interface{}{"point_count": len(s.points)},
		}), nil
	case "demo_ecdh":
		result := s.DemonstrateECDH()
		return cryptoActionResult("已演示一次 ECDH 密钥交换。", result, &types.ActionFeedback{
			Summary:     "双方已经各自计算出共享密钥，可观察两边结果如何一致。",
			NextHint:    "重点比较 Alice 和 Bob 的共享结果，理解公钥交换与私钥保密的关系。",
			EffectScope: "crypto",
			ResultState: map[string]interface{}{"curve": s.curveName, "secrets_match": result["secrets_match"]},
		}), nil
	default:
		return nil, fmt.Errorf("unsupported elliptic curve action: %s", action)
	}
}

// =============================================================================
// 工厂
// =============================================================================

// EllipticCurveFactory 椭圆曲线演示器工厂
type EllipticCurveFactory struct{}

// Create 创建演示器实例
func (f *EllipticCurveFactory) Create() engine.Simulator {
	return NewEllipticCurveSimulator()
}

// GetDescription 获取描述
func (f *EllipticCurveFactory) GetDescription() types.Description {
	return NewEllipticCurveSimulator().GetDescription()
}

// NewEllipticCurveFactory 创建工厂实例
func NewEllipticCurveFactory() *EllipticCurveFactory {
	return &EllipticCurveFactory{}
}

var _ engine.SimulatorFactory = (*EllipticCurveFactory)(nil)
