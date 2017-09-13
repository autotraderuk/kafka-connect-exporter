package hatchet

import (
	"os"
	"path"
)

// AppInfo appends application information to a log. These include pid, process
// name, and hostname.
func AppInfo(logger Logger) Logger {
	fields := L{
		PID: os.Getpid(),
	}
	if process := getProcess(); process != "" {
		fields[Process] = process
	}
	if hostname := getHostname(); hostname != "" {
		fields[Hostname] = hostname
	}
	return Fields(logger, fields, true)
}

func getProcess() string {
	if len(os.Args) == 0 || os.Args[0] == "" {
		return ""
	}
	return path.Base(os.Args[0])
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return ""
	}
	return hostname
}
