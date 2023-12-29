package main

import (
	"bufio"
	"encoding/json"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/bitrise-io/go-steputils/stepconf"
	"github.com/bitrise-io/go-utils/log"
	"golang.org/x/mod/semver"
)

type Runtime struct {
	BundlePath   string `json:"bundlePath"`
	BuildVersion string `json:"buildversion"`
	Platform     string `json:"platform"`
	RuntimeRoot  string `json:"runtimeRoot"`
	Identifier   string `json:"identifier"`
	Version      string `json:"version"`
	IsInternal   bool   `json:"isInternal"`
	IsAvailable  bool   `json:"isAvailable"`
	Name         string `json:"name"`
}

type SimulatorInfo struct {
	Runtimes []Runtime `json:"runtimes"`
}

func SplitLines(s string) []string {
	var lines []string
	sc := bufio.NewScanner(strings.NewReader(s))
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines
}

func RemoveQuote(s string) string {
	s = strings.ReplaceAll(s, "(", "")
	s = strings.ReplaceAll(s, ")", "")
	return s
}

func FindVersion(info SimulatorInfo, buildVersion string) string {
	for _, runtime := range info.Runtimes {
		if runtime.BuildVersion == buildVersion {
			return runtime.Version
		}
	}
	return ""
}

func RemoveSimulator(udid string) {
	_, err := exec.Command("xcrun", "simctl", "delete", udid).Output()
	if err != nil {
		failf(err.Error())
	}
}

// ConfigsModel ...
type ConfigsModel struct {
	RemoveVersionsLowerThan  string `env:"remove_versions_lower_than"`
	RemoveVersionsHigherThan string `env:"remove_versions_higher_than"`
	Platform                 string `env:"platform"`
}

func createConfigsModelFromEnvs() (ConfigsModel, error) {
	var c ConfigsModel
	if err := stepconf.Parse(&c); err != nil {
		return ConfigsModel{}, err
	}

	return c, nil
}

func failf(format string, v ...interface{}) {
	log.Errorf(format, v...)
	os.Exit(1)
}

func main() {
	configs, err := createConfigsModelFromEnvs()
	if err != nil {
		failf(err.Error())
	}
	stepconf.Print(configs)

	lower := configs.RemoveVersionsLowerThan
	upper := configs.RemoveVersionsHigherThan
	platform := configs.Platform
	vUpper := "v" + upper
	vLower := "v" + lower
	if !semver.IsValid(vUpper) {
		failf("Invalid version:" + upper)
	}
	if !semver.IsValid(vLower) {
		failf("Invalid version:" + lower)
	}
	// Clean up unused simulators
	_, err = exec.Command("xcrun", "simctl", "delete", "unavailable").Output()
	if err != nil {
		failf(err.Error())
	}
	// Read and store runtimes
	output, err := exec.Command("xcrun", "simctl", "list", "runtimes", "--json").Output()
	if err != nil {
		failf(err.Error())
	}
	var info SimulatorInfo
	err = json.Unmarshal(output, &info)
	if err != nil {
		failf(err.Error())
	}
	// Parse device UDID and their assigned runtimes
	output, err = exec.Command("xcrun", "simctl", "list", "devices", "-v").Output()
	if err != nil {
		failf(err.Error())
	}
	lines := SplitLines(string(output))
	currentVersion := ""
	currentPlatform := ""
	for _, line := range lines {
		if strings.HasPrefix(line, "==") {
			continue
		}
		if strings.HasPrefix(line, "--") {
			r, _ := regexp.Compile(`\([a-zA-Z0-9]+\)`)
			buildVersion := RemoveQuote(r.FindString(line))
			currentVersion = "v" + FindVersion(info, buildVersion)
			r, _ = regexp.Compile(`^--\ [a-zA-Z]+`)
			currentPlatform = strings.ReplaceAll(r.FindString(line), "-- ", "")
		} else {
			if currentPlatform != platform {
				continue
			}
			r, _ := regexp.Compile(`\([a-zA-Z0-9\-]{36}\)`)
			udid := RemoveQuote(r.FindString(line))
			result := semver.Compare(currentVersion, vLower)
			if result < 0 {
				log.Infof("Removed:" + line)
				RemoveSimulator(udid)
			}
			result = semver.Compare(currentVersion, vUpper)
			if result > 0 {
				log.Infof("Removed:" + line)
				RemoveSimulator(udid)
			}
		}
	}
	log.Infof("Completed")
}
