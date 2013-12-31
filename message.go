package main

type Message interface {
  SetHead(string, string)
  GetHead(string) string
  SetBody(string)
  GetBody() string
  Bytes() []byte
  String() string
}
