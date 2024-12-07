package middleware

import "github.com/BulizhnikGames/subbot/bot/tools"

func CreateStack(middlewares ...tools.MiddlewareFunc) tools.MiddlewareFunc {
	return func(next tools.Command) tools.Command {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}
