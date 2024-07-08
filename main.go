package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/distatus/battery"
	"github.com/gen2brain/beeep"
)

func initLogger() (*os.File, error) {
	logFile, err := os.OpenFile("battery_check.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, err
	}
	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	return logFile, nil
}

func getBatteryPercentage() (float64, error) {
	battery, err := battery.Get(0)
	if err != nil {
		return 0, err
	}
	if battery.State.Raw.String() != "Discharging" {
		return 0, nil
	}
	percentage := (battery.Current / battery.Full) * 100
	return percentage, nil
}

func setDuration(value float64) time.Duration {
	switch {
	case value > 60:
		return 20 * time.Minute
	case value > 40 && value < 60:
		return 15 * time.Minute
	case value > 20 && value < 40:
		return 10 * time.Minute
	default:
		return 4 * time.Minute
	}
}

func sendAlert() (time.Duration, error) {
	switch percentage, err := getBatteryPercentage(); err == nil {
	case percentage > 19 && percentage < 21:
		return setDuration(percentage), beeep.Alert(
			"Battery Level",
			"Your battery level is low. It's "+strconv.FormatFloat(percentage, 'f', 0, 64)+"%",
			"assets/information.png",
		)
	case percentage < 10 && percentage > 8:
		return setDuration(percentage), beeep.Alert(
			"Battery Level Warning",
			"PC battery level is very low. Your laptop will shutdown soon. Battery level < 9%",
			"assets/warning.png",
		)
	case percentage < 5 && percentage != 0:
		return setDuration(percentage), beeep.Alert(
			"Shutdown Warning",
			"Critical battery level. Your laptop is about to shutdown. Charge your PC or you'll lose your work.",
			"assets/warning.png",
		)
	default:
		return setDuration(percentage), nil
	}
}

func main() {
	logFile, err := initLogger()

	if err != nil {
		log.Fatalf("Could not initialize log file: %v", err)
	}

	defer logFile.Close()

	for {
		checkTime, err := sendAlert()
		if err != nil {
			log.Printf("Could not send alert: %v", err)
		}
		time.Sleep(checkTime)
	}
}
