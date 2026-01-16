package model

import (
	"time"

	"gorm.io/gorm"
)

// UserLearningRecord 对应 PRD 6.2 节：学习记录表
type UserLearningRecord struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	UserID         uint           `gorm:"index" json:"user_id"`
	Date           time.Time      `gorm:"type:date;index" json:"date"`
	SpeakingScore  float64        `gorm:"type:decimal(5,2)" json:"speaking_score"` // 综合口语分
	FluencyScore   float64        `gorm:"type:decimal(5,2)" json:"fluency_score"`  // 流利度
	AccuracyScore  float64        `gorm:"type:decimal(5,2)" json:"accuracy_score"` // 准确度
	Duration       int            `json:"duration"`                                // 学习时长(秒)
	CompletedTasks int            `json:"completed_tasks"`                         // 完成任务数
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (UserLearningRecord) TableName() string {
	return "user_learning_record"
}
