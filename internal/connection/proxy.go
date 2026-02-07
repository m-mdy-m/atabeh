package connection

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/m-mdy-m/atabeh/internal/logger"
)

const proxyTag = "proxy"

func NewSystemProxy() *SystemProxy {
	return &SystemProxy{}
}

func (s *SystemProxy) Enable(socksAddr, httpAddr string) error {
	logger.Infof(proxyTag, "Enabling system-wide proxy...")

	if runtime.GOOS == "linux" {
		return s.enableLinux(socksAddr, httpAddr)
	} else if runtime.GOOS == "windows" {
		return s.enableWindows(socksAddr, httpAddr)
	}

	return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
}

func (s *SystemProxy) Disable() error {
	if !s.enabled {
		return nil
	}

	logger.Infof(proxyTag, "Disabling system-wide proxy...")

	if runtime.GOOS == "linux" {
		return s.disableLinux()
	} else if runtime.GOOS == "windows" {
		return s.disableWindows()
	}

	return nil
}

func (s *SystemProxy) enableLinux(socksAddr, httpAddr string) error {
	host, port := splitAddr(httpAddr)

	cmds := [][]string{
		{"gsettings", "set", "org.gnome.system.proxy", "mode", "manual"},
		{"gsettings", "set", "org.gnome.system.proxy.http", "host", host},
		{"gsettings", "set", "org.gnome.system.proxy.http", "port", port},
		{"gsettings", "set", "org.gnome.system.proxy.https", "host", host},
		{"gsettings", "set", "org.gnome.system.proxy.https", "port", port},
		{"gsettings", "set", "org.gnome.system.proxy.socks", "host", strings.Split(socksAddr, ":")[0]},
		{"gsettings", "set", "org.gnome.system.proxy.socks", "port", strings.Split(socksAddr, ":")[1]},
	}

	for _, cmd := range cmds {
		if err := exec.Command(cmd[0], cmd[1:]...).Run(); err != nil {
			logger.Debugf(proxyTag, "gsettings command failed (may not be GNOME): %v", err)
		}
	}

	os.Setenv("http_proxy", "http://"+httpAddr)
	os.Setenv("https_proxy", "http://"+httpAddr)
	os.Setenv("HTTP_PROXY", "http://"+httpAddr)
	os.Setenv("HTTPS_PROXY", "http://"+httpAddr)
	os.Setenv("all_proxy", "socks5://"+socksAddr)
	os.Setenv("ALL_PROXY", "socks5://"+socksAddr)

	s.enabled = true
	logger.Infof(proxyTag, "Enabled (GNOME + environment)")
	return nil
}

func (s *SystemProxy) disableLinux() error {

	exec.Command("gsettings", "set", "org.gnome.system.proxy", "mode", "none").Run()

	os.Unsetenv("http_proxy")
	os.Unsetenv("https_proxy")
	os.Unsetenv("HTTP_PROXY")
	os.Unsetenv("HTTPS_PROXY")
	os.Unsetenv("all_proxy")
	os.Unsetenv("ALL_PROXY")

	s.enabled = false
	logger.Infof(proxyTag, "Disabled")
	return nil
}

func (s *SystemProxy) enableWindows(socksAddr, httpAddr string) error {

	cmd1 := exec.Command("reg", "add",
		`HKCU\Software\Microsoft\Windows\CurrentVersion\Internet Settings`,
		"/v", "ProxyEnable",
		"/t", "REG_DWORD",
		"/d", "1",
		"/f")

	if err := cmd1.Run(); err != nil {
		return fmt.Errorf("enable registry: %w", err)
	}

	proxyValue := fmt.Sprintf("http=%s;https=%s;socks=%s", httpAddr, httpAddr, socksAddr)
	cmd2 := exec.Command("reg", "add",
		`HKCU\Software\Microsoft\Windows\CurrentVersion\Internet Settings`,
		"/v", "ProxyServer",
		"/t", "REG_SZ",
		"/d", proxyValue,
		"/f")

	if err := cmd2.Run(); err != nil {
		return fmt.Errorf("set proxy: %w", err)
	}

	exec.Command("netsh", "winhttp", "import", "proxy", "source=ie").Run()

	s.enabled = true
	logger.Infof(proxyTag, "Enabled (registry)")
	return nil
}

func (s *SystemProxy) disableWindows() error {
	cmd := exec.Command("reg", "add",
		`HKCU\Software\Microsoft\Windows\CurrentVersion\Internet Settings`,
		"/v", "ProxyEnable",
		"/t", "REG_DWORD",
		"/d", "0",
		"/f")

	if err := cmd.Run(); err != nil {
		return err
	}

	exec.Command("netsh", "winhttp", "reset", "proxy").Run()

	s.enabled = false
	logger.Infof(proxyTag, "Disabled")
	return nil
}

func splitAddr(addr string) (host, port string) {
	parts := strings.Split(addr, ":")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return addr, "0"
}

func IsRoot() bool {
	if runtime.GOOS == "windows" {
		_, err := exec.Command("net", "session").Output()
		return err == nil
	}
	return os.Geteuid() == 0
}
