/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// kubeCmd represents the kube command
var kubeCmd = &cobra.Command{
	Use:   "kube",
	Short: "Command that contains some actions related to Kubernetes",
	Long:  `These commands will use current context or help to set it.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("kube called")
	},
}

func init() {
	rootCmd.AddCommand(kubeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// kubeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// kubeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
