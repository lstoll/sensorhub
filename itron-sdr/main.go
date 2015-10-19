package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/kouhin/envflag"
	"github.com/lstoll/go-librato"
)

var (
	previousRead       int
	consumptionCounter chan int64
	costGauge          chan interface{}
	gasPrice           float64
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
		gasPriceIn    = flag.String("gas-price", "OPTIONAL", "Price of gas Will report as additional metric")
	)
	if err := envflag.Parse(); err != nil {
		panic(err)
	}
	gasPriceInFl, err := strconv.ParseFloat(*gasPriceIn, 64)
	if err != nil {
		fmt.Println("gas-price has to be a number")
		os.Exit(1)
	}
	gasPrice = gasPriceInFl
	if *meterId == "REQUIRED" {
		fmt.Println("meter-id is a required field")
		os.Exit(1)
	}
	if *libratoUser == "REQUIRED" {
		fmt.Println("librato-user is a required field")
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

	metrics := librato.NewSimpleMetrics(*libratoUser, *libratoToken, *libratoSource)
	defer metrics.Wait()
	defer metrics.Close()
	consumptionCounter = metrics.GetCounter("sensor.utility.gas.cf")
	costGauge = metrics.GetGauge("sensor.utility.gas.spend")

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
			fmt.Print("reading received: ")
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
	fmt.Printf("at %s, consumption is %d", time.Now().Format("2006-01-02T15:04:05.999999-07:00"), read.Message.Consumption)
	consumptionCounter <- int64(read.Message.Consumption)

	if gasPrice != 0 && previousRead != 0 {
		// we're tracking price, compare to the previous read and multiply by price
		consumedSinceLast := read.Message.Consumption - previousRead
		spentSinceLast := float64(consumedSinceLast) * (gasPrice / 1000)
		fmt.Printf(" and amount spent since last read is %f", spentSinceLast)
		costGauge <- spentSinceLast
	}

	fmt.Println("")

	previousRead = read.Message.Consumption
}
