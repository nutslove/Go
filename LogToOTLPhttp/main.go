package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

// LogBuffer はログをバッファリングし、定期的にフラッシュするための構造体
type LogBuffer struct {
	buffer    *strings.Builder
	mutex     sync.Mutex
	batchSize int
	ticker    *time.Ticker
	sid       string
	label     map[string]string
	backend   LogBackend
	byteCount int // 現在のバイト数を追跡
}

// LogBackend はログの保存先を抽象化するインターフェース
type LogBackend interface {
	Flush(logs string) error
}

// MockBackend はテスト用のバックエンド実装
type MockBackend struct{}

func (b *MockBackend) Flush(logs string) error {
	fmt.Printf("Flushing logs to backend: %d bytes\n", len(logs))
	// 実際の実装ではDBや外部APIへの書き込み処理を行う
	return nil
}

// NewLogBuffer はバッファの新しいインスタンスを作成
func NewLogBuffer(batchSize int, flushInterval time.Duration, sid string, label map[string]string, backend LogBackend) *LogBuffer {
	buffer := &LogBuffer{
		buffer:    &strings.Builder{},
		batchSize: batchSize,
		sid:       sid,
		label:     label,
		backend:   backend,
		byteCount: 0,
	}

	// 定期的なフラッシュのためのティッカーを開始
	buffer.ticker = time.NewTicker(flushInterval)
	go buffer.flushRoutine()

	return buffer
}

// Write はログをバッファに追加し、必要に応じてフラッシュ
func (lb *LogBuffer) Write(logEntry string) error {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	// 改行を含むログエントリのサイズを計算
	entrySize := len(logEntry) + 1 // +1 for newline

	// ログエントリをバッファに追加
	lb.buffer.WriteString(logEntry)
	lb.buffer.WriteByte('\n')
	lb.byteCount += entrySize

	// バッファサイズがバッチサイズを超えたらフラッシュ
	if lb.byteCount >= lb.batchSize {
		return lb.flush()
	}
	return nil
}

// フラッシュループの処理
func (lb *LogBuffer) flushRoutine() {
	for range lb.ticker.C {
		lb.mutex.Lock()
		if lb.byteCount > 0 {
			if err := lb.flush(); err != nil {
				log.Printf("Error flushing logs: %v", err)
			}
		}
		lb.mutex.Unlock()
	}
}

// Stop はバッファの処理を停止し、残りのログをフラッシュ
func (lb *LogBuffer) Stop() error {
	lb.ticker.Stop()
	lb.mutex.Lock()
	defer lb.mutex.Unlock()
	return lb.flush()
}

// flush はバッファの内容をバックエンドに送信
func (lb *LogBuffer) flush() error {
	if lb.byteCount == 0 {
		return nil
	}

	// バッファの内容を文字列として取得
	logs := lb.buffer.String()

	// バッファをリセット
	lb.buffer.Reset()
	lb.byteCount = 0

	// バックエンドにログを送信
	return lb.backend.Flush(logs)
}

func main() {
	///// TODO: SQS or SNSからのS3オブジェクト情報から、SID・ログ種別を確認して、SID・ログ種別ごとにlogBufferインスタンスを作成する

	filePath := "/mnt/c/Users/nuts_/Github/Go/LogToOTLPhttp/log.txt"

	// // アプリケーション終了時にバッファを停止
	// defer logBuffer.Stop()

	for {
		_, err := os.Stat(filePath)
		if os.IsNotExist(err) {
			log.Printf("ファイルは存在しません: %s\n", filePath)
			time.Sleep(300 * time.Second)
			continue
		} else if err != nil {
			log.Printf("エラー: %v\n", err)
			time.Sleep(300 * time.Second)
			continue
		}

		label := map[string]string{
			"LogType": "alb",
		}
		// 例: 5MB以上または60秒ごとにフラッシュするバッファの作成
		logBuffer := NewLogBuffer(1024*1024*5, 60*time.Second, "test", label, &MockBackend{})

		log.Printf("ファイルが存在します: %s - 内容を読み込みます\n", filePath)
		// ファイルを開く
		file, err := os.Open(filePath)
		if err != nil {
			log.Printf("ファイルを開けませんでした: %v\n", err)
			time.Sleep(300 * time.Second)
			continue
		}

		// ファイルを行ごとに読み込む
		scanner := bufio.NewScanner(file)
		lineCount := 0

		log.Println("ファイルの内容:")
		for scanner.Scan() {
			lineCount++
			line := scanner.Text()
			if err := logBuffer.Write(line); err != nil {
				log.Printf("Error writing log: %v", err)
			}
			log.Printf("行 %d: %s\n", lineCount, line)
		}

		if err := scanner.Err(); err != nil {
			log.Printf("ファイル読み込み中にエラーが発生しました: %v\n", err)
		}

		log.Printf("ファイル読み込み完了。合計 %d 行\n", lineCount)

		file.Close()

		// 次の確認まで待機
		time.Sleep(300 * time.Second)
	}
}
