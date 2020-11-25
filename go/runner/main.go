// This tool executes suricata on windows.  It embeds knowledge of the locations of the suricata
// executable and suricata paths in the expanded 'zdeps/suricata'
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
	confRelPath = "brim-conf.yaml"
	execRelPath = "bin/suricata.exe"
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

func makeConfig(baseDir, source, dest string) error {
	ruleConfig := fmt.Sprintf(`
rule-files:
  - %s\var\lib\suricata\rules\suricata.rules
`, baseDir)

	input, err := ioutil.ReadFile(filepath.Join(baseDir, source))
	if err != nil {
		return err
	}
	input = append(input, []byte(ruleConfig)...)

	return ioutil.WriteFile(filepath.Join(baseDir, dest), input, 0644)
}

func runSuricata(baseDir, execPath string) error {
	cmd := exec.Command(execPath,
		"-c", filepath.Join(baseDir, "brim-conf-run.yaml"),
		"--set", fmt.Sprintf("classification-file=%s", filepath.FromSlash(filepath.Join(baseDir, "/share/suricata/classification.config"))),
		"--set", fmt.Sprintf("reference-config-file=%s", filepath.FromSlash(filepath.Join(baseDir, "/share/suricata/reference.config"))),
		"--set", fmt.Sprintf("threshold-file=%s", filepath.FromSlash(filepath.Join(baseDir, "/share/suricata/threshold.config"))),
		"-r", "-")

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	path := fmt.Sprintf("PATH=%s;%s", filepath.Join(baseDir, "dlls"), os.Getenv("PATH"))
	cmd.Env = append(os.Environ(), path)

	return cmd.Run()
}

func main() {
	baseDir, err := zdepsSuricataDirectory()
	if err != nil {
		log.Fatalln("zdepsSuricataDirectory failed:", err)
	}

	if err := makeConfig(baseDir, "brim-conf.yaml", "brim-conf-run.yaml"); err != nil {
		log.Fatalln("makeConfig failed:", err)
	}

	execPath := filepath.Join(baseDir, filepath.FromSlash(execRelPath))
	if _, err := os.Stat(execPath); err != nil {
		log.Fatalln("suricata executable not found at", execPath)
	}

	err = runSuricata(baseDir, execPath)
	if err != nil {
		log.Fatalln("launchSuricata failed", err)
	}
}
