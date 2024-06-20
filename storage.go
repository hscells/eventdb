package eventdb

import (
	"encoding/json"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Store is a persistent store for events.
// It is just a wrapper around a GORM database.
type Store struct {
	db *gorm.DB
}

// Event represents an event in the store.
type Event struct {
	gorm.Model
	EventId string
	Source  string
	Data    []byte
}

// OpenStore opens a new store at the given path.
func OpenStore(path string) (*Store, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(&Event{})
	if err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

// Close closes the store.
func (s *Store) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// AddEvent adds a new event to the store.
func (s *Store) AddEvent(id, source string, data []byte) error {
	return s.db.Create(&Event{EventId: id, Source: source, Data: data}).Error
}

// GetLastEvent returns the most recent event with the given ID.
func (s *Store) GetLastEvent(source, eventid string) (map[string]any, error) {
	var e Event
	err := s.db.Last(&e, "source = ? AND event_id = ?", source, eventid).Error
	var data map[string]any
	err = json.Unmarshal(e.Data, &data)
	return data, err
}
