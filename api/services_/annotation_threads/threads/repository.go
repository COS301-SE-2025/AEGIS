package annotationthreads

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AnnotationThreadRepository struct {
	DB *gorm.DB
}

func NewAnnotationThreadRepository(db *gorm.DB) *AnnotationThreadRepository {
	return &AnnotationThreadRepository{
		DB: db,
	}
}
func (r *AnnotationThreadRepository) CreateThread(thread *AnnotationThread, tags []string) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(thread).Error; err != nil {
			return err
		}
		for _, tag := range tags {
			t := ThreadTag{ThreadID: thread.ID, TagName: tag}
			if err := tx.Create(&t).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *AnnotationThreadRepository) GetThreadsByFile(fileID uuid.UUID) ([]AnnotationThread, error) {
	var threads []AnnotationThread
	err := r.DB.Preload("Tags").Preload("Participants").Where("file_id = ?", fileID).Find(&threads).Error
	return threads, err
}

func (r *AnnotationThreadRepository) GetThreadsByCase(caseID uuid.UUID) ([]AnnotationThread, error) {
	var threads []AnnotationThread
	err := r.DB.Preload("Tags").Preload("Participants").Where("case_id = ?", caseID).Find(&threads).Error
	return threads, err
}

func (r *AnnotationThreadRepository) UpdateThreadStatus(threadID uuid.UUID, status ThreadStatus) error {
	return r.DB.Model(&AnnotationThread{}).Where("id = ?", threadID).Update("status", status).Error
}

func (r *AnnotationThreadRepository) AddParticipant(threadID, userID uuid.UUID) error {
	participant := ThreadParticipant{ThreadID: threadID, UserID: userID}
	return r.DB.FirstOrCreate(&participant, ThreadParticipant{ThreadID: threadID, UserID: userID}).Error
}

func (r *AnnotationThreadRepository) GetThreadParticipants(threadID uuid.UUID) ([]ThreadParticipant, error) {
	var participants []ThreadParticipant
	err := r.DB.Where("thread_id = ?", threadID).Find(&participants).Error
	return participants, err
}

func (r *AnnotationThreadRepository) AddMentions(messageID uuid.UUID, mentions []uuid.UUID) error {
	if len(mentions) == 0 {
		return nil
	}

	query := `INSERT INTO message_mentions (message_id, mentioned_user_id, created_at) VALUES `
	vals := []interface{}{}
	placeholders := []string{}
	now := time.Now()

	for _, userID := range mentions {
		placeholders = append(placeholders, "(?, ?, ?)")
		vals = append(vals, messageID, userID, now)
	}

	query += strings.Join(placeholders, ", ")

	result := r.DB.Exec(query, vals...)
	return result.Error
}

func (r *AnnotationThreadRepository) GetThreadByID(threadID uuid.UUID) (*AnnotationThread, error) {
	var thread AnnotationThread
	err := r.DB.Where("id = ?", threadID).First(&thread).Error
	if err != nil {
		return nil, err
	}
	return &thread, nil
}
func (r *AnnotationThreadRepository) GetUserByID(userID uuid.UUID) (*User, error) {
	var user User
	err := r.DB.Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
