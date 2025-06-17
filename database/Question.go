package database

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Question struct {
	Uuid       string         `json:"uuid" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Header     string         `json:"header" gorm:"type:varchar(255);" validate:"required"`
	Content    string         `json:"content" gorm:"type:text;" validate:"required"`
	Image      pq.StringArray `json:"image" gorm:"type:varchar(255)[];default:null;" validate:"omitempty"`
	User_uuid  string         `json:"user_uuid" gorm:"type:uuid;"`
	User       User           `gorm:"foreignKey:User_uuid;constraint:OnDelete:CASCADE;" validate:"-"`
	Tags       []Tag          `gorm:"many2many:questions_tags"`
	ViewsCount int64          `json:"views_count" validate:"-" gorm:"type:bigint;"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

func (q *QuestionHandler) CreateTable() error {
	migrator := q.Db.Migrator()
	if !migrator.HasTable(&Question{}) {
		err := migrator.CreateTable(&Question{})
		if err != nil {
			return err
		}
	}

	return nil
}

func (q *QuestionHandler) GetQuestions(tags, sort string, limit, offset int32) ([]Question, error) {
	var count int64
	if err := q.Db.Model(&Question{}).Count(&count).Error; err != nil {
		return nil, fmt.Errorf("Veri tabanında soru sayısı alınırken hata oluştu: %v", err)
	}

	if count == 0 {
		return nil, fmt.Errorf("Veri tabanında hiç soru bulunmamaktadır.")
	}

	var questions []Question

	query := q.Db.Model(&Question{}).
		Preload("Tags").
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Omit("password")
		}).Order("created_at DESC")

	tags = strings.ReplaceAll(tags, "%20", " ")

	if tags != "" {

		tagList := strings.Split(tags, ",")

		for i, tag := range tagList {
			tagList[i] = strings.TrimSpace(tag)
		}

		query = query.Joins("JOIN questions_tags qt ON qt.question_uuid = questions.uuid").
			Joins("JOIN tags t ON t.uuid = qt.tag_uuid").
			Where("t.name IN ?", tagList).
			Group("questions.uuid")
	}

	if sort == "asc" {
		query = query.Order("questions.created_at ASC")
	} else {
		query = query.Order("questions.created_at DESC")
	}

	if limit > 0 {
		query = query.Limit(int(limit))
	}
	if offset > 0 {
		query = query.Offset(int(offset))
	}

	if err := query.Find(&questions).Error; err != nil {
		return nil, fmt.Errorf("Sorular getirilirken hata oluştu: %v", err)
	}

	return questions, nil
}
func (q *QuestionHandler) GetQuestion(uuid string) (Question, error) {
	var question Question
	res := q.Db.
		Preload("Tags").
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Omit("password")
		}).
		Where("uuid = ?", uuid).
		First(&question).Error

	if res != nil {
		if errors.Is(res, gorm.ErrRecordNotFound) {
			return Question{}, fmt.Errorf("Bu UUID'ye sahip soru bulunmamaktadır.")
		}
		return Question{}, fmt.Errorf("Soru getirilirken hata oluştu: %v", res)
	}

	return question, nil
}
func (q *QuestionHandler) GetQuestionsByUser(userUUID string) ([]Question, error) {
	var count int64
	if err := q.Db.Model(&User{}).Where("uuid = ?", userUUID).Count(&count).Error; err != nil {
		return nil, fmt.Errorf("Kullanıcı kontrolü sırasında hata oluştu: %v", err)
	}

	if count == 0 {
		return nil, errors.New("Bu UUID'ye ait kullanıcı bulunmamaktadır.")
	}

	var questions []Question

	err := q.Db.
		Preload("Tags").
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Omit("password")
		}).
		Joins("INNER JOIN users ON users.uuid = questions.user_uuid").
		Where("users.uuid = ?", userUUID).
		Order("created_at DESC").
		Find(&questions).
		Error

	if err != nil {
		return nil, fmt.Errorf("Sorgu çalıştırılırken hata oluştu: %v", err)
	}

	if len(questions) == 0 {
		return nil, errors.New("Kullanıcıya ait hiç soru bulunmamaktadır.")
	}

	return questions, nil
}
func (q *QuestionHandler) CreateQuestion(question Question) error {

	tx := q.Db.Begin()
	if err := tx.Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("Tx başlatılamadı: %w", err)
	}

	var count int64
	if err := tx.Model(&User{}).Where("uuid = ?", question.User_uuid).Count(&count).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("Kullanıcı kontrolü sırasında hata oluştu: %w", err)
	}

	if count == 0 {
		tx.Rollback()
		return errors.New("Bu UUID'ye ait kullanıcı bulunmamaktadır")
	}

	if err := tx.Create(&question).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("Soru oluşturulurken hata oluştu: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("Transaction commit işlemi başarısız: %w", err)
	}

	return nil

}
func (q *QuestionHandler) UpdateQuestion(question Question) error {

	tx := q.Db.Begin()

	if tx.Error != nil {
		tx.Rollback()
		return fmt.Errorf("Tx başlatılamadı. %w ", tx.Error)
	}

	if res := tx.Model(&Question{}).Where("uuid = ?", question.Uuid).UpdateColumns(&question); res.Error != nil {
		tx.Rollback()

		return res.Error
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("Transaction commit işlemi başarısız: %w", err)
	}

	return nil

}
func (q *QuestionHandler) DeleteQuestion(question_uuid, user_uuid string) error {

	tx := q.Db.Begin()

	userHandler := GetUserHandler(tx)

	question, err := q.IsQuestionExistByUuid(question_uuid)
	if err != nil {
		return err
	}

	user, err := userHandler.IsUserExistsByUuid(user_uuid)
	if err != nil {
		return err
	}

	if tx.Error != nil {
		tx.Rollback()
		return fmt.Errorf("Tx başlatılamadı. %w ", tx.Error)
	}

	if !(user.Role == "admin" || question.User_uuid == user.Uuid) {
		tx.Rollback()
		return errors.New("Sizin bu soruyu silme yetkiniz bulunmamaktadır.")
	}

	if res := tx.Model(&Question{}).Where("uuid = ?", question.Uuid).Delete(&Question{}); res.Error != nil {
		tx.Rollback()
		return fmt.Errorf("Sorgu çalıştırılırken hata ile karşılaşıldı : %w", res.Error)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("Transaction commit işlemi başarısız: %w", err)
	}

	return nil
}
func (q *QuestionHandler) DeleteImages(question_uuid string, user_uuid string, image_list []string) error {

	tx := q.Db.Begin()

	if tx.Error != nil {
		return fmt.Errorf("Tx başlatılamadı. %w ", tx.Error)
	}

	question, err := q.IsQuestionExistByUuid(question_uuid)
	if err != nil {
		return err
	}

	userHandler := GetUserHandler(tx)

	user, err := userHandler.IsUserExistsByUuid(user_uuid)
	if err != nil {
		return err
	}

	if !(user.Role == "admin" || question.User_uuid == user.Uuid) {
		tx.Rollback()
		return errors.New("Sizin bu imageleri silme yetkiniz bulunmamaktadır.")
	}

	for _, value := range image_list {
		res := tx.Model(&Question{}).Where("uuid = ?", question_uuid).
			UpdateColumn("image", gorm.Expr("array_remove(image, ?)", value))

		if res.Error != nil {
			return fmt.Errorf("Imageleri düzenleme sorgusu çalışırken bir problmem çıktı. %w ", res.Error)
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("Imageler düzenlenmedi. %w ", res.Error)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("Transaction commit işlemi başarısız: %w", err)
	}

	return nil
}
func (q *QuestionHandler) IsQuestionExistByUuid(uuid string) (Question, error) {

	var question Question

	result := q.Db.Where("uuid = ?", uuid).First(&question)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return Question{}, errors.New("Girdiğiniz bilgilere ait soru bulunmamaktadır.")
		}
		return Question{}, fmt.Errorf("Sorgu çalıştırılırken hata ile karşılaşıldı.%w", result.Error)
	}

	return question, nil

}
func (q *QuestionHandler) IncreaseQuestionsView(uuid string) (int64, error) {

	tx := q.Db.Begin()

	if tx.Error != nil {
		return 0, tx.Error
	}

	res := tx.Exec(`
		UPDATE public.questions
		SET views_count = views_count + 1
		WHERE uuid = @uuid
	`, sql.Named("uuid", uuid))

	if res.Error != nil {
		tx.Rollback()
		return 0, res.Error
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return 0, err
	}

	return res.RowsAffected, nil
}
