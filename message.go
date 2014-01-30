package main

type Message interface {
  SetHead(string, string)
  GetHead(string) (string, bool)
  SetBody(string)
  GetBody() string
  Bytes() []byte
  String() string
}
