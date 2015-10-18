package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/kouhin/envflag"
)

var (
	previousRead     int
	previousReadTime time.Time
)

type MeterRead struct {
	Time    *time.Time
	Offset  int
	Length  int
	Message *MeterMessage
}

type MeterMessage struct {
	ID          int
	Type        int
	TamperPhy   int
	TamperEnc   int
	Consumption int
	ChecksumVal int
}

func main() {
	var (
		meterId = flag.String("meter-id", "REQUIRED", "ID of the meter to read from")
	)
	if err := envflag.Parse(); err != nil {
		panic(err)
	}

	if *meterId == "REQUIRED" {
		fmt.Println("meter-id is a required field")
		os.Exit(1)
	}

	cmdName := "rtlamr"
	cmdArgs := []string{"-format=json", "-filterid=" + *meterId}

	cmd := exec.Command(cmdName, cmdArgs...)
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			text := scanner.Text()
			line := &MeterRead{}
			if err := json.Unmarshal([]byte(text), &line); err != nil {
				fmt.Fprintln(os.Stderr, fmt.Sprintf("Error unmarshaling line (%s):| %s", err, text))
				continue
			}
			fmt.Printf("line | %q\n", line)
		}
	}()

	err = cmd.Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error starting Cmd", err)
		os.Exit(1)
	}

	err = cmd.Wait()
	if err != nil {
		time.Sleep(2000) // let the scanner finish. there's gotta be a better way
		fmt.Fprintln(os.Stderr, "Command exited", err)
		os.Exit(1)
	}
}

func processLine(read *MeterRead) {
	if previousRead == 0 {
		// we have no baseline, set and wait for more
		previousRead = read.Message.Consumption
		previousReadTime = read.Time
		return
	}

}

func reportMetric() {
}
