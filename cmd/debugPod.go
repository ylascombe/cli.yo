/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"context"
	"flag"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/util/homedir"
	//
	// Uncomment to load all auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth"
	//
	// Or uncomment to load specific auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

// debugPodCmd represents the debugPod command
var debugPodCmd = &cobra.Command{
	Use:   "debugPod",
	Short: "Run a debug-pod",
	Long:  `Use debug-pod pod configuration to launch a debug pod`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("debugPod called")
		// tmpListPod()
		createPod()

	},
}

func init() {
	kubeCmd.AddCommand(debugPodCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// debugPodCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// debugPodCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func tmpListPod() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	for {
		pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

		// Examples for error handling:
		// - Use helper functions like e.g. errors.IsNotFound()
		// - And/or cast to StatusError and use its properties like e.g. ErrStatus.Message
		namespace := "default"
		pod := "example-xxxxx"
		_, err = clientset.CoreV1().Pods(namespace).Get(context.TODO(), pod, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			fmt.Printf("Pod %s in namespace %s not found\n", pod, namespace)
		} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
			fmt.Printf("Error getting pod %s in namespace %s: %v\n",
				pod, namespace, statusError.ErrStatus.Message)
		} else if err != nil {
			panic(err.Error())
		} else {
			fmt.Printf("Found pod %s in namespace %s\n", pod, namespace)
		}

		time.Sleep(10 * time.Second)
	}
}

func createPod() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	namespace := "default"
	clientset, err := kubernetes.NewForConfig(config)
	podsClient := clientset.CoreV1().Pods(namespace)
	if err != nil {
		panic(err.Error())
	}

	podName := "debug-pod"
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: podName,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "main",
					Image:   "digitalocean/doks-debug:latest",
					Command: []string{"sleep", "infinity"},
					// Ports: []apiv1.ContainerPort{
					// 	{
					// 		Name:          "http",
					// 		Protocol:      apiv1.ProtocolTCP,
					// 		ContainerPort: 80,
					// 	},
					// },
				},
			},
		},
	}

	result, err := podsClient.Create(context.TODO(), pod, metav1.CreateOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Pod créé avec succès: %s\n", result.GetName())

	// Attendre que le pod soit prêt
	fmt.Println("Attente de l'état 'Running' pour le pod...")
	err = waitForPodRunning(clientset, podName, namespace)
	if err != nil {
		panic(err.Error())
	}

	execCommandInPod(config, clientset, podName, namespace, []string{"bash"})

}

func waitForPodRunning(clientset *kubernetes.Clientset, podName, namespace string) error {
	// Vérifier l'état du pod jusqu'à ce qu'il soit "Running"
	for {
		pod, err := clientset.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("échec lors de la récupération de l'état du pod : %v", err)
		}

		// Vérifier l'état du pod
		if pod.Status.Phase == corev1.PodRunning {
			fmt.Printf("Pod '%s' est maintenant dans l'état 'Running'\n", podName)
			return nil
		}

		fmt.Printf("État actuel du pod '%s': %s\n", podName, pod.Status.Phase)
		time.Sleep(2 * time.Second) // Attendre avant de vérifier à nouveau
	}
}

func execCommandInPod(config *rest.Config, clientset *kubernetes.Clientset, podName, namespace string, command []string) error {
	fmt.Printf("Prepare to execute command in pod..\n")
	req := clientset.CoreV1().RESTClient().
		Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		Param("container", "main").
		Param("stdin", "true").
		Param("stdout", "true").
		Param("stderr", "true").
		Param("tty", "true")

	for _, cmd := range command {
		req.Param("command", cmd)
	}

	// Préparer l'exécution du remote command
	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("error while configuring command execution: %v", err)
	}
	fmt.Printf("SPDY loaded\n")
	// Connecter les flux standards pour une interaction interactive
	// err = exec.Stream(remotecommand.StreamOptions{

	err = exec.StreamWithContext(context.Background(), remotecommand.StreamOptions{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Tty:    true,
	})
	fmt.Printf("Stream with context\n")
	if err != nil {
		return fmt.Errorf("échec lors de l'exécution de la commande dans le pod : %v", err)
	}

	fmt.Printf("Commande exécutée avec succès dans le pod '%s'\n", podName)
	return nil
}
