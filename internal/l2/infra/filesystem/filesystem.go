package filesystem

type (
	Reader interface {
		ReadJSON(path string, target any) error
	}
	Writer interface {
		WriteJSON(path string, data any) error
		WriteBytes(path string, data []byte) error
	}
)
