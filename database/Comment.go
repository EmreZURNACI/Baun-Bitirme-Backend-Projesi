package database

import (
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Comment struct {
	Uuid          string            `json:"uuid" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Question_uuid string            `json:"question_uuid" gorm:"type:uuid" validate:"required"`
	User_uuid     string            `json:"user_uuid" gorm:"type:uuid" validate:"required"`
	Comment       string            `json:"comment" gorm:"type:text" validate:"required"`
	Image         pq.StringArray    `json:"image" gorm:"type:varchar(255)[];default:null;" validate:"omitempty"`
	LikesDislikes []CommentReaction `gorm:"foreignKey:CommentUuid" json:"likes_dislikes"`
	Question      Question          `json:"question" gorm:"foreignKey:Question_uuid" validate:"-"`
	User          User              `json:"user" gorm:"foreignKey:User_uuid" validate:"-"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

type CommentReaction struct {
	Uuid        string `json:"uuid" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CommentUuid string `json:"comment_uuid" gorm:"type:uuid;not null" validate:"required"`
	UserUuid    string `json:"user_uuid" gorm:"type:uuid;not null" validate:"required"`
	IsLike      bool   `json:"is_like" gorm:"not null" validate:"required"` // true = like, false = dislike

	User    User    `gorm:"foreignKey:UserUuid;references:Uuid" json:"user"`
	Comment Comment `gorm:"foreignKey:CommentUuid;references:Uuid" json:"comment"`

	CreatedAt time.Time
}

func (c *CommentHandler) CreateTable() error {
	migrator := c.Db.Migrator()
	if !migrator.HasTable(&Comment{}) {
		err := migrator.CreateTable(&Comment{})
		if err != nil {
			return err
		}
	}

	if !migrator.HasTable(&CommentReaction{}) {
		err := migrator.CreateTable(&CommentReaction{})
		if err != nil {
			return err
		}
	}
	return nil

}

func (c *CommentHandler) GetComments(uuid string) ([]Comment, error) {

	questionHandler := GetQuestionHandler(c.Db)

	if _, err := questionHandler.IsQuestionExistByUuid(uuid); err != nil {
		return nil, err
	}

	var comments []Comment

	res := c.Db.
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Omit("password")
		}).
		Preload("Question").
		Joins("INNER JOIN questions ON comments.question_uuid = questions.uuid").
		Where("questions.uuid = ?", uuid).
		Order("comments.created_at DESC").
		Find(&comments)

	if res.Error != nil {
		return nil, errors.New("Sorgu çalıştırılırken hata ile karşılaşıldı.")
	}

	if len(comments) == 0 {
		return nil, errors.New("Soruya ait yorum bulunmamaktadır.")
	}
	return comments, nil
}
func (c *CommentHandler) AddComment(comment Comment) (*gorm.DB, error) {

	tx := c.Db.Begin()

	if tx.Error != nil {
		tx.Rollback()
		return tx, fmt.Errorf("Tx başlatılamadı. %w", tx.Error)
	}

	res := tx.Model(&Comment{}).Create(&comment)
	if res.Error != nil {
		tx.Rollback()
		if res.Error.Error() == "pq: insert or update on table \"comments\" violates foreign key constraint \"fk_comments_user\"" {
			return nil, errors.New("Bu uuid'ye ait kullanıcı bulunmamaktadır.")
		}
		return tx, fmt.Errorf("Yorum ekleme sorgusu çalışırken hata ile karşılaşıldı. %w", res.Error)
	}

	if res.RowsAffected == 0 {
		tx.Rollback()
		return tx, errors.New("İlgili yorum silinemedi.")
	}

	if res := tx.Commit(); res.Error != nil {
		return tx, fmt.Errorf("Değişiklikler commit edilemedi. %w", res.Error)
	}

	return tx, nil

}
func (c *CommentHandler) UpdateComment(comment Comment, user_uuid string) error {
	tx := c.Db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("Transaction başlatılamadı: %w", tx.Error)
	}

	_comment, err := c.IsCommentExistByUuid(comment.Uuid)
	if err != nil {
		tx.Rollback()
		return errors.New("Yorum bulunamadı veya doğrulanamadı")
	}

	user, err := GetUserHandler(tx).IsUserExistsByUuid(user_uuid)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("Kullanıcı doğrulanamadı: %w", err)
	}

	if user.Role != "admin" && _comment.User_uuid != user.Uuid {
		tx.Rollback()
		return errors.New("Bu yorumu düzenleme yetkiniz bulunmamaktadır")
	}

	res := tx.Model(&Comment{}).Where("uuid = ?", _comment.Uuid).Updates(comment)
	if res.Error != nil {
		tx.Rollback()
		return fmt.Errorf("Yorum düzenlenirken hata oluştu: %w", res.Error)
	}

	if res.RowsAffected == 0 {
		tx.Rollback()
		return errors.New("Yorum güncellenemedi")
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("Değişiklikler commit edilemedi: %w", err)
	}

	return nil
}
func (c *CommentHandler) DeleteComment(uuid, userUUID string) error {
	tx := c.Db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("Tx başlatılamadı: %w", tx.Error)
	}

	comment, _ := c.IsCommentExistByUuid(uuid)

	userHandler := GetUserHandler(tx)
	user, err := userHandler.IsUserExistsByUuid(userUUID)
	if err != nil {
		tx.Rollback()
		return err
	}

	if user.Role != "admin" && comment.User_uuid != user.Uuid {
		tx.Rollback()
		return errors.New("Bu yorumu silme yetkiniz bulunmamaktadır")
	}

	res := tx.Where("uuid = ?", uuid).Delete(&Comment{})
	if res.Error != nil {
		tx.Rollback()
		return fmt.Errorf("Yorum silinirken hata oluştu: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		tx.Rollback()
		return errors.New("İlgili yorum bulunamadı veya silinemedi")
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("Değişiklikler commit edilemedi: %w", err)
	}

	return nil
}
func (c *CommentHandler) IsCommentExistByUuid(uuid string) (Comment, error) {

	var comment Comment

	result := c.Db.Table("comments").Where("uuid = ?", uuid).First(&comment)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return Comment{}, errors.New("Girdiğiniz bilgilere ait yorum bulunmamaktadır.")
		}
		return Comment{}, fmt.Errorf("Sorgu çalıştırılırken hata ile karşılaşıldı.%w", result.Error)
	}

	return comment, nil
}
func (c *CommentHandler) ReactionCount(comment_uuid string) ([]int64, error) {

	type Count struct {
		Like_count    int64 `json:"like_count"`
		Dislike_count int64 `json:"dislike_count"`
	}

	var counts Count

	tx := c.Db.Table("comment_reactions").
		Select(`
		COUNT(CASE WHEN is_like = true THEN 1 END) AS like_count,
		COUNT(CASE WHEN is_like = false THEN 1 END) AS dislike_count`).
		Where("comment_uuid = ?", comment_uuid).
		Scan(&counts)

	if tx.Error != nil {
		return []int64{}, errors.New("Sorgu çalıştırılırken hata ile karşılaşıldı.")
	}

	return []int64{counts.Like_count, counts.Dislike_count}, nil
}
func (c *CommentHandler) LikeComment(commentUUID, userUUID string) error {
	tx := c.Db.Begin()
	if tx.Error != nil {
		return errors.New("veritabanı işlemi başlatılamadı")
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	defer tx.Rollback()

	var reaction CommentReaction
	err := tx.Table("comment_reactions").
		Where("user_uuid = ? AND comment_uuid = ?", userUUID, commentUUID).
		First(&reaction).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = tx.Table("comment_reactions").Create(&CommentReaction{
				UserUuid:    userUUID,
				CommentUuid: commentUUID,
				IsLike:      true,
			}).Error
		} else {
			return errors.New("sorgu çalıştırılırken hata oluştu")
		}
	} else {
		if reaction.IsLike {
			err = tx.Table("comment_reactions").
				Where("user_uuid = ? AND comment_uuid = ?", userUUID, commentUUID).
				Delete(&reaction).Error
		} else {
			err = tx.Table("comment_reactions").
				Where("user_uuid = ? AND comment_uuid = ?", userUUID, commentUUID).
				UpdateColumn("is_like", true).Error
		}
	}

	if err != nil {
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return errors.New("işlem veritabanına kaydedilemedi")
	}

	return nil
}
func (c *CommentHandler) DislikeComment(commentUUID, userUUID string) error {
	tx := c.Db.Begin()
	if tx.Error != nil {
		return errors.New("veritabanı işlemi başlatılamadı")
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	defer tx.Rollback()

	var reaction CommentReaction
	err := tx.Table("comment_reactions").
		Where("user_uuid = ? AND comment_uuid = ?", userUUID, commentUUID).
		First(&reaction).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = tx.Table("comment_reactions").Create(&CommentReaction{
				UserUuid:    userUUID,
				CommentUuid: commentUUID,
				IsLike:      false,
			}).Error
		} else {
			return errors.New("sorgu çalıştırılırken hata oluştu")
		}
	} else {
		if !reaction.IsLike {
			err = tx.Table("comment_reactions").
				Where("user_uuid = ? AND comment_uuid = ?", userUUID, commentUUID).
				Delete(&reaction).Error
		} else {
			err = tx.Table("comment_reactions").
				Where("user_uuid = ? AND comment_uuid = ?", userUUID, commentUUID).
				UpdateColumn("is_like", false).Error
		}
	}

	if err != nil {
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return errors.New("işlem veritabanına kaydedilemedi")
	}

	return nil
}
