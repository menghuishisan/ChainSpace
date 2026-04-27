package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有可用的模拟器模块",
	Long: `列出所有已注册的模拟器模块及其描述信息。

可以按类别过滤:
  chainspace-sim list --category consensus
  chainspace-sim list --category crypto
  chainspace-sim list --category attack`,
	RunE: runList,
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().StringP("category", "c", "", "按类别过滤 (blockchain, consensus, crypto, network, evm, attacks, defi, crosschain)")
	listCmd.Flags().StringP("type", "t", "", "按类型过滤 (tool, demo, process, attack, defi)")
	listCmd.Flags().Bool("json", false, "以JSON格式输出")
}

func runList(cmd *cobra.Command, args []string) error {
	category, _ := cmd.Flags().GetString("category")
	componentType, _ := cmd.Flags().GetString("type")
	jsonOutput, _ := cmd.Flags().GetBool("json")

	// 创建临时引擎获取模块列表
	eng := engine.NewEngine()
	registerAllModules(eng)

	modules := eng.ListSimulators()

	// 过滤
	if category != "" {
		var filtered []types.Description
		for _, m := range modules {
			if m.Category == category {
				filtered = append(filtered, m)
			}
		}
		modules = filtered
	}

	if componentType != "" {
		var filtered []types.Description
		for _, m := range modules {
			if string(m.Type) == componentType {
				filtered = append(filtered, m)
			}
		}
		modules = filtered
	}

	if jsonOutput {
		// JSON输出
		fmt.Println("[")
		for i, m := range modules {
			comma := ","
			if i == len(modules)-1 {
				comma = ""
			}
			fmt.Printf(`  {"id": "%s", "name": "%s", "category": "%s", "type": "%s"}%s`+"\n",
				m.ID, m.Name, m.Category, m.Type, comma)
		}
		fmt.Println("]")
	} else {
		// 表格输出
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tCATEGORY\tTYPE\tDESCRIPTION")
		fmt.Fprintln(w, "──\t────\t────────\t────\t───────────")
		for _, m := range modules {
			desc := m.Description
			if len(desc) > 40 {
				desc = desc[:37] + "..."
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", m.ID, m.Name, m.Category, m.Type, desc)
		}
		w.Flush()

		fmt.Printf("\n共 %d 个模块\n", len(modules))
	}

	return nil
}
