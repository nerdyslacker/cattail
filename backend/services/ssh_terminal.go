package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/creack/pty"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/crypto/ssh"
)

type sshSessionType int

const (
	sshTypeTailscale sshSessionType = iota
	sshTypeDirect
)

type sshSession struct {
	ID          string
	PeerName    string
	SessionType sshSessionType
	Cancel      context.CancelFunc

	Cmd         *exec.Cmd
	PtyFile     *os.File
	pipeStdout  io.Reader

	SSHClient  *ssh.Client
	SSHSession *ssh.Session

	StdinPipe io.WriteCloser
}

func generateSessionID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("ssh_%x", b)
}

func (tailSvc *tailScaleService) ConnectTailscaleSSH(peerName string) (string, error) {
	sessionID := generateSessionID()
	ctx, cancel := context.WithCancel(tailSvc.ctx)

	cmd := exec.CommandContext(ctx, "tailscale", "ssh", peerName)

	var stdin io.WriteCloser
	var stdout io.Reader

	ptyFile, err := pty.StartWithSize(cmd, &pty.Winsize{Cols: 80, Rows: 24})
	if err != nil {
		log.Printf("PTY not available for tailscale ssh, using pipes: %v", err)
		stdin, _ = cmd.StdinPipe()
		stdout, _ = cmd.StdoutPipe()
		cmd.Stderr = cmd.Stdout
		if startErr := cmd.Start(); startErr != nil {
			cancel()
			return "", fmt.Errorf("starting tailscale ssh: %w", startErr)
		}
	} else {
		stdin = ptyFile
		stdout = ptyFile
	}

	session := &sshSession{
		ID:          sessionID,
		PeerName:    peerName,
		SessionType: sshTypeTailscale,
		Cancel:      cancel,
		Cmd:         cmd,
		PtyFile:     ptyFile,
		pipeStdout:  stdout,
		StdinPipe:   stdin,
	}

	tailSvc.sshMu.Lock()
	tailSvc.sshSessions[sessionID] = session
	tailSvc.sshMu.Unlock()

	go tailSvc.sshOutputReader(session)
	go tailSvc.sshWaiter(session)

	return sessionID, nil
}

func (tailSvc *tailScaleService) ConnectDirectSSH(host string, port int,
	username, password, privateKeyPath string) (string, error) {

	sessionID := generateSessionID()
	_, cancel := context.WithCancel(tailSvc.ctx)

	addr := fmt.Sprintf("%s:%d", host, port)
	config := &ssh.ClientConfig{
		User:            username,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	if password != "" {
		config.Auth = append(config.Auth, ssh.Password(password))
	}
	if privateKeyPath != "" {
		key, err := os.ReadFile(privateKeyPath)
		if err != nil {
			cancel()
			return "", fmt.Errorf("reading private key: %w", err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			cancel()
			return "", fmt.Errorf("parsing private key: %w", err)
		}
		config.Auth = append(config.Auth, ssh.PublicKeys(signer))
	}

	if len(config.Auth) == 0 {
		cancel()
		return "", fmt.Errorf("no authentication method provided")
	}

	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		cancel()
		return "", fmt.Errorf("dialing ssh: %w", err)
	}

	sshSess, err := client.NewSession()
	if err != nil {
		client.Close()
		cancel()
		return "", fmt.Errorf("creating ssh session: %w", err)
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if err := sshSess.RequestPty("xterm-256color", 24, 80, modes); err != nil {
		sshSess.Close()
		client.Close()
		cancel()
		return "", fmt.Errorf("requesting pty: %w", err)
	}

	stdin, _ := sshSess.StdinPipe()
	stdout, _ := sshSess.StdoutPipe()
	sshSess.Stderr = sshSess.Stdout

	if err := sshSess.Shell(); err != nil {
		sshSess.Close()
		client.Close()
		cancel()
		return "", fmt.Errorf("starting shell: %w", err)
	}

	session := &sshSession{
		ID:          sessionID,
		PeerName:    host,
		SessionType: sshTypeDirect,
		Cancel:      cancel,
		SSHClient:   client,
		SSHSession:  sshSess,
		pipeStdout:  stdout,
		StdinPipe:   stdin,
	}

	tailSvc.sshMu.Lock()
	tailSvc.sshSessions[sessionID] = session
	tailSvc.sshMu.Unlock()

	go tailSvc.sshOutputReader(session)
	go tailSvc.sshWaiter(session)

	return sessionID, nil
}

func (tailSvc *tailScaleService) WriteSSHInput(sessionID string, data string) error {
	tailSvc.sshMu.Lock()
	session, ok := tailSvc.sshSessions[sessionID]
	tailSvc.sshMu.Unlock()
	if !ok {
		return nil
	}
	_, err := session.StdinPipe.Write([]byte(data))
	return err
}

func (tailSvc *tailScaleService) ResizeSSHTerminal(sessionID string, cols, rows int) error {
	tailSvc.sshMu.Lock()
	session, ok := tailSvc.sshSessions[sessionID]
	tailSvc.sshMu.Unlock()
	if !ok {
		return nil
	}

	switch session.SessionType {
	case sshTypeTailscale:
		if session.PtyFile != nil {
			return pty.Setsize(session.PtyFile, &pty.Winsize{Cols: uint16(cols), Rows: uint16(rows)})
		}
	case sshTypeDirect:
		if session.SSHSession != nil {
			return session.SSHSession.WindowChange(rows, cols)
		}
	}
	return nil
}

func (tailSvc *tailScaleService) CloseSSHSession(sessionID string) error {
	tailSvc.sshMu.Lock()
	session, ok := tailSvc.sshSessions[sessionID]
	if !ok {
		tailSvc.sshMu.Unlock()
		return nil
	}
	delete(tailSvc.sshSessions, sessionID)
	tailSvc.sshMu.Unlock()

	session.Cancel()

	switch session.SessionType {
	case sshTypeTailscale:
		if session.Cmd != nil && session.Cmd.Process != nil {
			session.Cmd.Process.Kill()
		}
		if session.PtyFile != nil {
			session.PtyFile.Close()
		}
	case sshTypeDirect:
		if session.SSHSession != nil {
			session.SSHSession.Close()
		}
		if session.SSHClient != nil {
			session.SSHClient.Close()
		}
	}

	runtime.EventsEmit(tailSvc.ctx, "ssh:closed:"+sessionID)
	return nil
}

func (tailSvc *tailScaleService) BrowsePrivateKey() (string, error) {
	path, err := runtime.OpenFileDialog(tailSvc.ctx, runtime.OpenDialogOptions{
		Title: "Select SSH Private Key",
		Filters: []runtime.FileFilter{
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
			{DisplayName: "PEM Files (*.pem)", Pattern: "*.pem"},
		},
	})
	if err != nil {
		return "", err
	}
	return path, nil
}

func (tailSvc *tailScaleService) sshOutputReader(session *sshSession) {
	var source io.Reader
	if session.PtyFile != nil {
		source = session.PtyFile
	} else {
		source = session.pipeStdout
	}

	if source == nil {
		return
	}

	buf := make([]byte, 4096)
	for {
		n, err := source.Read(buf)
		if n > 0 {
			encoded := base64.StdEncoding.EncodeToString(buf[:n])
			runtime.EventsEmit(tailSvc.ctx, "ssh:output:"+session.ID, encoded)
		}
		if err != nil {
			if err != io.EOF {
				runtime.EventsEmit(tailSvc.ctx, "ssh:error:"+session.ID, err.Error())
			}
			return
		}
	}
}

func (tailSvc *tailScaleService) sshWaiter(session *sshSession) {
	var exitErr error

	switch session.SessionType {
	case sshTypeTailscale:
		exitErr = session.Cmd.Wait()
	case sshTypeDirect:
		exitErr = session.SSHSession.Wait()
	}

	if exitErr != nil {
		runtime.EventsEmit(tailSvc.ctx, "ssh:error:"+session.ID, exitErr.Error())
	}
	runtime.EventsEmit(tailSvc.ctx, "ssh:closed:"+session.ID)
}
