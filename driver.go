package main

type DatabaseDriver interface {
	authenticate(string, string) error
	register(string, string) error
	getNodeWithParents(*Url) (Object, error)
	createNode(*Url, *Object) error
	updateNode(*Url, *Object) error
	listNodes(*Url) ([]string, error)
	deleteNodes(*Url) error
	readAttachment(*Url) ([]byte, error)
	writeAttachment(*Url, []byte) error
}
