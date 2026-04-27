package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chainspace/simulations/modules/attacks"
	"github.com/chainspace/simulations/modules/blockchain"
	"github.com/chainspace/simulations/modules/consensus"
	"github.com/chainspace/simulations/modules/crosschain"
	"github.com/chainspace/simulations/modules/crypto"
	"github.com/chainspace/simulations/modules/defi"
	"github.com/chainspace/simulations/modules/evm"
	"github.com/chainspace/simulations/modules/network"
	"github.com/chainspace/simulations/pkg/api"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "启动模拟引擎HTTP/WebSocket服务",
	Long: `启动模拟引擎的HTTP和WebSocket服务器。

服务启动后，前端可以通过以下方式交互：
  - HTTP API: GET/POST /api/* 进行模拟器控制
  - WebSocket: /ws/simulator 进行实时状态推送

示例:
  chainspace-sim serve --port 8080
  chainspace-sim serve --port 8080 --module pbft --nodes 4`,
	RunE: runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().IntP("port", "p", 8080, "服务监听端口")
	serveCmd.Flags().StringP("host", "H", "0.0.0.0", "服务监听地址")
	serveCmd.Flags().StringP("module", "m", "", "初始加载的模块")
	serveCmd.Flags().IntP("nodes", "n", 4, "默认节点数量")
	serveCmd.Flags().StringP("mode", "M", "simulated", "运行模式 (simulated/real)")
	serveCmd.Flags().Bool("cors", true, "启用CORS")
	serveCmd.Flags().Bool("debug", false, "启用调试模式")

	viper.BindPFlag("server.port", serveCmd.Flags().Lookup("port"))
	viper.BindPFlag("server.host", serveCmd.Flags().Lookup("host"))
	viper.BindPFlag("server.cors", serveCmd.Flags().Lookup("cors"))
	viper.BindPFlag("server.debug", serveCmd.Flags().Lookup("debug"))
	viper.BindPFlag("simulator.module", serveCmd.Flags().Lookup("module"))
	viper.BindPFlag("simulator.nodes", serveCmd.Flags().Lookup("nodes"))
	viper.BindPFlag("simulator.mode", serveCmd.Flags().Lookup("mode"))
}

func runServe(cmd *cobra.Command, args []string) error {
	port := viper.GetInt("server.port")
	host := viper.GetString("server.host")
	enableCORS := viper.GetBool("server.cors")
	debug := viper.GetBool("server.debug")
	module := viper.GetString("simulator.module")
	nodes := viper.GetInt("simulator.nodes")
	mode := viper.GetString("simulator.mode")

	// 创建引擎
	eng := engine.NewEngine()

	// 注册所有模块
	registerAllModules(eng)

	// 如果指定了模块，初始化它
	if module != "" {
		config := types.Config{
			Module:    module,
			NodeCount: nodes,
			Mode:      types.RunMode(mode),
			Params:    make(map[string]interface{}),
		}
		if err := eng.Init(config); err != nil {
			return fmt.Errorf("初始化模块失败: %w", err)
		}
		fmt.Printf("已加载模块: %s (节点数: %d, 模式: %s)\n", module, nodes, mode)
	}

	// 创建服务器
	serverConfig := api.ServerConfig{
		Port:       port,
		Host:       host,
		EnableCORS: enableCORS,
		Debug:      debug,
	}
	server := api.NewServer(eng, serverConfig)

	// 启动服务器
	go func() {
		fmt.Printf("ChainSpace Simulation Engine 启动中...\n")
		fmt.Printf("HTTP API: http://%s:%d/api\n", host, port)
		fmt.Printf("WebSocket: ws://%s:%d/ws/simulator\n", host, port)
		fmt.Printf("健康检查: http://%s:%d/health\n", host, port)
		if err := server.Start(); err != nil {
			fmt.Printf("服务器错误: %v\n", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\n正在关闭服务器...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Stop(ctx); err != nil {
		return fmt.Errorf("关闭服务器失败: %w", err)
	}

	if err := eng.Stop(); err != nil {
		return fmt.Errorf("关闭引擎失败: %w", err)
	}

	fmt.Println("服务器已关闭")
	return nil
}

func registerAllModules(eng *engine.Engine) {
	// 注册所有功能模块到引擎
	// 共9大类，113个模块

	// P1 - 基础功能模块
	blockchain.RegisterAll(eng.Registry()) // 12个区块链基础模块
	consensus.RegisterAll(eng.Registry())  // 10个共识算法模块
	crypto.RegisterAll(eng.Registry())     // 15个密码学模块

	// P2 - 攻击与EVM
	evm.RegisterAll(eng.Registry())     // 11个EVM模块
	attacks.RegisterAll(eng.Registry()) // 24个攻击演示模块

	// P3 - 高级功能
	network.RegisterAll(eng.Registry())    // 10个P2P网络模块
	defi.RegisterAll(eng.Registry())       // 13个DeFi机制模块
	crosschain.RegisterAll(eng.Registry()) // 12个跨链/L2模块
	// evaluation模块已移至后端，不在simulations中
}
