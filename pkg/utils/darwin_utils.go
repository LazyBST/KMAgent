package utils

import (
	"html/template"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func getPlistTemplate() string {
	var template = `<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
	<plist version="1.0">
	<dict>
		<key>Label</key>
		<string>{{.Label}}</string>
		<key>ProgramArguments</key>
		<array>
			<string>{{.ExecPath}}</string>
			{{range .Args}}
        		<string>{{.}}</string>
        	{{end}}
		</array>
		<key>RunAtLoad</key>
		<true/>
		<key>KeepAlive</key>
		<true/>
		<key>StandardOutPath</key>
		<string>/tmp/{{.Label}}.out</string>
		<key>StandardErrorPath</key>
		<string>/tmp/{{.Label}}.err</string>
	</dict>
	</plist>
	`
	return template
}

type PlistConfig struct {
	Label    string
	ExecPath string
	Args     []string
}

func createPlistFile(config PlistConfig) (string, error) {
	t := template.New("kmagentplist")
	t, err := t.Parse(getPlistTemplate())
	if err != nil {
		return "", err
	}

	plistPath := filepath.Join(os.Getenv("HOME"), "Library", "LaunchAgents", config.Label+".plist")
	file, err := os.Create(plistPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if err := t.Execute(file, config); err != nil {
		return "", err
	}

	return plistPath, nil
}

func loadService(plistPath string) error {
	cmd := exec.Command("launchctl", "load", plistPath)
	return cmd.Run()
}

func StartServiceForDarwin(flags []string) {
	execPath := getExecPath()

	config := PlistConfig{
		Label:    SERVICE_NAME,
		ExecPath: execPath,
		Args:     flags,
	}

	plistPath, err := createPlistFile(config)
	if err != nil {
		log.Println("Error creating plist file:", err)
		return
	}

	log.Println("Plist file created at:", plistPath)

	if err := loadService(plistPath); err != nil {
		log.Println("Error loading service:", err)
		return
	}

	log.Println("Service loaded successfully.")
}
