package store

import (
	"database/sql"

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
