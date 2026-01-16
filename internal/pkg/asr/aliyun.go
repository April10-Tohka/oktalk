package asr

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"oktalk/internal/pkg/config"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type AliyunASR struct {
	conf *config.AliyunConfig
}

var dialer = websocket.DefaultDialer
var wsURL string
var asrModel string
var apiKey string

func NewAliyunASR(conf *config.AliyunConfig) *AliyunASR {
	wsURL = conf.ASR.WsURL
	asrModel = conf.ASR.Model
	apiKey = conf.DASHSCOPE_API_KEY
	return &AliyunASR{conf: conf}
}

// RecognizeOnce 对应你示例中的逻辑，但进行了工程化封装
func (a *AliyunASR) RecognizeOnce(ctx context.Context, audioPath string) (string, error) {
	// 连接websocket服务
	conn, err := connectWebSocket(apiKey)
	if err != nil {
		return "", fmt.Errorf("连接 WebSocket 失败: %w", err)
	}
	defer closeConnection(conn)

	// 发送run-task指令
	taskID, err := sendRunTaskCmd(conn)
	if err != nil {
		return "", fmt.Errorf("发送 run-task 失败: %w", err)
	}

	// 3. 启动结果接收器
	resultChan := make(chan string, 1)
	errorChan := make(chan error, 1)
	taskStarted := make(chan bool, 1)

	// 启动一个goroutine来接受websocket结果
	go receiveResults(ctx, conn, resultChan, errorChan, taskStarted)

	// 4. 等待 task-started
	select {
	case <-ctx.Done():
		logrus.WithContext(ctx).Warnf("")
	case <-taskStarted:
		// 任务启动成功
		logrus.WithContext(ctx).Info("✅ 任务启动成功")
	case err := <-errorChan:
		return "", err
	case <-time.After(10 * time.Second):
		return "", fmt.Errorf("等待 task-started 超时")
	}

	// 5.发送音频数据
	err = sendAudioData(ctx, conn, audioPath)
	if err != nil {
		return "", fmt.Errorf("发送音频数据失败: %w", err)
	}

	// 6.发送完音频后，发送 finish-task 指令
	err = sendFinishTaskCmd(ctx, conn, taskID)
	if err != nil {
		return "", fmt.Errorf("发送 finish-task 失败: %w", err)
	}
	// 7. 等待识别结果
	select {
	case <-ctx.Done():
		logrus.WithContext(ctx).Warningf("等待识别结果 %v", ctx.Err())
		return "", nil
	case text := <-resultChan:
		logrus.WithContext(ctx).Infof("识别到结果: %s", text)
		return text, nil
	case err = <-errorChan:
		logrus.WithContext(ctx).Errorf("等待识别结果遇到错误：%v", err)
		return "", nil
	case <-time.After(30 * time.Second):
		logrus.WithContext(ctx).Warningf("等待识别结果超时")
		return "", nil
	}
}

// 定义结构体来表示JSON数据
type Header struct {
	Action       string                 `json:"action"`
	TaskID       string                 `json:"task_id"`
	Streaming    string                 `json:"streaming"`
	Event        string                 `json:"event"`
	ErrorCode    string                 `json:"error_code,omitempty"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	Attributes   map[string]interface{} `json:"attributes"`
}

type Output struct {
	Sentence struct {
		BeginTime int64  `json:"begin_time"`
		EndTime   *int64 `json:"end_time"`
		Text      string `json:"text"`
		Words     []struct {
			BeginTime   int64  `json:"begin_time"`
			EndTime     *int64 `json:"end_time"`
			Text        string `json:"text"`
			Punctuation string `json:"punctuation"`
		} `json:"words"`
	} `json:"sentence"`
}

type Payload struct {
	TaskGroup  string `json:"task_group"`
	Task       string `json:"task"`
	Function   string `json:"function"`
	Model      string `json:"model"`
	Parameters Params `json:"parameters"`
	Input      Input  `json:"input"`
	Output     Output `json:"output,omitempty"`
	Usage      *struct {
		Duration int `json:"duration"`
	} `json:"usage,omitempty"`
}

type Params struct {
	Format                   string `json:"format"`
	SampleRate               int    `json:"sample_rate"`
	VocabularyID             string `json:"vocabulary_id"`
	DisfluencyRemovalEnabled bool   `json:"disfluency_removal_enabled"`
}

type Input struct {
}

type Event struct {
	Header  Header  `json:"header"`
	Payload Payload `json:"payload"`
}

// 连接WebSocket服务
func connectWebSocket(apiKey string) (*websocket.Conn, error) {
	header := make(http.Header)
	header.Add("Authorization", fmt.Sprintf("bearer %s", apiKey))
	conn, _, err := dialer.Dial(wsURL, header)
	return conn, err
}

// receiveResults 接收 WebSocket 结果
func receiveResults(ctx context.Context, conn *websocket.Conn, resultChan chan<- string, errorChan chan<- error, taskStarted chan<- bool) {
	var finalText string
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			logrus.WithContext(ctx).Errorf("解析服务器消息失败：%v", err)
			errorChan <- err
			return
		}
		var event Event
		err = json.Unmarshal(message, &event)
		if err != nil {
			logrus.WithContext(ctx).Warnf("解析事件失败：%v", err)
			continue
		}
		switch event.Header.Event {
		case "task-started":
			taskStarted <- true

		case "result-generated":
			// 从 output.sentence.text 提取识别结果
			text := event.Payload.Output.Sentence.Text
			if text != "" {
				finalText = text
				logrus.WithContext(ctx).Info("✅ 识别结果：%s", finalText)
			}

		case "task-finished":
			resultChan <- finalText
			return

		case "task-failed":
			errorMsg := event.Header.ErrorMessage
			if errorMsg == "" {
				errorMsg = "ASR 任务失败"
			}
			errorChan <- fmt.Errorf(errorMsg)
			return
		}
	}

}

// 发送run-task指令
func sendRunTaskCmd(conn *websocket.Conn) (string, error) {
	runTaskCmd, taskID, err := generateRunTaskCmd()
	if err != nil {
		return "", err
	}
	err = conn.WriteMessage(websocket.TextMessage, []byte(runTaskCmd))
	return taskID, err
}

// 生成run-task指令
func generateRunTaskCmd() (string, string, error) {
	taskID := uuid.New().String()
	runTaskCmd := Event{
		Header: Header{
			Action:    "run-task",
			TaskID:    taskID,
			Streaming: "duplex",
		},
		Payload: Payload{
			TaskGroup: "audio",
			Task:      "asr",
			Function:  "recognition",
			Model:     asrModel,
			Parameters: Params{
				Format:     "wav",
				SampleRate: 16000,
			},
			Input: Input{},
		},
	}
	runTaskCmdJSON, err := json.Marshal(runTaskCmd)
	return string(runTaskCmdJSON), taskID, err
}

// 发送音频数据
func sendAudioData(ctx context.Context, conn *websocket.Conn, audioPath string) error {
	select {
	case <-ctx.Done():
		logrus.WithContext(ctx).Warningf("")
		return ctx.Err()
	default:
		logrus.WithContext(ctx).Infof("发送音频数据")
	}

	file, err := os.Open(audioPath)
	if err != nil {
		return err
	}
	defer file.Close()

	buf := make([]byte, 1024)
	for {
		n, err := file.Read(buf)
		if n == 0 {
			break
		}
		if err != nil && err != io.EOF {
			return err
		}
		err = conn.WriteMessage(websocket.BinaryMessage, buf[:n])
		if err != nil {
			return err
		}
		time.Sleep(50 * time.Millisecond)
	}
	return nil
}

// 发送finish-task指令
func sendFinishTaskCmd(ctx context.Context, conn *websocket.Conn, taskID string) error {
	finishTaskCmd, err := generateFinishTaskCmd(taskID)
	if err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:

	}
	err = conn.WriteMessage(websocket.TextMessage, []byte(finishTaskCmd))
	return err
}

// 生成finish-task指令
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
	finishTaskCmdJSON, err := json.Marshal(finishTaskCmd)
	return string(finishTaskCmdJSON), err
}

// 关闭连接
func closeConnection(conn *websocket.Conn) {
	if conn != nil {
		conn.Close()
	}
}
