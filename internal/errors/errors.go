package cstmerr

import "errors"

var (
	ErrorConnection           = errors.New("connection error")
	ErrorServer               = errors.New("server error")
	ErrorNonRetriable         = errors.New("non retriable error")
	ErrorRetriable            = errors.New("retriable error")
	ErrorNonRetriablePostgres = errors.New("non retriable postgres error")
	ErrorRetriablePostgres    = errors.New("retriable postgres error")
	ErrorUserAlreadyExists    = errors.New("user already exists")
	ErrorUserDoesNotExist     = errors.New("user does not exist")
)
