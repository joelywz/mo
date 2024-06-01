package server

import "github.com/labstack/echo/v4"

type RouteHandler[T any] func(group *echo.Group, params T)

func NewRoute[T any](path string, handler RouteHandler[T]) func(e *echo.Echo, params T) {
	return func(e *echo.Echo, params T) {
		group := e.Group(path)
		handler(group, params)
	}
}
