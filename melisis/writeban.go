package main

import (
	"fmt"
	"net"
	"os"
)

type BanLogger struct {
	FilePath string
}

func (logger BanLogger) RegisterBan(ip net.IP) error {
	file, err := logger.openLogFile()
	if err != nil {
		return err
	}
	defer file.Close()

	return logger.writeRecord(file, ip)
}

func (logger BanLogger) openLogFile() (*os.File, error) {
	file, err := os.OpenFile(logger.FilePath, os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return nil, fmt.Errorf("Error opening lockout log.")
	}
	return file, nil
}

func (logger BanLogger) writeRecord(file *os.File, ip net.IP) error {
	record := fmt.Sprintf("[BAN_MELISIS] IP: %s\n", ip)
	if _, err := file.WriteString(record); err != nil {
		return fmt.Errorf("Error creating block log for IP: %s", ip)
	}
	return nil
}
