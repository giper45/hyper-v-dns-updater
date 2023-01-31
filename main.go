package main

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"regexp"
	"strings"
	"fmt"
	"github.com/abdfnx/gosh"
)

const HOSTS_FILE string = "C:\\Windows\\System32\\drivers\\etc\\hosts"

func GetVmInfo() (error, string, string) {
	return gosh.PowershellOutput(`get-vm | ?{$_.State -eq "Running"} | select -ExpandProperty networkadapters | select vmname, ipaddresses | ConvertTo-Json`)
}

func GetMap(input string) map[string]string {
	result := make(map[string]string)
	lines := strings.Split(input, "\r\n")

	for _, line := range lines {
		fields := strings.Fields(line)
		name := fields[0]
		ip := ""
		re := regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`)
		matches := re.FindStringSubmatch(line)
		if len(matches) > 0 {
			ip = matches[0]
		}
		result[name] = ip
	}
	return result
}

func HandleErr(err error, errout string) {
	if err != nil {
		log.Printf("error: %v\n", err)
		log.Println(errout)
		os.Exit(-1)
	}
}

type Vm struct {
	VmName      string
	IPAddresses [2]string
}

func updateOrAddLine(fileName string, newLine string) error {
	file, err := os.OpenFile(fileName, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var lines []string
	updated := false
	newFields := strings.Fields(newLine)
	// newIP := newFields[0]
	newName := newFields[1]
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		// ip := fields[0]
		name := fields[1]
		if name == newName {
			lines = append(lines, newLine)
			updated = true
		} else {
			lines = append(lines, line)
		}
	}
	if !updated {
		lines = append(lines, newLine)
	}

	file.Truncate(0)
	file.Seek(0, 0)
	writer := bufio.NewWriter(file)
	for _, line := range lines {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}
	writer.Flush()
	return nil
}

func main() {
	err, output, errout := GetVmInfo()
	HandleErr(err, errout)

	if !strings.HasPrefix(output, "[") || !strings.HasSuffix(output, "]") {
		output = "[" + output + "]"
	}
	
	var machines []Vm
	json.Unmarshal([]byte(output), &machines)
	fmt.Println("Machines: ", machines)
	for _, v := range machines {
		line := v.IPAddresses[0] + "\t" + v.VmName
		err := updateOrAddLine(HOSTS_FILE, line)
		HandleErr(err, "")
	}
}
