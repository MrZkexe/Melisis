package main

import (
	"net"
)

type ServerMode int

const (
	HighProtection ServerMode = iota
	FakeSSH
	FakeShellMode
	Troll
)

type SystemMessages struct {
	messages map[ServerMode]string
}

func NewSystemMessages() SystemMessages {
	return SystemMessages{
		messages: map[ServerMode]string{
			HighProtection: "Started in  highprotection mode.",
			FakeSSH:        "Started in  fakessh mode.",
			FakeShellMode:  "Started in  fakeshell mode.",
			Troll:          "Started in  troll mode.",
		},
	}
}

func (sm SystemMessages) GetMessage(mode ServerMode) string {
	return sm.messages[mode]
}

type IPAddress struct {
	Address net.Addr
}

func (ip IPAddress) ExtractIP() net.IP {
	tcpAddress, ok := ip.Address.(*net.TCPAddr)
	if !ok {
		return nil
	}
	return tcpAddress.IP
}
