package kube

import (
	"fmt"

	"context"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"

	"flag"
	"path/filepath"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type Kube struct {
	Clientset *kubernetes.Clientset `json:"-"`
	Config    *rest.Config          `json:"-"`
}

func NewKube() *Kube {
	config, clientset, err := setContext()
	if err != nil {
		panic(err.Error())
	}
	return &Kube{
		Clientset: clientset,
		Config:    config,
	}
}

func setContext() (*rest.Config, *kubernetes.Clientset, error) {
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

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return config, clientset, nil
}

func (k Kube) CreatePod(podName string, namespace string) {
	// create the clientset
	podsClient := k.Clientset.CoreV1().Pods(namespace)

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
	fmt.Printf("Pod successfully created: %s\n", result.GetName())

	fmt.Println("Waiting for pod become 'Running'...")
	err = k.waitForPodRunning(podName, namespace)
	if err != nil {
		panic(err.Error())
	}
}

func (k Kube) waitForPodRunning(podName, namespace string) error {
	// Loop while pod become running
	for {
		pod, err := k.Clientset.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("error while retrieving pod state: %v", err)
		}

		if pod.Status.Phase == corev1.PodRunning {
			fmt.Printf("Pod '%s' isq now in 'Running' state\n", podName)
			return nil
		}

		fmt.Printf("Pod is in state '%s': %s\n", podName, pod.Status.Phase)
		time.Sleep(2 * time.Second) // Wait 2s before new iteration
	}
}

func (k Kube) ExecCommandInPod(podName, namespace string, command []string) error {
	fmt.Printf("Prepare to execute command in pod..\n")
	req := k.Clientset.CoreV1().RESTClient().
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

	// Prepare remote command execution du remote command
	exec, err := remotecommand.NewSPDYExecutor(k.Config, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("error while configuring command execution: %v", err)
	}

	err = exec.StreamWithContext(context.Background(), remotecommand.StreamOptions{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Tty:    true,
	})

	if err != nil {
		return fmt.Errorf("error while running command into pod : %v", err)
	}

	fmt.Printf("Command successfully executed in pod '%s'\n", podName)
	return nil
}

func (k Kube) AlreadyExist(podName string, namespace string) bool {
	// Examples for error handling:
	// - Use helper functions like e.g. errors.IsNotFound()
	// - And/or cast to StatusError and use its properties like e.g. ErrStatus.Message
	_, err := k.Clientset.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		fmt.Printf("Pod %s in namespace %s not found\n", podName, namespace)
		return false
	} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
		fmt.Printf("Error getting pod %s in namespace %s: %v\n",
			podName, namespace, statusError.ErrStatus.Message)
		return false
	} else if err != nil {
		panic(err.Error())
	} else {
		fmt.Printf("Found pod %s in namespace %s\n", podName, namespace)
		return true
	}
}

func (k Kube) CreateDebugHostPod(nodeName string, podName string, namespace string) error {
	// create the clientset
	podsClient := k.Clientset.CoreV1().Pods(namespace)

	hostPathType := corev1.HostPathDirectory

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
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "host",
							MountPath: "/host",
						},
					},
					SecurityContext: &corev1.SecurityContext{
						Privileged: BoolAddr(true),
					},
				},
			},
			NodeSelector: map[string]string{
				"kubernetes.io/hostname": nodeName,
			},
			Volumes: []corev1.Volume{
				{
					Name: "host",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Type: &hostPathType,
							Path: "/",
						},
					},
				},
			},
		},
	}

	result, err := podsClient.Create(context.TODO(), pod, metav1.CreateOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Pod successfully created: %s\n", result.GetName())

	fmt.Println("Waiting for pod become 'Running'...")
	err = k.waitForPodRunning(podName, namespace)
	if err != nil {
		panic(err.Error())
	}
	return nil
}

func BoolAddr(b bool) *bool {
	boolVar := b
	return &boolVar
}
