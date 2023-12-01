package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/smithy-go"
)

const (
	MaxConcurrentGoroutines = 10 // 同時に実行されるgoroutineの最大数
)

var (
	sem               = semaphore.NewWeighted(MaxConcurrentGoroutines) // 同時に実行されるgoroutine数を管理(クローズするまでいつまでも実行可能)
	no_db_user_count  int64
	no_exist_db_user  []string
	no_exist_iam_user []string
	wg                sync.WaitGroup
	PostgresUser      = os.Getenv("POSTGRES_USER")
	PostgresPassword  = os.Getenv("POSTGRES_PASSWORD")
	PostgresDatabase  = os.Getenv("POSTGRES_DB")
	dsn               = "port=5432 sslmode=disable TimeZone=Asia/Tokyo host=postgres" + " user=" + PostgresUser + " password=" + PostgresPassword + " dbname=" + PostgresDatabase
)

type System struct {
	System string `gorm:"primaryKey;column:system"`
}

type Dbuserpassword struct {
	CombinationID int    `gorm:"primaryKey;column:combinationid"`
	System        string `gorm:"foreignKey:system;references:system;constraint:OnDelete:CASCADE;column:system"`
	Dbuser        string `gorm:"size:50;column:dbuser"`
	Password      string `gorm:"size:30;column:password"`
}

type Userprivilegestate struct {
	UserID         string
	CombinationID  int       `gorm:"foreignKey:combinationid;references:combinationid;constraint:OnDelete:CASCADE;column:combinationid"`
	StartTimeStamp time.Time `gorm:"type:timestamp with time zone;column:starttimestamp"`
	EndTimeStamp   time.Time `gorm:"type:timestamp with time zone;column:endtimestamp"`
	Index          int       `gorm:"primaryKey;autoIncrement;colum:index"`
}

// defaultではテーブル名は構造体名(末尾にsがつく)の複数形になるが、テーブル名を変更したい場合は、以下のようにTableName()を定義する
func (System) TableName() string {
	return "system"
}
func (Dbuserpassword) TableName() string {
	return "dbuserpassword"
}
func (Userprivilegestate) TableName() string {
	return "userprivilegestate"
}

func init() {
	// DB接続
	DB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("failed to connect database")
		panic(err)
	}

	// テーブルの作成
	err = DB.AutoMigrate(&System{})
	if err != nil {
		fmt.Println("failed to migrate System database")
		panic(err)
	}

	err = DB.AutoMigrate(&Dbuserpassword{})
	if err != nil {
		fmt.Println("failed to migrate Dbuserpassword database")
		panic(err)
	}

	err = DB.AutoMigrate(&Userprivilegestate{})
	if err != nil {
		fmt.Println("failed to migrate Userprivilegestate database")
		panic(err)
	}
}

///////////// ToDo /////////////
// AWS SDKを使って、IAMユーザの存在チェックを実装すること！
// goroutineを使って並列処理を実装すること！
// Ginを使ってGetリクエストを受け取ると、DBに登録されているユーザーの情報をjson形式で返すように実装！
// OpenTelemetry設定を追加すること！
///////////// ToDo /////////////

func main() {
	// DB接続
	DB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("failed to connect database")
		panic(err)
	}

	db_users := []string{"dbuser1", "dbuser2", "dbuser4", "dbuser5", "dbuser7"}
	wg.Add(2)
	go DbuserExistCheck(DB, db_users...)

	iam_users := []string{"jo-lee@kddi.com", "lee-testuser-for-iam", "iamuser3", "iamuser4", "iamuser5"}
	go AwsIamUserExistCheck(iam_users...)
	wg.Wait()

	fmt.Println("no_exist_user:", no_exist_db_user)
	fmt.Println("no_exist_iam_user:", no_exist_iam_user)
}

func DbuserExistCheck(db *gorm.DB, db_user ...string) {
	defer wg.Done()

	// セッションを利用するために、DBに接続する
	sqldb, err := db.DB()
	if err != nil {
		fmt.Println("failed to connect DB")
		panic(err)
	}
	// DBに対する処理が終わったら、DB接続を解除する
	defer sqldb.Close()

	for _, v := range db_user {
		db.Where("dbuser = ?", v).Find(&Dbuserpassword{}).Count(&no_db_user_count)
		if no_db_user_count == 0 {
			no_exist_db_user = append(no_exist_db_user, v)
			fmt.Printf("%v is not exist\n", v)
		} else {
			continue
		}
	}
}

func AwsIamUserExistCheck(iam_user ...string) {
	defer wg.Done()

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-northeast-1"))

	if err != nil {
		fmt.Println("failed to load config")
		panic(err)
	}

	svc := iam.NewFromConfig(cfg)

	// IAMユーザの存在チェック
	for _, v := range iam_user {
		input := &iam.GetUserInput{
			UserName: aws.String(v),
		}

		_, err := svc.GetUser(context.TODO(), input)
		if err != nil {
			var nsk *types.NoSuchEntityException
			if errors.As(err, &nsk) {
				fmt.Println("NoSuchEntityException")
				no_exist_iam_user = append(no_exist_iam_user, v)
			}
			var apiErr smithy.APIError
			if errors.As(err, &apiErr) {
				fmt.Println("StatusCode:", apiErr.ErrorCode(), ", Msg:", apiErr.ErrorMessage())
			}
		}
	}
}
