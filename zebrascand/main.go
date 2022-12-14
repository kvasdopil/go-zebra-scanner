package main

import (
	"encoding/json"
	"fmt"
	"github.com/elemecca/go-zebra-scanner/devices"
	"github.com/elemecca/go-zebra-scanner/snapi"
	log "github.com/sirupsen/logrus"
	"os"
)

func isASCIIPrintable(s []byte) bool {
	for _, c := range s {
		if c < 32 || c > 126 {
			return false
		}
	}
	return true
}

type UserInput struct {
	value int
}

func readInput(input chan<- interface{}) {
	for {
			var value int
			_, err := fmt.Scanf("%d\n", &value)
			if err != nil {
					panic(err)
			}
			// fmt.Println("read cmd", u)
			input <- UserInput{value}
	}
}

func main() {
	if os.Getenv("ZSDEBUG") != "" {
		log.SetLevel(log.DebugLevel)
		log.SetFormatter(&debugTextFormatter{&log.TextFormatter{}})
	}

	eventChan := make(chan interface{}, 10)
	dm, err := devices.NewDeviceManager(eventChan)
	if err != nil {
		log.Error("failed to start device manager: ", err)
		os.Exit(2)
	}
	defer dm.Close()

	go readInput(eventChan)

	for {
		switch event := (<-eventChan).(type) {
		case snapi.ScanEvent:
			msg := map[string]interface{}{
				"event": "scan",
				"type":  event.PrimaryType,
			}

			if isASCIIPrintable(event.PrimaryData) {
				msg["text"] = string(event.PrimaryData)
			} else {
				msg["data"] = event.PrimaryData
			}

			if event.SupplementalType != "" {
				sup := map[string]interface{}{
					"type": event.SupplementalType,
				}

				if isASCIIPrintable(event.SupplementalData) {
					sup["text"] = string(event.SupplementalData)
				} else {
					sup["data"] = event.SupplementalData
				}

				msg["supplemental"] = sup
			}

			code, _ := json.Marshal(msg) 
			fmt.Println(string(code[:]))

		case UserInput:
			fmt.Println("userinput", event.value)
			fmt.Println(dm)

		default:
			log.Debug(fmt.Sprintf("main loop received unknown event %T", event))
		}
	}
}
