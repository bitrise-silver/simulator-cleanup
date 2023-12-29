package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

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
		log.Fatal(err)
		os.Exit(1)
	}
}

func main() {
	lower := flag.String("l", "0.0", "Remove simulators with versions lower than this")
	upper := flag.String("u", "99.0", "Remove simulators with versions higher than this")
	platform := flag.String("p", "iOS", "Remove simulators with platform name")
	flag.Parse()
	vUpper := "v" + *upper
	vLower := "v" + *lower
	if !semver.IsValid(vUpper) {
		fmt.Fprintf(os.Stderr, "Invalid version:"+*upper)
		os.Exit(1)
	}
	if !semver.IsValid(vLower) {
		fmt.Fprintf(os.Stderr, "Invalid version:"+*lower)
		os.Exit(1)
	}
	// Clean up unused simulators
	_, err := exec.Command("xcrun", "simctl", "delete", "unavailable").Output()
	if err != nil {
		log.Fatal(err)
	}
	// Read and store runtimes
	output, err := exec.Command("xcrun", "simctl", "list", "runtimes", "--json").Output()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	var info SimulatorInfo
	err = json.Unmarshal(output, &info)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	// Parse device UDID and their assigned runtimes
	output, err = exec.Command("xcrun", "simctl", "list", "devices", "-v").Output()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
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
			if currentPlatform != *platform {
				continue
			}
			r, _ := regexp.Compile(`\([a-zA-Z0-9\-]{36}\)`)
			udid := RemoveQuote(r.FindString(line))
			result := semver.Compare(currentVersion, vLower)
			if result <= 0 {
				fmt.Println("Removed:" + line)
				RemoveSimulator(udid)
			}
			result = semver.Compare(currentVersion, vUpper)
			if result >= 0 {
				fmt.Println("Removed:" + line)
				RemoveSimulator(udid)
			}
		}
	}
}
