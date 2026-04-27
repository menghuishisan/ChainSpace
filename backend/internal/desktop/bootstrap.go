package desktop

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

type Config struct {
	ProjectRoot string
	ServerURL   string
}

type Bootstrap struct {
	projectRoot string
	serverURL   string
}

func (b *Bootstrap) ProjectRoot() string {
	return b.projectRoot
}

func New(cfg Config) (*Bootstrap, error) {
	projectRoot := cfg.ProjectRoot
	if projectRoot == "" {
		resolved, err := ResolveProjectRoot()
		if err != nil {
			return nil, err
		}
		projectRoot = resolved
	}

	serverURL := cfg.ServerURL
	if serverURL == "" {
		serverURL = "http://127.0.0.1:3000"
	}

	return &Bootstrap{
		projectRoot: projectRoot,
		serverURL:   strings.TrimRight(serverURL, "/"),
	}, nil
}

func ResolveProjectRoot() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolve executable: %w", err)
	}

	dir := filepath.Dir(exePath)
	for {
		if fileExists(composeFilePath(dir)) && fileExists(configFilePath(dir)) {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	if cwd, err := os.Getwd(); err == nil {
		if fileExists(composeFilePath(cwd)) && fileExists(configFilePath(cwd)) {
			return cwd, nil
		}
	}

	return "", errors.New("project root not found near executable")
}

func (b *Bootstrap) Prepare(ctx context.Context) error {
	if err := b.ensureDockerCLI(ctx); err != nil {
		return err
	}

	if err := b.startInfrastructure(ctx); err != nil {
		return err
	}

	return b.waitForInfrastructure(ctx)
}

func (b *Bootstrap) OpenUI() error {
	return openBrowser(b.serverURL)
}

func (b *Bootstrap) WaitForServer(ctx context.Context) error {
	client := &http.Client{Timeout: 5 * time.Second}
	healthURL := b.serverURL + "/health"
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
		if err == nil {
			resp, err := client.Do(req)
			if err == nil {
				_ = resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					return nil
				}
			}
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("wait for server: %w", ctx.Err())
		case <-ticker.C:
		}
	}
}

func (b *Bootstrap) startInfrastructure(ctx context.Context) error {
	composeFile := composeFilePath(b.projectRoot)
	cmd := exec.CommandContext(
		ctx,
		"docker",
		"compose",
		"-f",
		composeFile,
		"up",
		"-d",
		"postgres",
		"redis",
		"minio",
		"minio-init",
		"rabbitmq",
	)
	cmd.Dir = b.projectRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("start infrastructure: %w", err)
	}
	return nil
}

func (b *Bootstrap) ensureDockerCLI(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "docker", "info")
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker is not available, please start Docker Desktop first: %w", err)
	}
	return nil
}

func (b *Bootstrap) waitForInfrastructure(ctx context.Context) error {
	checks := []func(context.Context) error{
		waitForTCP("127.0.0.1:5433"),
		waitForTCP("127.0.0.1:6381"),
		waitForHTTP("http://127.0.0.1:9000/minio/health/live"),
		waitForTCP("127.0.0.1:5672"),
	}

	for _, check := range checks {
		if err := check(ctx); err != nil {
			return err
		}
	}

	return nil
}

func waitForTCP(addr string) func(context.Context) error {
	return func(ctx context.Context) error {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
			if err == nil {
				_ = conn.Close()
				return nil
			}

			select {
			case <-ctx.Done():
				return fmt.Errorf("wait for %s: %w", addr, ctx.Err())
			case <-ticker.C:
			}
		}
	}
}

func waitForHTTP(url string) func(context.Context) error {
	return func(ctx context.Context) error {
		client := &http.Client{Timeout: 5 * time.Second}
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err == nil {
				resp, err := client.Do(req)
				if err == nil {
					_ = resp.Body.Close()
					if resp.StatusCode == http.StatusOK {
						return nil
					}
				}
			}

			select {
			case <-ctx.Done():
				return fmt.Errorf("wait for %s: %w", url, ctx.Err())
			case <-ticker.C:
			}
		}
	}
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func composeFilePath(projectRoot string) string {
	return filepath.Join(projectRoot, "deploy", "docker-compose.desktop.yml")
}

func configFilePath(projectRoot string) string {
	return filepath.Join(projectRoot, "backend", "configs", "config.desktop.yaml")
}

func ShowErrorDialog(title string, message string) {
	if runtime.GOOS != "windows" {
		fmt.Fprintf(os.Stderr, "%s: %s\n", title, message)
		return
	}

	user32 := syscall.NewLazyDLL("user32.dll")
	messageBoxW := user32.NewProc("MessageBoxW")
	_, _, _ = messageBoxW.Call(
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(message))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))),
		0x00000010,
	)
}
