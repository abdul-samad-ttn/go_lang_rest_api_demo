package constants

const apiVersion = "v1"
const apiBasePath = "/api/" + apiVersion

const (
	HealthCheckPath = apiBasePath + "/health"
)

const (
	GET    = "GET"
	POST   = "POST"
	PUT    = "PUT"
	PATCH  = "PATCH"
	DELETE = "DELETE"
)

const (
	PostsPath    = apiBasePath + "/posts"
	PostByIdPath = apiBasePath + "/posts/{id}"
)

const (
	CommentsPath    = PostsPath + "/{postId}/comments"
	CommentByIdPath = apiBasePath + "/comments/{commentId}"
)

const (
	InternalServerErrorMessage   = "Internal server error"
	InvalidPostIdErrorMessage    = "Invalid post id"
	InvalidCommentIdErrorMessage = "Invalid comment id"
	CommentNotFoundErrorMessage  = "Comment not found"
	PostNotFoundErrorMessage     = "Post not found"
	ForbiddenErrorMessage        = "Forbidden"
)

const (
	InvalidNameErrorMessage   = "Invalid name"
	InvalidEmailErrorMessage  = "Invalid email"
	InvalidBodyErrorMessage   = "Invalid body"
	InvalidUserIdErrorMessage = "Invalid user id"
)
