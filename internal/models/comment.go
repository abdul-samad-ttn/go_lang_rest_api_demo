package models

type Comment struct {
	Id     int64  `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Body   string `json:"body"`
	PostId int64  `json:"postId"`
}

type CreateCommentRequest struct {
	Name   string `json:"name"`
	Email  string `json:"email"`
	Body   string `json:"body"`
	UserId int64  `json:"userId"`
}
