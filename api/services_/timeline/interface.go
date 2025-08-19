package timeline

// Repository defines DB-level operations for timeline events.
type Repository interface {
	Create(event *TimelineEvent) error
	GetByID(id string) (*TimelineEvent, error)
	ListByCase(caseID string) ([]*TimelineEvent, error)
	Update(event *TimelineEvent) error
	Delete(id string) error
	UpdateOrder(caseID string, orderedIDs []string) error
	AutoMigrate() error
	FindByID(eventID string) (*TimelineEvent, error)
}
