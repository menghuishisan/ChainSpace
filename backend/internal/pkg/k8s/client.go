package k8s

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/chainspace/backend/internal/config"
	"github.com/chainspace/backend/internal/pkg/logger"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

type Client struct {
	mu         sync.RWMutex
	clientset  *kubernetes.Clientset
	restConfig *rest.Config
	namespace  string
	cfg        *config.KubernetesConfig
}

func resolveKubeconfig(path string) string {
	if path == "" {
		home := os.Getenv("HOME")
		if home == "" && runtime.GOOS == "windows" {
			home = os.Getenv("USERPROFILE")
		}
		if home != "" {
			return filepath.Join(home, ".kube", "config")
		}
		return ""
	}

	if strings.HasPrefix(path, "~/") || strings.HasPrefix(path, "~\\") {
		home := os.Getenv("HOME")
		if home == "" && runtime.GOOS == "windows" {
			home = os.Getenv("USERPROFILE")
		}
		if home != "" {
			return filepath.Join(home, path[2:])
		}
	}

	return path
}

func buildRESTConfig(cfg *config.KubernetesConfig) (*rest.Config, error) {
	if cfg == nil {
		return nil, fmt.Errorf("k8s config is required")
	}

	var (
		restConfig *rest.Config
		err        error
	)
	if cfg.InCluster {
		restConfig, err = rest.InClusterConfig()
	} else {
		restConfig, err = clientcmd.BuildConfigFromFlags("", resolveKubeconfig(cfg.Kubeconfig))
	}
	if err != nil {
		return nil, fmt.Errorf("failed to build k8s config: %w", err)
	}
	return restConfig, nil
}

func NewClient(cfg *config.KubernetesConfig) (*Client, error) {
	restConfig, err := buildRESTConfig(cfg)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s clientset: %w", err)
	}

	return &Client{
		clientset:  clientset,
		restConfig: restConfig,
		namespace:  cfg.Namespace,
		cfg:        cfg,
	}, nil
}

func (c *Client) ensureFreshClientset() error {
	if c == nil {
		return fmt.Errorf("k8s client is nil")
	}
	if c.cfg == nil {
		return fmt.Errorf("k8s config is nil")
	}
	if c.cfg.InCluster {
		return nil
	}

	restConfig, err := buildRESTConfig(c.cfg)
	if err != nil {
		return err
	}
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("failed to create k8s clientset: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.restConfig != nil && c.restConfig.Host == restConfig.Host && c.clientset != nil {
		return nil
	}
	c.restConfig = restConfig
	c.clientset = clientset
	return nil
}

func (c *Client) currentClientset() (*kubernetes.Clientset, error) {
	if err := c.ensureFreshClientset(); err != nil {
		return nil, err
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.clientset == nil {
		return nil, fmt.Errorf("k8s clientset is not initialized")
	}
	return c.clientset, nil
}

func (c *Client) currentRESTConfig() (*rest.Config, error) {
	if err := c.ensureFreshClientset(); err != nil {
		return nil, err
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.restConfig == nil {
		return nil, fmt.Errorf("k8s rest config is not initialized")
	}
	return c.restConfig, nil
}

type PodConfig struct {
	EnvID        string
	UserID       uint
	SchoolID     uint
	ExperimentID uint
	Image        string
	Command      []string
	Args         []string
	CPU          string
	Memory       string
	Storage      string
	Timeout      time.Duration
	Ports        []int32
	ProbePort    int32
	EnvVars      map[string]string
	InitFiles    string
}

func requiresImagePullSecret(image string) bool {
	repository := strings.TrimSpace(image)
	if repository == "" {
		return false
	}

	firstSegment := repository
	if slash := strings.Index(repository, "/"); slash >= 0 {
		firstSegment = repository[:slash]
	}

	return strings.Contains(firstSegment, ".") || strings.Contains(firstSegment, ":") || firstSegment == "localhost"
}

func (c *Client) CreatePod(ctx context.Context, cfg *PodConfig) (*corev1.Pod, error) {
	clientset, err := c.currentClientset()
	if err != nil {
		return nil, err
	}

	labels := map[string]string{
		"app":           "chainspace-env",
		"env-id":        cfg.EnvID,
		"user-id":       fmt.Sprintf("%d", cfg.UserID),
		"school-id":     fmt.Sprintf("%d", cfg.SchoolID),
		"experiment-id": fmt.Sprintf("%d", cfg.ExperimentID),
	}

	resources := corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(cfg.CPU),
			corev1.ResourceMemory: resource.MustParse(cfg.Memory),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("128Mi"),
		},
	}

	envVars := []corev1.EnvVar{
		{Name: "ENV_ID", Value: cfg.EnvID},
		{Name: "USER_ID", Value: fmt.Sprintf("%d", cfg.UserID)},
		{Name: "EXPERIMENT_ID", Value: fmt.Sprintf("%d", cfg.ExperimentID)},
	}
	for key, value := range cfg.EnvVars {
		envVars = append(envVars, corev1.EnvVar{Name: key, Value: value})
	}

	makeContainerPorts := func(ports []int32) []corev1.ContainerPort {
		containerPorts := make([]corev1.ContainerPort, len(ports))
		for index, port := range ports {
			containerPorts[index] = corev1.ContainerPort{
				ContainerPort: port,
				Protocol:      corev1.ProtocolTCP,
			}
		}
		return containerPorts
	}

	probePort := cfg.ProbePort
	if probePort <= 0 && len(cfg.Ports) > 0 && cfg.Ports[0] > 0 {
		probePort = cfg.Ports[0]
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cfg.EnvID,
			Namespace: c.namespace,
			Labels:    labels,
			Annotations: map[string]string{
				"chainspace.io/timeout": cfg.Timeout.String(),
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:            "experiment",
					Image:           cfg.Image,
					ImagePullPolicy: corev1.PullIfNotPresent,
					Command:         cfg.Command,
					Args:            cfg.Args,
					Resources:       resources,
					Env:             envVars,
					Ports:           makeContainerPorts(cfg.Ports),
					ReadinessProbe: func() *corev1.Probe {
						if probePort <= 0 {
							return nil
						}
						return &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								TCPSocket: &corev1.TCPSocketAction{
									Port: intstrFromInt32(probePort),
								},
							},
							InitialDelaySeconds: 3,
							PeriodSeconds:       5,
							TimeoutSeconds:      2,
							FailureThreshold:    12,
						}
					}(),
					LivenessProbe: func() *corev1.Probe {
						if probePort <= 0 {
							return nil
						}
						return &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								TCPSocket: &corev1.TCPSocketAction{
									Port: intstrFromInt32(probePort),
								},
							},
							InitialDelaySeconds: 15,
							PeriodSeconds:       10,
							TimeoutSeconds:      2,
							FailureThreshold:    6,
						}
					}(),
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "workspace",
							MountPath: "/workspace",
						},
					},
				},
			},
			ImagePullSecrets: func() []corev1.LocalObjectReference {
				if c.cfg == nil || strings.TrimSpace(c.cfg.ImagePullSecret) == "" || !requiresImagePullSecret(cfg.Image) {
					return nil
				}
				return []corev1.LocalObjectReference{{Name: strings.TrimSpace(c.cfg.ImagePullSecret)}}
			}(),
			Volumes: []corev1.Volume{
				{
					Name: "workspace",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{
							SizeLimit: func() *resource.Quantity {
								quantity := resource.MustParse(cfg.Storage)
								return &quantity
							}(),
						},
					},
				},
			},
			RestartPolicy:                 corev1.RestartPolicyNever,
			AutomountServiceAccountToken:  func() *bool { value := false; return &value }(),
			TerminationGracePeriodSeconds: func() *int64 { value := int64(30); return &value }(),
		},
	}

	created, err := clientset.CoreV1().Pods(c.namespace).Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create pod: %w", err)
	}

	logger.Info("Pod created", zap.String("name", created.Name), zap.String("namespace", c.namespace))
	return created, nil
}

func (c *Client) Namespace() string {
	if c == nil {
		return ""
	}
	return c.namespace
}

func (c *Client) DeletePod(ctx context.Context, name string) error {
	clientset, err := c.currentClientset()
	if err != nil {
		return err
	}
	if err := clientset.CoreV1().Pods(c.namespace).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		return fmt.Errorf("failed to delete pod: %w", err)
	}
	logger.Info("Pod deleted", zap.String("name", name))
	return nil
}

func (c *Client) GetPod(ctx context.Context, name string) (*corev1.Pod, error) {
	clientset, err := c.currentClientset()
	if err != nil {
		return nil, err
	}
	return clientset.CoreV1().Pods(c.namespace).Get(ctx, name, metav1.GetOptions{})
}

func (c *Client) GetPodStatus(ctx context.Context, name string) (string, error) {
	pod, err := c.GetPod(ctx, name)
	if err != nil {
		return "", err
	}
	return string(pod.Status.Phase), nil
}

func (c *Client) GetPodIP(ctx context.Context, name string) (string, error) {
	pod, err := c.GetPod(ctx, name)
	if err != nil {
		return "", err
	}
	return pod.Status.PodIP, nil
}

func (c *Client) WaitForPodReady(ctx context.Context, name string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		pod, err := c.GetPod(ctx, name)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}

		if pod.Status.Phase == corev1.PodRunning {
			for _, cond := range pod.Status.Conditions {
				if cond.Type == corev1.PodReady && cond.Status == corev1.ConditionTrue {
					return nil
				}
			}
		}

		if pod.Status.Phase == corev1.PodFailed || pod.Status.Phase == corev1.PodSucceeded {
			return fmt.Errorf("pod terminated with phase: %s", pod.Status.Phase)
		}

		time.Sleep(time.Second)
	}

	return fmt.Errorf("timeout waiting for pod ready")
}

func (c *Client) ListPodsByUser(ctx context.Context, userID uint) (*corev1.PodList, error) {
	clientset, err := c.currentClientset()
	if err != nil {
		return nil, err
	}
	return clientset.CoreV1().Pods(c.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=chainspace-env,user-id=%d", userID),
	})
}

func (c *Client) ListPodsBySchool(ctx context.Context, schoolID uint) (*corev1.PodList, error) {
	clientset, err := c.currentClientset()
	if err != nil {
		return nil, err
	}
	return clientset.CoreV1().Pods(c.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=chainspace-env,school-id=%d", schoolID),
	})
}

func (c *Client) CreateService(ctx context.Context, envID string, ports []int32) (*corev1.Service, error) {
	clientset, err := c.currentClientset()
	if err != nil {
		return nil, err
	}

	servicePorts := make([]corev1.ServicePort, len(ports))
	for index, port := range ports {
		servicePorts[index] = corev1.ServicePort{
			Name:     fmt.Sprintf("port-%d", port),
			Port:     port,
			Protocol: corev1.ProtocolTCP,
		}
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-svc", envID),
			Namespace: c.namespace,
			Labels: map[string]string{
				"app":    "chainspace-env",
				"env-id": envID,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"env-id": envID,
			},
			Ports: servicePorts,
			Type:  corev1.ServiceTypeClusterIP,
		},
	}

	return clientset.CoreV1().Services(c.namespace).Create(ctx, service, metav1.CreateOptions{})
}

func (c *Client) DeleteService(ctx context.Context, name string) error {
	clientset, err := c.currentClientset()
	if err != nil {
		return err
	}
	return clientset.CoreV1().Services(c.namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (c *Client) ExecCommand(ctx context.Context, podName string, command []string) (string, error) {
	return c.ExecCommandInContainer(ctx, podName, "experiment", command)
}

func (c *Client) ExecCommandInContainer(ctx context.Context, podName, containerName string, command []string) (string, error) {
	clientset, err := c.currentClientset()
	if err != nil {
		return "", err
	}
	restConfig, err := c.currentRESTConfig()
	if err != nil {
		return "", err
	}

	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(c.namespace).
		SubResource("exec").
		Param("container", containerName).
		Param("stdout", "true").
		Param("stderr", "true")

	for _, cmd := range command {
		req.Param("command", cmd)
	}

	executor, err := remotecommand.NewSPDYExecutor(restConfig, "POST", req.URL())
	if err != nil {
		return "", fmt.Errorf("create executor: %w", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := executor.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	}); err != nil {
		return "", fmt.Errorf("exec stream: %w, stderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

func (c *Client) GetContainerLogs(ctx context.Context, podName, containerName string, tailLines int64) (string, error) {
	clientset, err := c.currentClientset()
	if err != nil {
		return "", err
	}
	req := clientset.CoreV1().Pods(c.namespace).GetLogs(podName, &corev1.PodLogOptions{
		Container:  containerName,
		Timestamps: true,
		TailLines:  &tailLines,
	})

	stream, err := req.Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("stream logs: %w", err)
	}
	defer stream.Close()

	data, err := io.ReadAll(stream)
	if err != nil {
		return "", fmt.Errorf("read logs: %w", err)
	}
	return string(data), nil
}

func (c *Client) getRestConfig() *rest.Config {
	if restConfig, err := c.currentRESTConfig(); err == nil {
		return restConfig
	}
	restConfig, _ := buildRESTConfig(c.cfg)
	return restConfig
}

func (c *Client) GetProxyURL(podName string, port int, path string) string {
	restConfig := c.getRestConfig()
	host := restConfig.Host
	if !strings.HasSuffix(host, "/") {
		host += "/"
	}
	return fmt.Sprintf("%sapi/v1/namespaces/%s/pods/%s:%d/proxy/%s", host, c.namespace, podName, port, strings.TrimPrefix(path, "/"))
}

func (c *Client) GetRESTConfig() *rest.Config {
	return c.getRestConfig()
}

func (c *Client) GetClientset() *kubernetes.Clientset {
	clientset, err := c.currentClientset()
	if err != nil {
		return nil
	}
	return clientset
}

func intstrFromInt32(value int32) intstr.IntOrString {
	return intstr.FromInt32(value)
}
