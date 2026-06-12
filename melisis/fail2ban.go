package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
)

type DependencyManager struct {
	Fail2BanBinary string
	LogFilePath    string
	LogCommandPath string
}

func (manager DependencyManager) CheckAndInstall() error {
	if manager.binaryMissing() {
		return fmt.Errorf("Fail2ban was not found, please install.")
	}

	return manager.createConfigurationFiles()
}

func (manager DependencyManager) binaryMissing() bool {
	_, err := os.Stat(manager.Fail2BanBinary)
	return os.IsNotExist(err)
}

func (manager DependencyManager) createConfigurationFiles() error {
	files := manager.getRequiredFiles()

	for filepath, content := range files {
		err := manager.ensureFileExists(filepath, content)
		if err != nil {
			return err
		}
	}
	return nil
}

func (manager DependencyManager) ensureFileExists(filepath string, content string) error {
	_, err := os.Stat(filepath)
	if !os.IsNotExist(err) {
		return nil
	}

	color.Set(color.FgCyan)
	log.Printf("Creating file: %s\n", filepath)
	color.Unset()

	return manager.writeFile(filepath, []byte(content))
}

func (manager DependencyManager) writeFile(filepath string, text []byte) error {
	file, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0655)
	if err != nil {
		return fmt.Errorf("Something went wrong while creating the file: %s", filepath)
	}
	defer file.Close()

	if _, err = file.Write(text); err != nil {
		return fmt.Errorf("Something went wrong while trying to write to the file: %s", filepath)
	}

	manager.setPermissionsIfNeeded(filepath)
	return nil
}

func (manager DependencyManager) setPermissionsIfNeeded(filepath string) {
	if strings.Contains(filepath, "log") {
		color.Set(color.FgCyan)
		os.Chmod(filepath, 0666)
		color.Unset()
	}
}

func (manager DependencyManager) getRequiredFiles() map[string]string {
	jailContent := fmt.Sprintf("[melisis]\nenabled = true\nfilter = melisis\nlogpath = %s\nmaxretry = 1\nfindtime = 5m\nbantime = 5d\naction = iptables-allports[name=MelisisHost]\n	iptables-allports[name=Melisis, chain=DOCKER-USER]", manager.LogFilePath)
	filterContent := "[Definition]\nfailregex = ^.*\\[BAN_MELISIS\\] IP: <HOST>.*$\nignoreregex ="

	return map[string]string{
		"/etc/fail2ban/jail.d/melisis.conf":   jailContent,
		"/etc/fail2ban/filter.d/melisis.conf": filterContent,
		manager.LogFilePath:                   "",
		manager.LogCommandPath:                "",
	}
}
