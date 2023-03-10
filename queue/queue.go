package queue

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type QueueStatus string

func (q QueueStatus) String() string {
	return string(q)
}

const (
	Waiting QueueStatus = "waiting"
	Served  QueueStatus = "served"
)

type Queue struct {
	ID          string      `json:"id" gorm:"primaryKey type:varchar(255) not null"`
	Status      QueueStatus `json:"status" gorm:"type:varchar(255) not null"`
	QueueAt     string      `json:"queue_at" gorm:"type:varchar(255) not null"`
	QueueNumber int         `json:"queue_number" gorm:"type:int not null"`
	CreatedAt   time.Time   `json:"created_at" gorm:"type:datetime not null"`
}

type MySqlDataStore struct {
	db       *gorm.DB
	location *time.Location
}

func NewMySqlDataStore(db *gorm.DB) *MySqlDataStore {
	location := time.FixedZone("Asia/Jakarta", 7*60*60)
	time.Local = location

	return &MySqlDataStore{
		db:       db,
		location: location,
	}
}

func (m *MySqlDataStore) GetAvailableQueue() (*Queue, error) {
	date := time.Now().In(m.location).Format("2006-01-02")
	var queue Queue
	err := m.db.Transaction(func(tx *gorm.DB) error {
		tx.Set("gorm:query_option", "FOR UPDATE")

		// get the latest queue number
		result := tx.Order("created_at desc").Limit(1).Find(&queue)
		if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
			return result.Error
		}

		// if there is no queue number on db, then create a new queue number with value 1
		if result.RowsAffected == 0 {
			// if there is no queue number for today, then create a new queue number with value 1
			queueData, err := m.AddQueueNumber(tx, 1)
			if err != nil {
				return err
			}

			queue = *queueData
			return nil
		}

		// if there is a queue number on db, then check if the queue number is from today
		if queue.QueueAt != date {
			// if the queue number is not from today, then create a new queue number with value 1
			queueData, err := m.AddQueueNumber(tx, 1)
			if err != nil {
				return err
			}

			queue = *queueData
			return nil
		}

		// if the queue number is from today
		// then create a new queue with value queue number + 1
		queueData, err := m.AddQueueNumber(tx, queue.QueueNumber+1)
		if err != nil {
			return err
		}

		queue = *queueData
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &queue, nil
}

func (m *MySqlDataStore) AddQueueNumber(tx *gorm.DB, queueNumber int) (*Queue, error) {
	now := time.Now().In(m.location)

	queue := Queue{
		ID:          uuid.New().String(),
		Status:      Waiting,
		QueueAt:     now.Format("2006-01-02"),
		QueueNumber: queueNumber,
		CreatedAt:   now,
	}

	result := tx.Create(&queue)
	if result.Error != nil {
		return nil, result.Error
	}

	return &queue, nil
}

func (m *MySqlDataStore) ChangeQueueStatus(id string, status QueueStatus) error {
	result := m.db.
		Model(&Queue{}).
		Where("id = ?", id).
		Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (m *MySqlDataStore) GetAllQueueByCurrentDate() ([]Queue, error) {
	date := time.Now().In(m.location).Format("2006-01-02") // year-month-day

	var queues []Queue
	result := m.db.
		Where("queue_at = ?", date).
		Order("queue_number").
		Find(&queues)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, result.Error
	}

	return queues, nil
}
