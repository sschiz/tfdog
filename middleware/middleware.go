package middleware

import tb "gopkg.in/tucnak/telebot.v2"

// Middleware is a function that is called before handlers.
type Middleware func(upd *tb.Update) bool

// BuildMiddlewares puts middlewares in one middleware.
func BuildMiddlewares(middlewares ...Middleware) Middleware {
	return func(upd *tb.Update) (ok bool) {
		for _, middleware := range middlewares {
			if ok = middleware(upd); !ok {
				return ok
			}
		}

		return true
	}
}
