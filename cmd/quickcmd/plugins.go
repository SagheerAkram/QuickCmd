package main

import (
	"fmt"
	"os"
	"text/tabwriter"
	
	"github.com/spf13/cobra"
	"github.com/yourusername/quickcmd/core/plugins"
)

var pluginsCmd = &cobra.Command{
	Use:   "plugins",
	Short: "Manage QuickCMD plugins",
	Long:  `List, enable, disable, and get information about QuickCMD plugins.`,
}

var listPluginsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered plugins",
	RunE:  listPlugins,
}

var pluginInfoCmd = &cobra.Command{
	Use:   "info <plugin-name>",
	Short: "Show detailed information about a plugin",
	Args:  cobra.ExactArgs(1),
	RunE:  showPluginInfo,
}

func init() {
	rootCmd.AddCommand(pluginsCmd)
	pluginsCmd.AddCommand(listPluginsCmd)
	pluginsCmd.AddCommand(pluginInfoCmd)
	
	listPluginsCmd.Flags().Bool("enabled-only", false, "show only enabled plugins")
}

func listPlugins(cmd *cobra.Command, args []string) error {
	enabledOnly, _ := cmd.Flags().GetBool("enabled-only")
	
	var pluginList []plugins.Plugin
	if enabledOnly {
		pluginList = plugins.ListEnabled()
	} else {
		pluginList = plugins.List()
	}
	
	if len(pluginList) == 0 {
		fmt.Println("No plugins registered.")
		return nil
	}
	
	fmt.Printf("%sRegistered Plugins%s\n\n", colorBold, colorReset)
	
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "NAME\tVERSION\tSTATUS\tSCOPES")
	fmt.Fprintln(w, "----\t-------\t------\t------")
	
	for _, plugin := range pluginList {
		metadata, err := plugins.DefaultRegistry().GetMetadata(plugin.Name())
		if err != nil {
			continue
		}
		
		status := colorRed + "disabled" + colorReset
		if metadata.Enabled {
			status = colorGreen + "enabled" + colorReset
		}
		
		scopes := ""
		if len(plugin.Scopes()) > 0 {
			scopes = plugin.Scopes()[0]
			if len(plugin.Scopes()) > 1 {
				scopes += fmt.Sprintf(" (+%d)", len(plugin.Scopes())-1)
			}
		}
		
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", 
			plugin.Name(), 
			metadata.Version, 
			status, 
			scopes)
	}
	
	w.Flush()
	
	return nil
}

func showPluginInfo(cmd *cobra.Command, args []string) error {
	pluginName := args[0]
	
	plugin, err := plugins.Get(pluginName)
	if err != nil {
		return fmt.Errorf("plugin not found: %s", pluginName)
	}
	
	metadata, err := plugins.DefaultRegistry().GetMetadata(pluginName)
	if err != nil {
		return err
	}
	
	fmt.Printf("%sPlugin: %s%s\n\n", colorBold, pluginName, colorReset)
	
	fmt.Printf("%sVersion:%s %s\n", colorBold, colorReset, metadata.Version)
	fmt.Printf("%sDescription:%s %s\n", colorBold, colorReset, metadata.Description)
	fmt.Printf("%sAuthor:%s %s\n", colorBold, colorReset, metadata.Author)
	
	status := "disabled"
	statusColor := colorRed
	if metadata.Enabled {
		status = "enabled"
		statusColor = colorGreen
	}
	fmt.Printf("%sStatus:%s %s%s%s\n", colorBold, colorReset, statusColor, status, colorReset)
	
	fmt.Printf("\n%sRequired Scopes:%s\n", colorBold, colorReset)
	scopes := plugin.Scopes()
	if len(scopes) == 0 {
		fmt.Println("  (none)")
	} else {
		for _, scope := range scopes {
			fmt.Printf("  • %s\n", scope)
		}
	}
	
	return nil
}
