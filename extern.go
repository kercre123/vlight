package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/digital-dream-labs/hugh/grpc/client"
	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
)

type Option func(*options)

// options holds the options for the vector robot.
type options struct {
	SerialNo  string
	RobotName string `ini:"name"`
	CertPath  string `ini:"cert"`
	Target    string `ini:"ip"`
	Token     string `ini:"guid"`
}

// WithTarget sets the ip of the vector robot.
func WithTarget(s string) Option {
	return func(o *options) {
		if len(s) > 0 {
			o.Target = s
		}
	}
}

// WithToken set the token for the vector robot.
func WithToken(s string) Option {
	return func(o *options) {
		if len(s) > 0 {
			o.Token = s
		}
	}
}

// WithSerialNo set the serialno for the vector robot.
func WithSerialNo(s string) Option {
	return func(o *options) {
		if len(s) > 0 {
			o.SerialNo = s
		}
	}
}

type RobotSDKInfoStore struct {
	GlobalGUID string `json:"global_guid"`
	Robots     []struct {
		Esn       string `json:"esn"`
		IPAddress string `json:"ip_address"`
		GUID      string `json:"guid"`
		Activated bool   `json:"activated"`
	} `json:"robots"`
}

func NewWpExternal(podURL, serial string) (*vector.Vector, error) {
	cfg := options{}

	var apiUrl = "http://" + podURL + "/api-sdk/get_sdk_info"

	httpClient := http.Client{
		Timeout: time.Second * 2,
	}

	req, err := http.NewRequest(http.MethodGet, apiUrl, nil)
	if err != nil {
		log.Println("Error getting data from wirepod API", err)
		return nil, err
	}
	res, getErr := httpClient.Do(req)
	if getErr != nil {
		log.Println("Error getting data from wirepod API", getErr)
		return nil, getErr
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		log.Println("unable to read data from the wirepod API")
		return nil, readErr
	}

	robotSDKInfo := RobotSDKInfoStore{}
	jsonErr := json.Unmarshal(body, &robotSDKInfo)

	if jsonErr != nil {
		log.Println("error reading JSON data")
		return nil, jsonErr
	}

	matched := false
	for _, robot := range robotSDKInfo.Robots {
		if strings.TrimSpace(strings.ToLower(robot.Esn)) == strings.TrimSpace(strings.ToLower(serial)) {
			cfg.Target = "127.0.0.1:443"
			cfg.Token = robot.GUID
			matched = true
			break
		}
	}
	if !matched {
		log.Println("vector-go-sdk error: serial did not match any bot in wirepod API")
		return nil, errors.New("vector-go-sdk error: serial did not match any bot in wirepod API")
	}

	c, err := client.New(
		client.WithTarget(cfg.Target),
		client.WithInsecureSkipVerify(),
	)
	if err != nil {
		return nil, err
	}
	if err := c.Connect(); err != nil {
		return nil, err
	}

	vect, _ := vector.New(
		vector.WithSerialNo(cfg.SerialNo),
		vector.WithTarget(cfg.Target),
		vector.WithToken(cfg.Token),
	)

	_, err = vect.Conn.BatteryState(
		context.Background(),
		&vectorpb.BatteryStateRequest{},
	)
	if err != nil {
		return nil, err
	}

	return vect, nil
}

func PlayCustomSound(file string) {
	go func() {
		pcmFile, _ := os.ReadFile(file)
		client, _ := victor.Conn.ExternalAudioStreamPlayback(
			context.Background(),
		)
		client.SendMsg(&vectorpb.ExternalAudioStreamRequest{
			AudioRequestType: &vectorpb.ExternalAudioStreamRequest_AudioStreamPrepare{
				AudioStreamPrepare: &vectorpb.ExternalAudioStreamPrepare{
					AudioFrameRate: 16000,
					AudioVolume:    70,
				},
			},
		})
		var audioChunks [][]byte
		for len(pcmFile) >= 1024 {
			audioChunks = append(audioChunks, pcmFile[:1024])
			pcmFile = pcmFile[1024:]
		}
		for _, chunk := range audioChunks {
			client.SendMsg(&vectorpb.ExternalAudioStreamRequest{
				AudioRequestType: &vectorpb.ExternalAudioStreamRequest_AudioStreamChunk{
					AudioStreamChunk: &vectorpb.ExternalAudioStreamChunk{
						AudioChunkSizeBytes: 1024,
						AudioChunkSamples:   chunk,
					},
				},
			})
			time.Sleep(time.Millisecond * 30)
		}
		fmt.Println("done sending audio")
		client.SendMsg(&vectorpb.ExternalAudioStreamRequest{
			AudioRequestType: &vectorpb.ExternalAudioStreamRequest_AudioStreamComplete{
				AudioStreamComplete: &vectorpb.ExternalAudioStreamComplete{},
			},
		})
	}()
}
