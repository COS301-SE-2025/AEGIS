package annotationthreads

import (
	"time"

	"github.com/google/uuid"
)

type ThreadStatus string
type ThreadPriority string

const (
	StatusOpen            ThreadStatus = "open"             // The thread is open and can be replied to
	StatusClosed          ThreadStatus = "closed"           // The thread is closed and cannot be replied to
	StatusResolved        ThreadStatus = "resolved"         // The thread is resolved and can be closed
	StatusArchived        ThreadStatus = "archived"         // The thread is archived and cannot be replied to
	StatusPendingApproval ThreadStatus = "pending_approval" // The thread is pending approval

	PriorityHigh   ThreadPriority = "high"   // High priority thread
	PriorityMedium ThreadPriority = "medium" // Medium priority thread
	PriorityLow    ThreadPriority = "low"    // Low priority thread
)

type AnnotationThread struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primary_key"`
	Title      string         `json:"title" gorm:"type:varchar(255);not null"`
	FileID     uuid.UUID      `json:"file_id" gorm:"type:uuid;not null"`
	CaseID     uuid.UUID      `json:"case_id" gorm:"type:uuid;not null"`
	CreatedBy  uuid.UUID      `json:"created_by" gorm:"type:uuid;not null"`
	CreatedAt  time.Time      `json:"created_at" gorm:"type:timestamp;default:current_timestamp"`
	UpdatedAt  time.Time      `json:"updated_at" gorm:"type:timestamp;default:current_timestamp"`
	Status     ThreadStatus   `json:"status" gorm:"type:varchar(50);default:'open'"`
	Priority   ThreadPriority `json:"priority" gorm:"type:varchar(50);default:'medium'"`
	IsActive   bool           `json:"is_active" gorm:"type:boolean;default:true"`
	ResolvedAt *time.Time     `json:"resolved_at,omitempty" gorm:"type:timestamp;default:null"`

	Tags         []ThreadTag         `gorm:"foreignKey:ThreadID"`
	Participants []ThreadParticipant `gorm:"foreignKey:ThreadID"`
}

type ThreadTag struct {
	ID       uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primary_key"`
	ThreadID uuid.UUID `json:"thread_id" gorm:"type:uuid;not null"`
	TagName  string    `json:"tag_name" gorm:"type:varchar(255);not null"`
}

type ThreadParticipant struct {
	ThreadID uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID   uuid.UUID `gorm:"type:uuid;primaryKey"`
	JoinedAt time.Time `gorm:"autoCreateTime"`
}

type User struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey"`
	FullName          string    `gorm:"not null"` // This is a derived field, not stored in the DB
	Email             string    `gorm:"uniqueIndex"`
	PasswordHash      string
	Role              string    `gorm:"type:user_role"`
	CreatedAt         time.Time `gorm:"autoCreateTime"`
	IsVerified        bool
	VerificationToken string //We send the token to the user’s email as a verification link, e.g.:
	EmailVerifiedAt   *time.Time
	AcceptedTermsAt   *time.Time
}

// Force GORM to use the existing "threads" table
func (AnnotationThread) TableName() string {
	return "threads"
}
