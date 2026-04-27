package main

import (
	"log"
	"net"
	"time"

	"github.com/fatih/color"
)

type Shield struct {
	Connection net.Conn
	Logger     BanLogger
}

func (shield Shield) Activate() error {
	defer shield.Connection.Close()
	shield.Connection.SetDeadline(time.Now().Add(5 * time.Second))

	clientIP := IPAddress{Address: shield.Connection.RemoteAddr()}.ExtractIP()

	err := shield.Logger.RegisterBan(clientIP)
	if err != nil {
		shield.printError(err)
		return err
	}

	shield.printSuccess(clientIP)
	return nil
}

func (shield Shield) printError(err error) {
	color.Set(color.FgWhite, color.Bold)
	color.BgRGB(100, 0, 0).Println(err)
	color.Unset()
}

func (shield Shield) printSuccess(ip net.IP) {
	color.Set(color.FgMagenta)
	log.Printf("IP: %s was successfully blocked.\n", ip)
	color.Unset()
}
