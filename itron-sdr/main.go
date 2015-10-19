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
	"github.com/lstoll/go-librato"
)

var (
	previousRead     int
	previousReadTime time.Time
	gauge            chan interface{}
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
		meterId       = flag.String("meter-id", "REQUIRED", "ID of the meter to read from")
		libratoUser   = flag.String("librato-user", "REQUIRED", "User for Librato")
		libratoToken  = flag.String("librato-token", "REQUIRED", "Token for Librato")
		libratoSource = flag.String("librato-source", "REQUIRED", "Source name for librato")
		libratoMetric = flag.String("librato-metric", "REQUIRED", "Metric name for librato")
	)
	if err := envflag.Parse(); err != nil {
		panic(err)
	}
	if *meterId == "REQUIRED" {
		fmt.Println("meter-id is a required field")
		os.Exit(1)
	}
	if *libratoUser == "REQUIRED" {
		fmt.Println("librato-usser is a required field")
		os.Exit(1)
	}
	if *libratoToken == "REQUIRED" {
		fmt.Println("librato-token is a required field")
		os.Exit(1)
	}
	if *libratoSource == "REQUIRED" {
		fmt.Println("librato-source is a required field")
		os.Exit(1)
	}
	if *libratoSource == "REQUIRED" {
		fmt.Println("librato-metric is a required field")
		os.Exit(1)
	}

	metrics := librato.NewSimpleMetrics(*libratoUser, *libratoToken, *libratoSource)
	defer metrics.Wait()
	defer metrics.Close()
	gauge = metrics.GetGauge(*libratoMetric)

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
			fmt.Printf("received | %q\n", line)
			processLine(line)
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
		previousReadTime = *read.Time
		return
	}

	timeBetweenReads := read.Time.Sub(previousReadTime)
	if timeBetweenReads.Hours() < 1.0 {
		// not long enough, bail out
		return
	}

	// it's been more than an hour, so we can report data. Compare the two readings, then factor in
	// that the time might have been more than an hour to get a cf/hr measurement.
	consumed := float64(read.Message.Consumption - previousRead)
	minsBetweenReads := timeBetweenReads.Minutes()
	consumedPerMin := consumed / minsBetweenReads
	consumedPerHour := consumedPerMin * 60
	reportMetric(consumedPerHour)
	// and start again.
	previousRead = read.Message.Consumption
	previousReadTime = *read.Time
}

func reportMetric(cfHr float64) {
	gauge <- cfHr
	fmt.Printf("At %s, consumption rate is %f cf/hr\n", time.Now().Format("2006-01-02T15:04:05.999999-07:00"), cfHr)
}
