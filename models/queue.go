package models

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// MessageStatus The current status of a message
type MessageStatus string

const (
	created   MessageStatus = "created"
	inTransit MessageStatus = "in_transit"
	queued    MessageStatus = "queued"
	reQueued  MessageStatus = "requeued"
	processed MessageStatus = "processed"
)

// Queue the struct to hold queue information
type Queue struct {
	ID        int
	Name      string
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

// CreateQueue adds a queue to the database
func (q Queue) CreateQueue() (Queue, error) {
	if q.Name == "" {
		err := errors.New("Name must be specified")
		return Queue{}, err
	}
	connString := fmt.Sprintf(
		"%s:%s@/%s?charset=utf8&parseTime=True&loc=Local",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)
	db, err := gorm.Open(os.Getenv("DIALECT"), connString)
	defer db.Close()
	if err != nil {
		return Queue{}, err
	}
	// check if queue already exist
	exists := Queue{}
	err = db.Where("name = ?", q.Name).First(&exists).Error
	if err != nil {
		return Queue{}, err
	}
	if (Queue{}) != exists {
		return Queue{}, errors.New("Queue already exist")
	}
	if err := db.Create(&q).Error; err != nil {
		return Queue{}, err
	}
	return q, nil
}

// GetQueueByID returns queue detail of the given id
func (q *Queue) GetQueueByID(id int) error {
	db, err := gorm.Open("mysql", "root:root@/message_queue?charset=utf8&parseTime=True&loc=Local")
	defer db.Close()
	if err != nil {
		return err
	}

	if err := db.First(&q, id).Error; err != nil {
		return err
	}

	return nil
}

// GetQueueByName returns queue details of the given queue_name
func (q *Queue) GetQueueByName(queueName string) error {
	db, err := gorm.Open("mysql", "root:root@/message_queue?charset=utf8&parseTime=True&loc=Local")
	defer db.Close()
	if err != nil {
		return err
	}
	if err := db.Where("name = ? ", queueName).First(&q).Error; err != nil {
		return err
	}

	return nil
}

// GetMessages returns the messages attached to this queue
func (q Queue) GetMessages() ([]Message, error) {
	db, err := gorm.Open("mysql", "root:root@/message_queue?charset=utf8&parseTime=True&loc=Local")
	defer db.Close()
	if err != nil {
		return []Message{}, err
	}

	var messages []Message
	if err := db.Model(&q).Related(&messages).Error; err != nil {
		return []Message{}, err
	}

	return messages, nil
}

// GetMessage returns the oldest message inside the queue
func (q Queue) GetMessage() (Message, error) {
	db, err := gorm.Open("mysql", "root:root@/message_queue?charset=utf8&parseTime=True&loc=Local")
	defer db.Close()
	if err != nil {
		return Message{}, err
	}

	var message Message
	// fmt.Println(time.Now())
	err = db.Where("queue_id = ? AND (status = ? OR status = ?) AND available_at <= ?", q.ID, created, reQueued, time.Now()).
		Order("id asc").
		Limit(1).
		Find(&message).Error
	if err != nil {
		return Message{}, err
	}

	// change the message status
	if err := db.Model(&message).Update("status", inTransit).Error; err != nil {
		return Message{}, err
	}
	return message, nil
}
