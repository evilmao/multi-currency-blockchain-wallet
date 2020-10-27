package models

import (
	"github.com/jinzhu/gorm"
)

// Broadcast task status.
const (
	BroadcastTaskStatusNotRecord = 0
	BroadcastTaskStatusRecord    = 1

	BroadcastTaskStatusDone = 50
)

// Separator of tx signatures and pubKey for strings.Join.
const (
	TaskTxSigPubKeySep = "'"
)

// BroadcastTask represents broadcast task.
type BroadcastTask struct {
	gorm.Model
	TxSequenceID string `gorm:"size:32;index"`
	TxSignatures string `gorm:"type:mediumtext"`
	TxPubKeys    string `gorm:"type:text"`
	TaskStatus   int8   `gorm:"type:tinyint;index"`
}

func (*BroadcastTask) TableName() string { return "broadcast_task" }

// FirstOrCreate find first matched record or create a new one.
func (task *BroadcastTask) FirstOrCreate(db *gorm.DB) error {
	if task.TaskStatus == BroadcastTaskStatusNotRecord {
		task.TaskStatus = BroadcastTaskStatusRecord
	}
	return db.FirstOrCreate(task, "tx_sequence_id = ?", task.TxSequenceID).Error
}

// Done updates the task status to done.
func (task *BroadcastTask) Done(db *gorm.DB) error {
	return db.Model(task).Updates(M{
		"task_status": BroadcastTaskStatusDone,
	}).Error
}

// IsBroadcastTaskExist returns whether broadcast task is exist by tx sequence id.
func IsBroadcastTaskExist(dbInst *gorm.DB, txSequenceID string) (bool, error) {
	var task BroadcastTask
	err := dbInst.Where("tx_sequence_id = ?", txSequenceID).First(&task).Error
	return task.TxSequenceID == txSequenceID, err
}

// GetBroadcastTasksByStatus returns broadcast tasks in the status.
func GetBroadcastTasksByStatus(dbInst *gorm.DB, status uint) []*BroadcastTask {
	var tasks []*BroadcastTask
	dbInst.Where("task_status = ?", status).Find(&tasks)
	return tasks
}
