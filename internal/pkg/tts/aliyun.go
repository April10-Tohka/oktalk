package tts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"oktalk/internal/pkg/config"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type AliyunTTS struct {
	conf *config.AliyunConfig
}

var dialer = websocket.DefaultDialer
var wsURL string
var ttsModel string
var apiKey string

// --- 协议结构体定义 ---

type Header struct {
	Action       string `json:"action"`
	TaskID       string `json:"task_id"`
	Streaming    string `json:"streaming"`
	Event        string `json:"event"`
	ErrorCode    string `json:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

type Payload struct {
	TaskGroup  string `json:"task_group"`
	Task       string `json:"task"`
	Function   string `json:"function"`
	Model      string `json:"model"`
	Parameters Params `json:"parameters"`
	Input      Input  `json:"input"`
}

type Params struct {
	TextType   string `json:"text_type"`
	Voice      string `json:"voice"`
	Format     string `json:"format"`
	SampleRate int    `json:"sample_rate"`
	Volume     int    `json:"volume"`
	Rate       int    `json:"rate"`
	Pitch      int    `json:"pitch"`
	EnableSsml bool   `json:"enable_ssml"`
}

type Input struct {
	Text string `json:"text,omitempty"`
}

type Event struct {
	Header  Header  `json:"header"`
	Payload Payload `json:"payload"`
}

func NewAliyunTTS(conf *config.AliyunConfig) *AliyunTTS {
	wsURL = conf.TTS.WsURL
	ttsModel = conf.TTS.Model
	apiKey = conf.DASHSCOPE_API_KEY
	return &AliyunTTS{conf: conf}
}

// Synthesize 语音合成
func (p *AliyunTTS) Synthesize(ctx context.Context, text string) ([]byte, error) {
	// 1. 建立 WebSocket 连接
	conn, err := connectWebSocket()
	if err != nil {
		return nil, fmt.Errorf("连接 WebSocket 失败: %w", err)
	}
	defer closeConnection(conn)

	// 2. 发送 run-task 指令
	taskID, err := sendRunTaskCmd(ctx, conn)
	if err != nil {
		return nil, fmt.Errorf("发送 run-task 失败: %w", err)
	}

	// 3. 启动音频接收器
	audioChan := make(chan []byte, 1)
	errorChan := make(chan error, 1)
	taskStarted := make(chan bool, 1)

	// 启动一个goroutine来接收websocket结果
	go receiveResults(ctx, conn, audioChan, errorChan, taskStarted)

	// 4. 等待 task-started
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-taskStarted:
		// 任务启动成功
		logrus.WithContext(ctx).Infof("tts任务启动成功")
	case err := <-errorChan:
		return nil, err
	case <-time.After(10 * time.Second):
		return nil, fmt.Errorf("等待 task-started 超时")
	}

	// 5. 发送文本内容
	if err := sendText(ctx, conn, taskID, text); err != nil {
		logrus.WithContext(ctx).Errorf("发送文本失败 %v", err)
		return nil, fmt.Errorf("发送文本失败: %w", err)
	}

	// 6. 发送 finish-task指令
	if err := sendFinishTaskCmd(ctx, conn, taskID); err != nil {
		logrus.WithContext(ctx).Errorf("发送finish-task 失败 %v", err)
		return nil, fmt.Errorf("发送 finish-task 失败: %w", err)
	}

	// 7. 等待音频数据
	select {
	case audioData := <-audioChan:
		return audioData, nil
	case err := <-errorChan:
		return nil, err
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("合成超时")
	}
}

// connectWebSocket 建立 WebSocket 连接
func connectWebSocket() (*websocket.Conn, error) {
	header := make(http.Header)
	header.Add("Authorization", fmt.Sprintf("bearer %s", apiKey))
	header.Add("X-DashScope-DataInspection", "enable")

	conn, _, err := dialer.Dial(wsURL, header)
	return conn, err
}

// sendRunTask 发送 run-task 指令
func sendRunTaskCmd(ctx context.Context, conn *websocket.Conn) (string, error) {
	runTaskCmd, taskID, err := generateRunTaskCmd()
	if err != nil {
		logrus.WithContext(ctx).Warningf("生成tts run-task指令失败 %v", err)
	}
	err = conn.WriteMessage(websocket.TextMessage, []byte(runTaskCmd))
	return taskID, err
}
func generateRunTaskCmd() (string, string, error) {
	// 生成任务ID
	taskID := uuid.New().String()
	// 生成 run-task指令
	runTaskCmd := Event{
		Header: Header{
			Action:    "run-task",
			TaskID:    taskID,
			Streaming: "duplex",
		},
		Payload: Payload{
			TaskGroup: "audio",
			Task:      "tts",
			Function:  "SpeechSynthesizer",
			Model:     ttsModel,
			Parameters: Params{
				TextType:   "PlainText",
				Voice:      "longanyang",
				Format:     "mp3",
				SampleRate: 22050,
				Volume:     50,
				Rate:       1,
				Pitch:      1,
				// 如果enable_ssml设为true，只允许发送一次continue-task指令，否则会报错“Text request limit violated, expected 1.”
				EnableSsml: false,
			},
			Input: Input{},
		},
	}

	runTaskJSON, err := json.Marshal(runTaskCmd)
	return string(runTaskJSON), taskID, err
}

// receiveResults 接收 WebSocket 结果
func receiveResults(ctx context.Context, conn *websocket.Conn, audioChan chan<- []byte, errorChan chan<- error, taskStart chan<- bool) {
	var audioBuffer bytes.Buffer

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			logrus.WithContext(ctx).Errorf("读取消息失败: %v", err)
			errorChan <- fmt.Errorf("读取消息失败: %w", err)
			return
		}

		// 处理二进制消息(音频数据)
		if messageType == websocket.BinaryMessage {
			audioBuffer.Write(message)
			continue
		}

		// 处理文本消息(事件)
		var event Event
		if err := json.Unmarshal(message, &event); err != nil {
			continue
		}

		switch event.Header.Event {
		case "task-started":
			taskStart <- true

		case "task-finished":
			audioChan <- audioBuffer.Bytes()
			return

		case "task-failed":
			errorMsg := event.Header.ErrorMessage
			if errorMsg == "" {
				errorMsg = "TTS 任务失败"
			}
			errorChan <- fmt.Errorf(errorMsg)
			return
		}
	}
}

// sendText 发送待合成文本
func sendText(ctx context.Context, conn *websocket.Conn, taskID string, text string) error {
	continueTaskCmd, err := generateContinueTaskCmd(taskID, text)
	if err != nil {
		logrus.WithContext(ctx).Errorf("生成continue-task指令失败 %v", err)
		return err
	}
	err = conn.WriteMessage(websocket.TextMessage, []byte(continueTaskCmd))
	return err
}

// generateContinueTaskCmd 生成continue-task指令
func generateContinueTaskCmd(taskID string, text string) (string, error) {
	continueTaskCmd := Event{
		Header: Header{
			Action:    "continue-task",
			TaskID:    taskID,
			Streaming: "duplex",
		},
		Payload: Payload{
			Input: Input{
				Text: text,
			},
		},
	}

	continueTaskJson, err := json.Marshal(continueTaskCmd)
	return string(continueTaskJson), err
}

// generateFinishTaskCmd 生成finish-task指令
func generateFinishTaskCmd(taskID string) (string, error) {
	finishTaskCmd := Event{
		Header: Header{
			Action:    "finish-task",
			TaskID:    taskID,
			Streaming: "duplex",
		},
		Payload: Payload{
			Input: Input{},
		},
	}
	finishTaskJson, err := json.Marshal(finishTaskCmd)
	return string(finishTaskJson), err
}

// sendFinishTask 发送 finish-task 指令
func sendFinishTaskCmd(ctx context.Context, conn *websocket.Conn, taskID string) error {
	// 稍微延迟,确保文本已发送
	time.Sleep(200 * time.Millisecond)
	finishTaskCmd, err := generateFinishTaskCmd(taskID)
	if err != nil {
		logrus.WithContext(ctx).Errorf("生成finish-task指令失败 %v", err)
		return err
	}
	return conn.WriteMessage(websocket.TextMessage, []byte(finishTaskCmd))
}

// 关闭连接
func closeConnection(conn *websocket.Conn) {
	if conn != nil {
		conn.Close()
	}
}
