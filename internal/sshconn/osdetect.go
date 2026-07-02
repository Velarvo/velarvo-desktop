package sshconn

import (
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

const osProbeTimeout = 5 * time.Second

const (
	OSLinux   = "linux"
	OSUbuntu  = "ubuntu"
	OSDebian  = "debian"
	OSAlpine  = "alpine"
	OSFedora  = "fedora"
	OSCentOS  = "centos"
	OSRHEL    = "rhel"
	OSArch    = "arch"
	OSMacOS   = "macos"
	OSWindows = "windows"
	OSFreeBSD = "freebsd"
)

func DetectOS(client *ssh.Client) string {
	if out, err := runProbe(client, "uname -s"); err == nil {
		switch kernel := strings.ToLower(strings.TrimSpace(out)); {
		case strings.Contains(kernel, "darwin"):
			return OSMacOS
		case strings.Contains(kernel, "freebsd"):
			return OSFreeBSD
		case strings.Contains(kernel, "linux"):
			return detectLinuxDistro(client)
		}
	}

	if out, err := runProbe(client, "cmd /c ver"); err == nil {
		if strings.Contains(strings.ToLower(out), "windows") {
			return OSWindows
		}
	}

	return ""
}

func detectLinuxDistro(client *ssh.Client) string {
	out, err := runProbe(client, "cat /etc/os-release")
	if err != nil {
		return OSLinux
	}

	switch osReleaseID(out) {
	case "ubuntu":
		return OSUbuntu
	case "debian", "raspbian":
		return OSDebian
	case "alpine":
		return OSAlpine
	case "fedora":
		return OSFedora
	case "centos":
		return OSCentOS
	case "rhel", "redhat":
		return OSRHEL
	case "arch":
		return OSArch
	default:
		return OSLinux
	}
}

func osReleaseID(content string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "ID=") {
			continue
		}
		value := strings.TrimPrefix(line, "ID=")
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		return strings.ToLower(value)
	}
	return ""
}

func runProbe(client *ssh.Client, cmd string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer func() { _ = session.Close() }()

	type result struct {
		out []byte
		err error
	}
	done := make(chan result, 1)
	go func() {
		out, err := session.Output(cmd)
		done <- result{out: out, err: err}
	}()

	select {
	case r := <-done:
		if r.err != nil {
			return "", r.err
		}
		return string(r.out), nil
	case <-time.After(osProbeTimeout):
		_ = session.Close()
		return "", errProbeTimeout
	}
}
