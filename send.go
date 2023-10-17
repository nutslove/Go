package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

const (
	QueueURL = "https://sqs.ap-northeast-1.amazonaws.com/299413808364/lee_test.fifo"
)

func main() {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("ap-northeast-1"),
	}))

	svc := sqs.New(sess)

	messageBody := "Hello, SQS from Go!"

	// 乱数生成器を初期化します。これは一度だけ実行する必要があります。
	rand.Seed(time.Now().UnixNano())

	MessageDedupNum := rand.Int() // 重複排除IDは、同一のメッセージを複数回送信しないようにするためのもの。
	MessageDedupId := strconv.Itoa(MessageDedupNum)

	// メッセージの送信
	sendMsgInput := &sqs.SendMessageInput{
		MessageBody: aws.String(messageBody),
		QueueUrl:    aws.String(QueueURL),
		// FIFOキューを使用している場合、MessageGroupIdが必要
		MessageGroupId: aws.String("SQS_TEST"), // 通常、同じ処理を行うメッセージに共通の値を設定
		// MessageDeduplicationIdはオプション
		MessageDeduplicationId: aws.String(MessageDedupId), // MessageDeduplicationId は、可能な限りユニークな値を提供することが重要です（ただし、必須ではありません）。これは、同一の MessageDeduplicationId を持つメッセージが重複排除期間内に複数回送信された場合、後続のメッセージが受け入れられないことを意味します。
	}

	_, err := svc.SendMessage(sendMsgInput)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Message sent:", messageBody)
}
