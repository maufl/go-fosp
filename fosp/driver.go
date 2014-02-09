package fosp

type DatabaseDriver interface {
	Authenticate(string, string) error
	Register(string, string) error
	GetNodeWithParents(*Url) (Object, error)
	CreateNode(*Url, *Object) error
	UpdateNode(*Url, *Object) error
	ListNodes(*Url) ([]string, error)
	DeleteNodes(*Url) error
	ReadAttachment(*Url) ([]byte, error)
	WriteAttachment(*Url, []byte) error
}
