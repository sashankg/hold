package core

import "net/http"

type Route interface {
	http.Handler
	Route() string
}
