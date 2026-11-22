package store

import (
	"database/sql"

	"github.com/abdul-samad-ttn/go_lang_rest_api_demo.git/internal/models"
)

type PostStore struct {
	db *sql.DB
}

func NewPostStore(db *sql.DB) *PostStore {
	return &PostStore{db: db}
}

func (s *PostStore) ListPosts() ([]models.Post, error) {
	query := "SELECT id, title, body, user_id FROM posts ORDER BY id ASC"

	rows, err := s.db.Query(query)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var posts []models.Post

	for rows.Next() {
		var post models.Post
		err = rows.Scan(&post.Id, &post.Title, &post.Body, &post.UserId)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return posts, nil
}

func (s *PostStore) GetPostById(id int64) (*models.Post, error) {
	query := "SELECT id, title, body, user_id FROM posts WHERE id = ?"

	row := s.db.QueryRow(query, id)

	var post models.Post

	err := row.Scan(&post.Id, &post.Title, &post.Body, &post.UserId)
	if err != nil {
		return nil, err
	}

	return &post, nil
}
