/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/ylascombe/cli.yo/pkg/kube"
)

// debugHostCmd represents the debugHost command
var debugHostCmd = &cobra.Command{
	Use:   "debugHost",
	Short: "Run a debug pod on specific host in host mode",
	Long: `Launch a debug pod on a specific nodes with advanced permissions
and host volumes mount in order to debug node`,
	Run: func(cmd *cobra.Command, args []string) {
		k := kube.NewKube()
		exists := k.AlreadyExist(PodName, Namespace)
		if !exists {
			k.CreateDebugHostPod(NodeName, PodName, Namespace)
		}
		k.ExecCommandInPod(PodName, Namespace, []string{"bash"})
	},
}

var NodeName string

func init() {
	kubeCmd.AddCommand(debugHostCmd)
	debugHostCmd.Flags().StringVarP(&NodeName, "hostname", "j", "", "Kubernetes node name to debug")
	debugHostCmd.MarkFlagRequired("hostname")

	debugHostCmd.Flags().StringVarP(&PodName, "name", "p", "debug-pod", "Debug pod name")
	debugHostCmd.Flags().StringVarP(&Namespace, "namespace", "n", "default", "Namespace on which debug pod will be created")

}
