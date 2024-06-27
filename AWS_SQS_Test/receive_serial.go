// 受信を並列にすると同じメッセージが複数回処理されたので、並列化をやめてシングルスレッドで処理するようにした
package main

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

const (
	QueueURL = "https://sqs.ap-northeast-1.amazonaws.com/<AWSアカウントID>/xx.fifo"
)

func main() {
	// AWSセッションを作成
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("ap-northeast-1"),
	})
	if err != nil {
		fmt.Println("Error creating session:", err)
		return
	}

	// SQSサービスのクライアントを作成
	svc := sqs.New(sess)

	// キューからメッセージを受信するためのパラメータ
	receiveParams := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(QueueURL),
		MaxNumberOfMessages: aws.Int64(10), // 一度に受信する最大メッセージ数
		VisibilityTimeout:   aws.Int64(30), // メッセージが他の受信者から見えなくなる時間（秒）
		WaitTimeSeconds:     aws.Int64(20), // ロングポーリングを有効にする時間（秒）
	}

	for {
		// メッセージを受信
		resp, err := svc.ReceiveMessage(receiveParams)
		time.Sleep(1 * time.Second) // 1秒ごとに最大10件のメッセージを受信(処理)するためのsleep
		if err != nil {
			fmt.Println("Error receiving message:", err)
			return
		}

		// キューが空の場合、"Received no messages"を出力してループの最初に戻る
		if len(resp.Messages) == 0 {
			fmt.Println("Received no messages")
			continue
		}

		// 受信したメッセージを処理
		for _, message := range resp.Messages {
			fmt.Printf("Message Body:  %s\n", *message.Body)
			_, err := svc.DeleteMessage(&sqs.DeleteMessageInput{
				QueueUrl:      aws.String(QueueURL),
				ReceiptHandle: message.ReceiptHandle,
			})
			if err != nil {
				fmt.Println("Failed to delete message:", err)
			}
		}

		fmt.Println("----------------------------------------")
	}
}
