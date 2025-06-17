package database

import (
	"errors"
	"fmt"
	"mime/multipart"
	"time"

	"gorm.io/gorm"
)

type User struct {
	Uuid      string `json:"uuid" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Name      string `json:"name" gorm:"type:varchar(255);"`
	Lastname  string `json:"lastname" gorm:"type:varchar(255);"`
	Nickname  string `json:"nickname" gorm:"type:varchar(255);unique;not null" validate:"required"`
	Website   string `json:"website" gorm:"type:varchar(255);"`
	About     string `json:"about" gorm:"type:text;"`
	Password  string `json:"password" gorm:"type:varchar(255);not null" validate:"required"`
	Email     string `json:"email" gorm:"type:varchar(255);unique;not null" validate:"required"`
	Tel       string `json:"tel" gorm:"type:varchar(255);unique;not null" validate:"required"`
	Role      string `json:"role" gorm:"type:role;default:user;not null"`
	Avatar    string `json:"avatar" gorm:"type:varchar(255)"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (u *UserHandler) CreateTable() error {
	migrator := u.Db.Migrator()
	if !migrator.HasTable(&User{}) {
		err := migrator.CreateTable(&User{})
		if err != nil {
			return err
		}
	}
	return nil
}

func (u *UserHandler) ResetPassword(email, password, repassword string) error {
	tx := u.Db.Begin()

	if tx.Error != nil {
		tx.Rollback()
		return tx.Error
	}

	if err := tx.Model(&User{}).Where("email = ?", email).UpdateColumn("password", Encrypted([]byte(password))).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}
func (u *UserHandler) LoginWithEmail(email string, password string) (string, error) {

	res := u.Db.Model(&User{}).Where("email = ? AND password = ?", email, Encrypted([]byte(password))).Find(&User{})

	if res.Error != nil {
		return "", errors.New("Sorgu çalışırken problem ile karşılaşıldı.")
	}

	if res.RowsAffected == 1 {
		return "Giriş başarılı", nil
	} else {
		return "", fmt.Errorf("Kullanıcı adı veya şifre yanlış")
	}
}
func (u *UserHandler) LoginWithTel(tel string, password string) (string, error) {

	res := u.Db.Model(&User{}).Where("tel = ? AND password = ?", tel, Encrypted([]byte(password))).Find(&User{})

	if res.Error != nil {
		return "", errors.New("Sorgu çalışırken problem ile karşılaşıldı.")
	}

	if res.RowsAffected == 1 {
		return "Giriş başarılı", nil
	} else {
		return "", fmt.Errorf("Kullanıcı adı veya şifre yanlış")
	}
}
func (u *UserHandler) SignUp(user User) []string {

	tx := u.Db.Begin()

	if tx.Error != nil {
		return []string{fmt.Sprintf("Tx başlatılamadı. %s ", tx.Error)}
	}

	user.Password = Encrypted([]byte(user.Password))

	res := tx.Model(&User{}).Create(&user)

	if res.Error != nil {

		errors := []string{}

		type errCheck struct {
			field   string
			message string
		}

		messages := map[string]errCheck{
			"users.nickname": {field: user.Nickname, message: "Bu nickname başka kullanıcı tarafından kullanılmaktadır."},
			"users.email":    {field: user.Email, message: "Bu email adresi başka kullanıcı tarafından kullanılmaktadır."},
			"users.tel":      {field: user.Tel, message: "Bu telefon numarası başka kullanıcı tarafından kullanılmaktadır."},
		}

		var count int

		for field, data := range messages {

			//transaction koptuğunda üzerinden bir daha sql sorgusu yapılamamaktadır.
			u.Db.Model(&User{}).Select("COUNT(*)").Where(fmt.Sprintf("%s = ?", field), data.field).Scan(&count)

			if count >= 1 {
				errors = append(errors, fmt.Sprintf("%s", data.message))
			}
		}
		tx.Rollback()
		return errors
	}
	if err := tx.Commit().Error; err != nil {
		return []string{"Transaction commit işlemi başarısız"}
	}

	return nil
}

func (u *UserHandler) UpdateUser(user User) []string {

	tx := u.Db.Begin()

	if tx.Error != nil {
		tx.Rollback()
		return []string{"Tx başlatılamadı."}
	}

	res := tx.Model(&User{}).Where("uuid = ?", user.Uuid).UpdateColumns(&user)

	if res.Error != nil {

		errors := []string{}

		type errCheck struct {
			field   string
			message string
		}

		messages := map[string]errCheck{
			"users.nickname": {field: user.Nickname, message: "Bu nickname başka kullanıcı tarafından kullanılmaktadır."},
			"users.email":    {field: user.Email, message: "Bu email adresi başka kullanıcı tarafından kullanılmaktadır."},
			"users.tel":      {field: user.Tel, message: "Bu telefon numarası başka kullanıcı tarafından kullanılmaktadır."},
		}

		var count int64

		for field, data := range messages {

			u.Db.Model(&User{}).Select("COUNT(*)").Where(fmt.Sprintf("%s = ?", field), data.field).Scan(&count)

			if count >= 1 {
				errors = append(errors, fmt.Sprintf("%s", data.message))
			}
		}
		tx.Rollback()
		return errors
	}

	if res := tx.Commit(); res.Error != nil {
		tx.Rollback()
		return []string{"Veriler commit edilemedi."}
	}

	return nil
}
func (u *UserHandler) GetUsers() ([]User, error) {

	if err := u.AnyUserExists(); err != nil {
		return nil, err
	}

	rows, err := u.Db.Model(&User{}).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User

	for rows.Next() {
		var user User
		if err := u.Db.ScanRows(rows, &user); err != nil {
			return nil, fmt.Errorf("Satırlar scan edilirken hata ile karşılaşıldı: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
func (u *UserHandler) GetUser(user_uuid string) (User, error) {

	user, err := u.IsUserExistsByUuid(user_uuid)
	if err != nil {
		return User{}, err
	}

	return user, nil
}
func (u *UserHandler) LoadAvatar(user_uuid string, file *multipart.FileHeader) (string, error) {

	allowedTypes := map[string]bool{"image/jpeg": true, "image/png": true, "image/jpg": true}

	if _, exists := allowedTypes[file.Header.Get("Content-Type")]; !exists {
		return "", errors.New("Dosya istenilen formatta değil (sadece JPEG veya PNG desteklenmektedir).")
	}

	if file.Size > 3*1024*1024 {
		return "", errors.New("Dosya boyutu 3 MB'ı aşmamalıdır.")
	}

	tx := u.Db.Begin()
	if tx.Error != nil {
		return "", fmt.Errorf("Transaction başlatılamadı: %w", tx.Error)
	}

	if _, err := u.IsUserExistsByUuid(user_uuid); err != nil {
		return "", fmt.Errorf("Transaction başlatılamadı: %w", tx.Error)
	}

	res := tx.Model(&User{}).Where("uuid = ?", user_uuid).Update("avatar", EncryptedFileName(file))
	if res.Error != nil {
		tx.Rollback()
		return "", fmt.Errorf("Avatar güncellenirken hata oluştu: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		tx.Rollback()
		return "", errors.New("Avatar güncellemesi başarısız. UUID bulunamadı.")
	}

	if err := SaveImage(file); err != nil {
		tx.Rollback()
		return "", fmt.Errorf("Avatar kaydedilirken hata oluştu: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return "", fmt.Errorf("Değişiklikler kaydedilemedi: %w", err)
	}

	return "Avatar başarıyla güncellendi", nil
}
func (u *UserHandler) DeleteAvatar(user_uuid string) error {
	tx := u.Db.Begin()

	if tx.Error != nil {
		return fmt.Errorf("Tx başlatılamadı. %w ", tx.Error)
	}

	user, err := u.IsUserExistsByUuid(user_uuid)
	if err != nil {
		return err
	}

	if user.Avatar == "" {
		tx.Rollback()
		return fmt.Errorf("Kullanıcının kaldırılacak avatarı bulunmamaktadır.")
	}

	res := tx.Model(&User{}).Where("uuid = ?", user_uuid).UpdateColumn("avatar", nil)
	if res.Error != nil {
		tx.Rollback()
		return fmt.Errorf("Sorgu çalıştırılırken hata ile karşılaşıldı.")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("Değişiklikler kaydedilemedi: %w", err)
	}

	return nil

}
func (u *UserHandler) IsUserExists(field string) (User, error) {

	var user User

	result := u.Db.Where("email = ?", field).Or("tel = ?", field).First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return User{}, errors.New("Girdiğiniz bilgilere ait kullanıcı bulunmamaktadır.")
		}
		return User{}, fmt.Errorf("Sorgu çalıştırılırken hata ile karşılaşıldı.%w", result.Error)
	}

	return user, nil

}
func (u *UserHandler) IsUserExistsByUuid(uuid string) (User, error) {

	var user User

	result := u.Db.Where("uuid = ?", uuid).First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return User{}, errors.New("Girdiğiniz bilgilere ait kullanıcı bulunmamaktadır.")
		}
		return User{}, fmt.Errorf("Sorgu çalıştırılırken hata ile karşılaşıldı.%w", result.Error)
	}

	return user, nil

}
func (u *UserHandler) AnyUserExists() error {

	var count int64

	u.Db.Model(&User{}).Select("COUNT(*)").Scan(&count)
	if count <= 0 {
		return errors.New("Hiç kullanıcı bulunmamaktadır.")
	}

	return nil

}
