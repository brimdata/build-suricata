// This tool executes suricata-update on windows.  It embeds knowledge of the locations of the suricata
// executable and suricata paths path in the expanded 'zdeps/suricata'
// directory inside a Brim installation.
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// These paths are relative to the zdeps/suricata directory.
var (
	execRelPath = "bin/suricata-update.exe"
)

// zdepsSuricataDirectory returns the absolute path of the zdeps/suricata directory,
// based on the assumption that this executable is located directly in it.
func zdepsSuricataDirectory() (string, error) {
	execFile, err := os.Executable()
	if err != nil {
		return "", err
	}

	return filepath.Dir(execFile), nil
}

func makeConfig(installDir, dataDir, dest string) error {
	ruleConfig := fmt.Sprintf(`
data-directory: %s
dist-rule-directory: %s\share\suricata\rules
`, dataDir, installDir)

	return ioutil.WriteFile(filepath.Join(dataDir, dest), []byte(ruleConfig), 0644)
}

func runSuricataUpdate(installDir, dataDir, execPath string) error {
	args := append([]string{
		"--suricata", filepath.Join(installDir, "bin/suricata.exe"),
		"--config", filepath.Join(dataDir, "update.yaml"),
		"--suricata-conf", filepath.Join(installDir, "brim-conf.yaml"),
		"--no-test",
		"--no-reload",
	}, os.Args[1:]...)
	cmd := exec.Command(execPath, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	path := fmt.Sprintf("PATH=%s;%s", filepath.Join(installDir, "dlls"), os.Getenv("PATH"))
	cmd.Env = append(os.Environ(), path)

	return cmd.Run()
}

func main() {
	installDir, err := zdepsSuricataDirectory()
	if err != nil {
		log.Fatalln("zdepsSuricataDirectory failed:", err)
	}
	dataDir := os.Getenv("BRIM_SURICATA_USER_DIR")
	if dataDir == "" {
		dataDir = filepath.Join(installDir, "var", "lib", "suricata")
	}
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalln("os.Mkdir %s failed:", dataDir, err)
	}

	if err := makeConfig(installDir, dataDir, "update.yaml"); err != nil {
		log.Fatalln("makeConfig failed:", err)
	}

	execPath := filepath.Join(installDir, filepath.FromSlash(execRelPath))
	if _, err := os.Stat(execPath); err != nil {
		log.Fatalln("suricata-update executable not found at", execPath)
	}

	err = runSuricataUpdate(installDir, dataDir, execPath)
	if err != nil {
		log.Fatalln("launchSuricata failed", err)
	}
}
