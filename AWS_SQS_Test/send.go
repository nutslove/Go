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

// func sendmsg(i int, c chan int, svc *sqs.SQS) {
func sendmsg(i int, svc *sqs.SQS) {

	// 乱数生成器を初期化。これは一度だけ実行する必要がある。
	rand.Seed(time.Now().UnixNano())

	n := rand.Intn(1000000000000)

	MessageDedupId := strconv.Itoa(n)
	messageBody := "Hello, SQS from Go! " + strconv.Itoa(i)

	// メッセージの送信
	sendMsgInput := &sqs.SendMessageInput{
		MessageBody: aws.String(messageBody),
		QueueUrl:    aws.String(QueueURL),
		// FIFOキューを使用している場合、MessageGroupIdが必要
		MessageGroupId: aws.String("SQS_TEST"), // 通常、同じ処理を行うメッセージに共通の値を設定
		// MessageDeduplicationIdはオプション
		MessageDeduplicationId: aws.String(MessageDedupId),
		// MessageDeduplicationId は、可能な限りユニークな値を提供することが重要（ただし、必須ではない）。
		// これは、同一の MessageDeduplicationId を持つメッセージが重複排除期間内に複数回送信された場合、後続のメッセージが受け入れられないことを意味する。(MessageDeduplicationIdが重複するメッセージはreceive側で受信しない)
	}
	_, err := svc.SendMessage(sendMsgInput)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	// c <- i
}

func main() {
	// c := make(chan int)

	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("ap-northeast-1"),
	}))

	svc := sqs.New(sess)

	// for i := 0; i < 10; i++ {
	// 	go sendmsg(i, c, svc)
	// }
	// defer close(c)

	for i := 0; i < 1000; i++ {
		sendmsg(i, svc)
		fmt.Println("Message sent: ", i)
	}

	// for range [10]int{} {
	// 	select {
	// 	case index := <-c:
	// 		fmt.Println("Message sent: ", index)
	// 	}
	// }
}
