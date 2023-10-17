package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

const (
	QueueURL = "https://sqs.ap-northeast-1.amazonaws.com/299413808364/lee_test.fifo"
)

func main() {
	// AWSセッションを作成します
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("ap-northeast-1"), // 例: us-west-1
	})
	if err != nil {
		fmt.Println("Error creating session:", err)
		return
	}

	// SQSサービスのクライアントを作成します
	svc := sqs.New(sess)

	// キューからメッセージを受信するためのパラメータ
	receiveParams := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(QueueURL),
		MaxNumberOfMessages: aws.Int64(10), // 一度に受信する最大メッセージ数
		VisibilityTimeout:   aws.Int64(30), // メッセージが他の受信者から見えなくなる時間（秒）
		WaitTimeSeconds:     aws.Int64(20), // ポーリングの最大時間（ロングポーリング）
	}

	for {
		// メッセージを受信
		receiveResp, err := svc.ReceiveMessage(receiveParams)
		if err != nil {
			fmt.Println("Error receiving message:", err)
			return
		}

		// キューが空の場合、ループを続ける
		if len(receiveResp.Messages) == 0 {
			fmt.Println("Received no messages")
			continue
		}

		// 受信したメッセージを処理
		for _, message := range receiveResp.Messages {
			fmt.Printf("Message ID: %s\n", *message.MessageId)
			fmt.Printf("Message Body: %s\n", *message.Body)
			// 本番環境では、メッセージを処理した後、キューから削除するコードを追加する必要があります。
		}
	}
}
