package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
)

var rootCmd = &cobra.Command{
	Use:   "chainspace-sim",
	Short: "ChainSpace Simulation Engine",
	Long: `ChainSpace Simulation Engine - 区块链可视化实验的算法实现引擎

该程序提供区块链共识算法、攻击演示、DeFi机制等的真实实现，
运行在学生实验容器内，通过HTTP/WebSocket API与前端交互。

支持两种运行模式：
  - simulated: 单进程内存队列模式，适合单人实验
  - real: 真实TCP/gRPC通信模式，适合多人协作实验`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.chainspace-sim.yaml)")
	rootCmd.PersistentFlags().String("log-level", "info", "log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().String("log-format", "json", "log format (json, text)")

	viper.BindPFlag("log.level", rootCmd.PersistentFlags().Lookup("log-level"))
	viper.BindPFlag("log.format", rootCmd.PersistentFlags().Lookup("log-format"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".chainspace-sim")
	}

	viper.AutomaticEnv()
	viper.SetEnvPrefix("CHAINSPACE")

	viper.ReadInConfig()
}
