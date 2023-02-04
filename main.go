package main

import (
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type parameter struct {
	name  string
	value string
}

func main() {
	log.Println("Start Reading yaml file")
	yamlMap := readYaml("conf.yaml")
	log.Println("Configuration content is OK")
	log.Println(yamlMap)
}

func readYaml(fileName string) map[string][]parameter {
	yamlContent := getYamlContent(fileName)
	yamlMap := findSections(yamlContent)
	server, serverFound := yamlMap["server"]
	if serverFound == false {
		stopWithError("server section must be exist")
	}
	validateServer(server)
	hostParameter := findParameter(server, "host")
	if hostParameter.name == "" {
		yamlMap["server"] = append(yamlMap["server"], parameter{"host", "localhost"})
	}
	rateLimits, rateLimitsFound := yamlMap["rate_limits"]
	if rateLimitsFound == true {
		validateRateLimit(rateLimits)
	}
	return yamlMap
}
func findParameter(limits []parameter, pname string) parameter {
	for _, limit := range limits {
		if limit.name == pname {
			return limit
		}
	}
	return parameter{}
}

func validateRateLimit(limits []parameter) {
	ipParameter := findParameter(limits, "ip_requests_per_sec")
	if ipParameter.name == "" {
		stopWithError("`ip_requests_per_sec` parameter is required in `rate_limits` section")
	} else {
		ipValue, err := strconv.Atoi(ipParameter.value)
		if err != nil {
			stopWithError("`ip_requests_per_sec` parameter is not a number")
		}
		if ipValue > 1000 || ipValue < 60 {
			stopWithError("`ip_requests_per_sec` parameter must be <1000 and >60")
		}
	}
	otpParameter := findParameter(limits, "otp_sms_interval_sec")
	if otpParameter.name != "" {
		otpValue, err := strconv.Atoi(otpParameter.value)
		if err != nil {
			stopWithError("`otp_sms_interval_sec` parameter is not a number")
		}
		if otpValue > 300 || otpValue <= 60 {
			stopWithError("`otp_sms_interval_sec` parameter must be <300 and >=60")
		}
	}
}

func validateServer(server []parameter) {
	portParameter := findParameter(server, "port")
	if portParameter.name == "" {
		stopWithError("port parameter is required in server section")
	}
	portRegex := regexp.MustCompile(`([1-9][0-9]{0,3}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-5])`)
	if portRegex.ReplaceAllString(portParameter.value, "") != "" {
		stopWithError("port parameter is not valid number")
	}
}
func getYamlContent(fileName string) string {
	fileContent, err := os.ReadFile(fileName)
	if err != nil {
		stopWithError(fileName + " file not found")
	}
	return regexp.MustCompile(`(?m)#(.*)`).ReplaceAllString(string(fileContent), "")
}

func findSections(inputString string) map[string][]parameter {
	var sectionRegex = regexp.MustCompile(`([a-z_\d]+ *: *\n)( {2,}[a-z_\d]+ *: *(.*)\n*)*`)
	var titleRegex = regexp.MustCompile(`([a-z_0-9]+) *: *\n`)
	var subSectionRegex = regexp.MustCompile(` {2,}([a-z_0-9]+ *: *(.*)\n{0,})*`)
	var title string
	var splitParameter []string
	all := make(map[string][]parameter)
	sections := sectionRegex.FindAllString(inputString, -1)
	for _, section := range sections {
		title = trimSection(titleRegex.FindString(section))
		for _, subSection := range subSectionRegex.FindAllString(section, -1) {
			splitParameter = strings.Split(subSection, ":")
			all[title] = append(all[title], parameter{trimSection(splitParameter[0]), trimSection(splitParameter[1])})
		}
	}
	return all
}
func trimSection(str string) string {
	return strings.Trim(str, ":\n ")
}

func stopWithError(message string) {
	log.Fatal(message)
	panic(message)
}
