package utils

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"os/exec"
)

const systemdServiceTemplate = `
[Unit]
Description=My Go Application
After=network.target

[Service]
ExecStart={{.ExecPath}} {{range .Args}} {{ . }} {{end}}
Restart=always
User={{.User}}

[Install]
WantedBy=multi-user.target
`

type ServiceConfig struct {
	ExecPath string
	User     string
	Args     []string
}

func createSystemdServiceFile(config ServiceConfig, serviceName string) (string, error) {
	tmpl, err := template.New("kmagentservice").Parse(systemdServiceTemplate)
	if err != nil {
		return "", err
	}

	serviceFilePath := fmt.Sprintf("/etc/systemd/system/%s.service", serviceName)
	file, err := os.Create(serviceFilePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if err := tmpl.Execute(file, config); err != nil {
		return "", err
	}

	return serviceFilePath, nil
}

func enableAndStartSystemdService(serviceName string) error {
	cmds := []string{
		"systemctl daemon-reload",
		fmt.Sprintf("systemctl enable %s.service", serviceName),
		fmt.Sprintf("systemctl start %s.service", serviceName),
	}

	for _, cmd := range cmds {
		if err := exec.Command("sh", "-c", cmd).Run(); err != nil {
			return err
		}
	}

	return nil
}

func StartServiceForLinux(flags []string) {
	execPath := getExecPath()

	config := ServiceConfig{
		ExecPath: execPath,
		User:     "root",
		Args:     flags,
	}

	serviceFilePath, err := createSystemdServiceFile(config, SERVICE_NAME)
	if err != nil {
		log.Println("Error creating systemd service file:", err)
		return
	}

	log.Println("Service file created at:", serviceFilePath)

	if err := enableAndStartSystemdService(SERVICE_NAME); err != nil {
		log.Println("Error enabling and starting systemd service:", err)
		return
	}

	log.Println("Service loaded successfully.")
}
