package queue_test

import (
	"ngantri/queue"
	"ngantri/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAvailableQueue(t *testing.T) {
	db, _ := utils.ConnectDB()
	store := queue.NewMySqlDataStore(db)

	//	clean up
	db.Exec("DELETE FROM queues")

	// Case 1: if there is no queue number for today, then create a new queue number with value 1
	for i := 1; i <= 100; i++ {
		queueData, err := store.GetAvailableQueue()
		assert.Nil(t, err)
		assert.NotNil(t, queueData)
		assert.Equal(t, i, queueData.QueueNumber)
	}

	// Case 2: if there is a queue number for today, then create a new queue number with value queueNumber + 1
	for i := 101; i <= 200; i++ {
		queueData, err := store.GetAvailableQueue()
		assert.Nil(t, err)
		assert.NotNil(t, queueData)
		assert.Equal(t, i, queueData.QueueNumber)
	}

	//	clean up
	db.Exec("DELETE FROM queues")
}
