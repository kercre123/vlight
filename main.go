package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/gorilla/websocket"
	"github.com/nxadm/tail"
)

var podURL string = "192.168.1.4:8080"
var lightsOnEndpoint string = "http://192.168.1.75:8080/lights_on"
var lightsOffEndpoint string = "http://192.168.1.75:8080/lights_off"

var implCommands [][]string = [][]string{
	// implemented commands
	lightOnWords, lightOffWords,
}

var lightOnWords []string = []string{
	"lights on", "light on", "on the light", "on the late",
}

var lightOffWords []string = []string{
	"lights off", "light off", "off the light", "off the late",
}

func LightsOff() {
	time.Sleep(time.Second / 2)
	Behavior("Wait")
	time.Sleep(time.Second / 3)
	PlayAnim("lights_off")
	go func() {
		time.Sleep(time.Second / 3)
		AudioEvent("Play__Robot_Vic_Sfx__Head_Down_Short_Curious")
		time.Sleep(time.Millisecond * 1200)
		PlayAudioOnVictor("/data/light.pcm", false)
		time.Sleep(time.Millisecond * 1300)
		AudioEvent("Play__Robot_Vic_Sfx__Lift_High_Down_Short_Curious")
	}()
	time.Sleep(time.Millisecond * 1600)

	// make lightsoff post request
	POSTreq(lightsOffEndpoint, "")

	time.Sleep(time.Millisecond * 2100)
	Behavior("InitNormalOperation")
}

func LightsOn() {
	time.Sleep(time.Second / 2)
	Behavior("Wait")
	time.Sleep(time.Second / 3)
	PlayAnim("lights_on")
	go func() {
		time.Sleep(time.Second / 3)
		AudioEvent("Play__Robot_Vic_Sfx__Head_Down_Short_Curious")
		time.Sleep(time.Millisecond * 1200)
		PlayAudioOnVictor("/data/light.pcm", false)
		time.Sleep(time.Millisecond * 1300)
		AudioEvent("Play__Robot_Vic_Sfx__Lift_High_Down_Short_Curious")
	}()
	time.Sleep(time.Millisecond * 1600)

	// make lightson post request
	POSTreq(lightsOnEndpoint, "")

	time.Sleep(time.Millisecond * 2100)
	Behavior("InitNormalOperation")
}

func DoAction(queryText string) {
	skipLines = true
	defer skipLinesFalse()
	if contains(queryText, lightOffWords) {
		LightsOff()
	} else if contains(queryText, lightOnWords) {
		LightsOn()
	}
}

var implWords []string
var skipLines bool

var victor *vector.Vector

var logFile = "/var/log/messages"

func main() {
	if !VerifyThisIsAVector() {
		fmt.Println("This program is meant to be run internally on a Vector robot.")
		os.Exit(1)
	}

	// Init SDK (requires wire-pod availability)
	InitVector()
	// Communicate locally
	victor.Cfg.Target = "127.0.0.1:443"

	// Create list of command words
	for _, command := range implCommands {
		implWords = append(implWords, command...)
	}
	// look at log file. in for loop in case of deletion of /var/log/messages
	for {
		t, err := tail.TailFile(logFile, tail.Config{
			Follow:   true,
			Location: &tail.SeekInfo{Offset: 0, Whence: io.SeekEnd}, // <- line changed
		})
		if err != nil {
			fmt.Println(err)
		}

		for line := range t.Lines {
			if !skipLines {
				//  01-21 19:43:03.461 info logwrapper 4291 4291 vic-cloud: Intent response -> query_text:"turn the light on" action:"intent_lights_on"
				if strings.Contains(line.Text, "Intent response -> ") {
					splitString := strings.Split(strings.TrimSpace(strings.Split(line.Text, "Intent response -> ")[1]), " action:")
					queryText := strings.TrimSuffix(strings.TrimPrefix(splitString[0], "query_text:\""), "\"")
					if contains(queryText, implWords) {
						fmt.Println(queryText)
						// actually do the action
						go DoAction(queryText)
					}
				}
			}
		}
		time.Sleep(time.Second / 2)
	}
}

func contains(substring string, stringsList []string) bool {
	for _, word := range stringsList {
		if strings.Contains(substring, word) {
			return true
		}
	}
	return false
}

func skipLinesFalse() {
	skipLines = false
}

func POSTreq(URL string, content string) {
	resp, err := http.Post(URL, "application/x-www-form-urlencoded", bytes.NewBuffer([]byte(content)))
	if err != nil {
		fmt.Println(err)
	}
	io.ReadAll(resp.Body)
}

func AudioEvent(event string) {
	POSTreq("http://localhost:8889/consolefunccall", "func=PostAudioEvent&args="+event)
}

func PlayAnim(anim string) {
	POSTreq("http://localhost:8889/consolefunccall", "func=PlayAnimation&args="+anim)
}

func GetESN() string {
	out, _ := exec.Command("/bin/emr-cat", "e").Output()
	return string(out)
}

func InitVector() {
	var err error
	var i int
	for {
		i = i + 1
		if i == 6 {
			fmt.Println("Vector SDK conn didn't work after 5 tries, exiting")
			os.Exit(1)
		}
		victor, err = NewWpExternal(podURL, GetESN())
		if err != nil {
			fmt.Println(err)
			time.Sleep(time.Second)
			continue
		} else {
			return
		}
	}
}

type BehaviorMessage struct {
	Type   string `json:"type"`
	Module string `json:"module"`
	Data   struct {
		BehaviorName     string `json:"behaviorName"`
		PresetConditions bool   `json:"presetConditions"`
	} `json:"data"`
}

//{"type":"data","module":"behaviors","data":{"behaviorName":"DevBaseBehavior","presetConditions":false}}

func Behavior(behavior string) {
	u := url.URL{Scheme: "ws", Host: "localhost:8888", Path: "/socket"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	message := BehaviorMessage{
		Type:   "data",
		Module: "behaviors",
		Data: struct {
			BehaviorName     string `json:"behaviorName"`
			PresetConditions bool   `json:"presetConditions"`
		}{
			BehaviorName:     behavior,
			PresetConditions: false,
		},
	}

	marshaledMessage, err := json.Marshal(message)
	if err != nil {
		log.Fatal("marshal:", err)
	}

	err = c.WriteMessage(websocket.TextMessage, marshaledMessage)
	if err != nil {
		log.Fatal("write:", err)
	}
}

func VerifyThisIsAVector() bool {
	out, _ := exec.Command("uname", "-a").Output()
	return strings.Contains(string(out), "Vector-")
}
