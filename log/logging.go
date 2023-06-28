package log

import "os"

func WriteToLog(message string) {
	f, err := os.OpenFile("file.txt",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		println(err)
	}
	defer f.Close()
	if _, err := f.WriteString(message + "\n"); err != nil {
		println(err)
	}
}
