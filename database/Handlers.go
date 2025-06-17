package database

import "gorm.io/gorm"

type UserHandler struct {
	Db *gorm.DB
}

func GetUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{Db: db}
}

type QuestionHandler struct {
	Db *gorm.DB
}

func GetQuestionHandler(db *gorm.DB) *QuestionHandler {
	return &QuestionHandler{Db: db}
}

type CommentHandler struct {
	Db *gorm.DB
}

func GetCommentHandler(db *gorm.DB) *CommentHandler {
	return &CommentHandler{Db: db}
}

type CodeHandler struct {
	Db *gorm.DB
}

func GetCodeHandler(db *gorm.DB) *CodeHandler {
	return &CodeHandler{Db: db}
}

type TagHandler struct {
	Db *gorm.DB
}

func GetTagHandler(db *gorm.DB) *TagHandler {
	return &TagHandler{Db: db}
}

type AdminHandler struct {
	Db *gorm.DB
}

func GetAdminHandler(db *gorm.DB) *AdminHandler {
	return &AdminHandler{
		Db: db,
	}
}
