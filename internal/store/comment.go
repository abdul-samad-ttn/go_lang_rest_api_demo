package store

import (
	"database/sql"
	"errors"

	"github.com/abdul-samad-ttn/go_lang_rest_api_demo.git/constants"
	"github.com/abdul-samad-ttn/go_lang_rest_api_demo.git/internal/models"
)

type CommentStore struct {
	db *sql.DB
}

func NewCommentStore(db *sql.DB) *CommentStore {
	return &CommentStore{db: db}
}

func (s *CommentStore) ListComments(postId int64) ([]models.Comment, error) {
	query := "SELECT id, name, email, body, post_id FROM comments WHERE post_id = ? ORDER BY id ASC"

	rows, queryError := s.db.Query(query, postId)
	if queryError != nil {
		return nil, queryError
	}

	defer rows.Close()

	comments := make([]models.Comment, 0)

	for rows.Next() {
		var comment models.Comment

		scanError := rows.Scan(&comment.Id, &comment.Name, &comment.Email, &comment.Body, &comment.PostId)
		if scanError != nil {
			return nil, scanError
		}
		comments = append(comments, comment)
	}

	rowsError := rows.Err()
	if rowsError != nil {
		return nil, rowsError
	}

	return comments, nil
}

func (s *CommentStore) GetCommentById(id int64) (*models.Comment, error) {
	query := "SELECT id, name, email, body, post_id FROM comments WHERE id = ?"

	row := s.db.QueryRow(query, id)

	if row.Err() != nil {
		return nil, row.Err()
	}

	var comment models.Comment

	scanError := row.Scan(&comment.Id, &comment.Name, &comment.Email, &comment.Body, &comment.PostId)

	if scanError != nil {
		return nil, scanError
	}

	return &comment, nil
}

func (s *CommentStore) CreateComment(postId int64, comment models.CreateCommentRequest) (*models.Comment, error) {
	query := "INSERT INTO comments (post_id, name, email, body, author_user_id, created_at) VALUES (?, ?, ?, ?, ?, datetime('now'))"

	result, queryError := s.db.Exec(query, postId, comment.Name, comment.Email, comment.Body, comment.UserId)

	if queryError != nil {
		return nil, queryError
	}

	commentId, insertError := result.LastInsertId()

	if insertError != nil {
		return nil, insertError
	}

	newComment := models.Comment{
		Id:     commentId,
		Name:   comment.Name,
		Email:  comment.Email,
		Body:   comment.Body,
		PostId: postId,
	}

	return &newComment, nil
}

func (s *CommentStore) DeleteComment(commentId int64, userId int64) error {
	query := "DELETE FROM comments WHERE id = ? AND author_user_id = ?"

	result, queryError := s.db.Exec(query, commentId, userId)

	if queryError != nil {
		return queryError
	}

	rowsAffected, deleteError := result.RowsAffected()

	if deleteError != nil {
		if errors.Is(deleteError, sql.ErrNoRows) {
			return errors.New(constants.CommentNotFoundErrorMessage)
		}
		return deleteError
	}

	if rowsAffected == 0 {
		return errors.New(constants.CommentNotFoundErrorMessage)
	}

	return nil
}
