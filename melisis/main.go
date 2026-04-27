package main

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/fatih/color"
	"golang.org/x/crypto/ssh"
	"gopkg.in/ini.v1"
)

type Configuration struct {
	Path        string
	Filename    string
	BindIP      string
	Port        int
	RunningMode ServerMode
}

type ServerApp struct {
	Config   Configuration
	Logger   BanLogger
	Messages SystemMessages
	Signer   ssh.Signer
}

func main() {
	app, err := initializeApp()
	if err != nil {
		app.printFatalError(err)
		return
	}

	app.Start()
}

func initializeApp() (ServerApp, error) {
	app := ServerApp{
		Config: Configuration{
			Path:     "/etc/Melisis",
			Filename: "melisis.conf",
		},
		Logger: BanLogger{
			FilePath: "/var/log/Melisis.txt",
		},
		Messages: NewSystemMessages(),
	}

	signer, err := app.generateKeys()
	if err != nil {
		return app, err
	}
	app.Signer = signer

	manager := DependencyManager{
		Fail2BanBinary: "/usr/bin/fail2ban-client",
		LogFilePath:    app.Logger.FilePath,
	}

	return app, manager.CheckAndInstall()
}

func (app *ServerApp) generateKeys() (ssh.Signer, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	return ssh.NewSignerFromKey(key)
}

func (app *ServerApp) Start() {
	for {
		err := app.loadConfig()
		if err != nil {
			app.handleConfigError(err)
			return
		}

		app.checkPrivileges()
		app.startListening()
	}
}

func (app *ServerApp) loadConfig() error {
	fullPath := fmt.Sprintf("%s/%s", app.Config.Path, app.Config.Filename)
	data, err := ini.Load(fullPath)

	if err != nil {
		return app.createDefaultConfig(fullPath)
	}

	app.Config.BindIP = data.Section("conf").Key("ip").String()
	app.Config.Port = data.Section("conf").Key("port").MustInt(22)
	app.Config.RunningMode = ServerMode(data.Section("conf").Key("mode").MustInt(0))
	return nil
}

func (app *ServerApp) createDefaultConfig(fullPath string) error {
	if os.MkdirAll(app.Config.Path, 0655) != nil {
		return fmt.Errorf("Error creating configuration directory.")
	}

	file, err := os.OpenFile(fullPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0655)
	if err != nil {
		return fmt.Errorf("Error creating %s file.", fullPath)
	}
	defer file.Close()

	_, err = file.Write([]byte("[conf]\n# Set the IP address of the network interface where you want to run the honeypot; you can keep 0.0.0.0\nip = 0.0.0.0\n\n# Port where the honeypot will run; default is 22 since it simulates SSH and is more convincing\nport = 22\n\n# Modes range from 0 to 3 with the following behaviors:\n# mode 0: extreme mode; bans the IP at any sign of connection to the honeypot port\n# mode 1: deceives scanners by behaving like a legitimate SSH server, but blocks the IP upon connection attempt\n# mode 2: allows access to a fake shell and blocks the IP after disconnection\n# mode 3: similar to mode 2, but does not ban; attempts to use social engineering to compromise the attacker's system\nmode = 0\n"))

	if err != nil {
		return fmt.Errorf("error writing to file.d")
	}
	return fmt.Errorf("Configuration file created. Restart the application.")
}

func (app *ServerApp) checkPrivileges() {
	if os.Getegid() == 0 {
		color.Set(color.FgYellow)
		log.Println("Warning, you are running as a superuser; I recommend using a separate user account for this purpose.")
		color.Unset()
	}
}

func (app *ServerApp) startListening() {
	address := fmt.Sprintf("%s:%d", app.Config.BindIP, app.Config.Port)
	listener, err := net.Listen("tcp", address)

	if err != nil {
		app.printFatalError(err)
		os.Exit(1)
	}
	defer listener.Close()

	app.printStartupBanner(address)
	app.acceptConnections(listener)
}

func (app *ServerApp) acceptConnections(listener net.Listener) {
	for {
		connection, err := listener.Accept()
		if err != nil {
			color.Set(color.FgRed)
			log.Println("Connection refused.")
			color.Unset()
			continue
		}

		app.routeConnection(connection)
	}
}

func (app *ServerApp) routeConnection(connection net.Conn) {
	if app.Config.RunningMode == HighProtection {
		shield := Shield{Connection: connection, Logger: app.Logger}
		go shield.Activate()
		return
	}

	trap := Trap{
		Connection: connection,
		Mode:       app.Config.RunningMode,
		Signer:     app.Signer,
		Logger:     app.Logger,
	}

	go func() {
		err := trap.HandleConnection()
		if err != nil {
			color.Set(color.FgRed)
			log.Println(err)
			color.Unset()
		}
	}()
}

func (app *ServerApp) printStartupBanner(address string) {
	color.Set(color.FgGreen)
	log.Println(app.Messages.GetMessage(app.Config.RunningMode))
	color.Set(color.FgYellow)
	log.Printf("running on IP: %s\n", address)
	color.Unset()
}

func (app *ServerApp) printFatalError(err error) {
	color.Set(color.FgRed)
	log.Println(err)
	color.Unset()
}

func (app *ServerApp) handleConfigError(err error) {
	color.Set(color.FgRed)
	log.Println("Something went wrong.")
	log.Println(err)
	color.Set(color.FgYellow)
	log.Printf("\n\nThe first execution must be done with the root user.\n\n")
	color.Unset()
}
