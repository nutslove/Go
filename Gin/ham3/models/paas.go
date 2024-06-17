package models

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Projects struct {
	ProjectId   string `gorm:"primaryKey;column:project_id"`
	ProjectName string `gorm:"not null;column:project_name"`
}

type LOGaaS struct {
	gorm.Model
	ProjectId   string `gorm:"not null;index;foreignKey:ProjectId;references:Projects.ProjectId;constraint:OnDelete:SET NULL;column:project_id"`
	ClusterName string `gorm:"not null;column:cluster_name;unique"`
	ClusterType string `gorm:"not null;column:cluster_type"`
	GuiEndpoint string `gorm:"not null;column:gui_endpoint"`
	ApiEndpoint string `gorm:"not null;column:api_endpoint"`
	Status      string `gorm:"not null;column:status"`
}

type CaaS struct {
	gorm.Model
	ProjectId string `gorm:"not null;index;foreignKey:ProjectId;references:Projects.ProjectId;constraint:OnDelete:SET NULL;column:project_id"`
	Namespace string `grom:"not null;column:namespace"`
	Status    string `gorm:"not null;column:status"`
}

type AAPaaS struct {
	gorm.Model
	ProjectId string `gorm:"not null;index;foreignKey:ProjectId;references:Projects.ProjectId;constraint:OnDelete:SET NULL;column:project_id"`
	Status    string `gorm:"not null;column:status"`
}

func ConnectDb() *gorm.DB {
	// DB接続
	db, err := gorm.Open(sqlite.Open("ham3.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// テーブル作成/更新
	db.AutoMigrate(&Projects{})
	db.AutoMigrate(&LOGaaS{})
	db.AutoMigrate(&CaaS{})
	db.AutoMigrate(&AAPaaS{})

	return db
}

func (Projects) TableName() string {
	return "projects"
}

func (LOGaaS) TableName() string {
	return "logaas"
}

func (CaaS) TableName() string {
	return "caas"
}

func (AAPaaS) TableName() string {
	return "aapaas"
}
