package database

import (
	"fmt"
)

func (a *AdminHandler) GetAllStatics() (map[string]int, error) {

	countEntities := map[string]any{
		"Kullanıcı Adedi": &User{},
		"Soru Adedi":      &Question{},
		"Yorum Adedi":     &Comment{},
	}

	results := make(map[string]int)

	for key, model := range countEntities {
		var count int64
		if err := a.Db.Model(model).Count(&count).Error; err != nil {
			return nil, fmt.Errorf("%s alınamadı: %v", key, err)
		}
		results[key] = int(count)
	}

	return results, nil
}
func (u *AdminHandler) DeleteUser(user_uuid string) error {
	tx := u.Db.Begin()

	if tx.Error != nil {
		tx.Rollback()
		return fmt.Errorf("Tx başlatılamadı. %w ", tx.Error)
	}

	relatedTables := []string{
		"comment_reactions",
		"comments",
		"questions",
	}

	for _, table := range relatedTables {
		if err := tx.Exec("DELETE FROM "+table+" WHERE user_uuid = ?", user_uuid).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("%s tablosundan veriler silinemedi: %w", table, err)
		}
	}

	if res := tx.Where("uuid= ?", user_uuid).Delete(&User{}); res.Error != nil {
		tx.Rollback()
		return fmt.Errorf("Sorgu çalıştırılırken hata ile karşılaşıldı.%w", res.Error)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("Commit hatası : %w ", tx.Error)
	}

	return nil
}
func (u *AdminHandler) GetSixMonthsData() (map[string][]int64, error) {
	months := GetSixMonths()
	months_datas := make(map[string][]int64)

	var question_count int64 = 0
	var comment_count int64 = 0

	for _, month := range months {

		err := u.Db.Table("questions").
			Select("COUNT(*)").
			Where("TO_CHAR(created_at, 'YYYY-MM') = ?", month).
			Scan(&question_count).Error
		if err != nil {
			return nil, fmt.Errorf("questions için hata (%s): %v", month, err)
		}

		err = u.Db.Table("comments").
			Select("COUNT(*)").
			Where("TO_CHAR(created_at, 'YYYY-MM') = ?", month).
			Scan(&comment_count).Error
		if err != nil {
			return nil, fmt.Errorf("comments için hata (%s): %v", month, err)
		}

		months_datas[month] = []int64{question_count, comment_count}
	}

	return months_datas, nil
}
