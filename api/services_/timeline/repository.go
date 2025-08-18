package timeline

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type repo struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repo{db: db}
}

func (r *repo) AutoMigrate() error {
	return r.db.AutoMigrate(&TimelineEvent{})
}

func (r *repo) Create(event *TimelineEvent) error {
	// If Order not set, set it to last+1 for the case
	if event.Order == 0 {
		var maxOrder int
		// use COALESCE to handle no rows
		r.db.Model(&TimelineEvent{}).
			Where("case_id = ?", event.CaseID).
			Select("COALESCE(MAX(\"order\"), 0)").
			Scan(&maxOrder)
		event.Order = maxOrder + 1
	}
	if err := r.db.Create(event).Error; err != nil {
		return err
	}
	return nil
}

func (r *repo) GetByID(id string) (*TimelineEvent, error) {
	var ev TimelineEvent
	if err := r.db.First(&ev, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &ev, nil
}

func (r *repo) ListByCase(caseID string) ([]*TimelineEvent, error) {
	var events []*TimelineEvent
	if err := r.db.
		Where("case_id = ?", caseID).
		Order("created_at ASC").
		Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

func (r *repo) Update(event *TimelineEvent) error {
	if event.ID == "" {
		return errors.New("missing id on update")
	}
	return r.db.Model(&TimelineEvent{}).Where("id = ?", event.ID).Updates(event).Error
}

func (r *repo) Delete(id string) error {
	return r.db.Delete(&TimelineEvent{}, "id = ?", id).Error
}

// UpdateOrder sets events order in the DB according to orderedIDs slice.
// It uses a transaction and updates the "order" column.
func (r *repo) UpdateOrder(caseID string, orderedIDs []string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for idx, id := range orderedIDs {
			res := tx.Model(&TimelineEvent{}).
				Where("id = ? AND case_id = ?", id, caseID).
				Update("order", idx+1)
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return fmt.Errorf("event %s not found for case %s", id, caseID)
			}
		}
		return nil
	})
}
func (r *repo) FindByID(eventID string) (*TimelineEvent, error) {
	var event TimelineEvent
	if err := r.db.First(&event, "id = ?", eventID).Error; err != nil {
		return nil, err
	}
	return &event, nil
}
