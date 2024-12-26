/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/ylascombe/cli.yo/pkg/kube"
)

// debugPodCmd represents the debugPod command
var debugPodCmd = &cobra.Command{
	Use:   "debugPod",
	Short: "Run a debug-pod",
	Long:  `Use debug-pod pod configuration to launch a debug pod`,
	Run: func(cmd *cobra.Command, args []string) {
		k := kube.NewKube()
		exists := k.AlreadyExist(PodName, Namespace)
		if !exists {
			k.CreatePod(PodName, Namespace)
		}
		err := k.ExecCommandInPod(PodName, Namespace, []string{"bash"})
		if err != nil {
			panic(err.Error())
		}
	},
}

var Namespace string
var PodName string

func init() {
	kubeCmd.AddCommand(debugPodCmd)
	debugPodCmd.Flags().StringVarP(&PodName, "name", "p", "debug-pod", "Debug pod name")
	debugPodCmd.Flags().StringVarP(&Namespace, "namespace", "n", "default", "Namespace on which debug pod will be created")
}
