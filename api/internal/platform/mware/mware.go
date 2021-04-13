package mware

import "github.com/julienschmidt/httprouter"

type Middleware func(httprouter.Handle) httprouter.Handle

type Chain struct {
	middlewares []Middleware
}

func NewChain(mwares ...Middleware) Chain {
	return Chain{middlewares: mwares}
}

func (chain Chain) With(mwares ...Middleware) Chain {
	chain.middlewares = append(chain.middlewares, mwares...)
	return chain
}

func (chain Chain) Wrap(handler httprouter.Handle) httprouter.Handle {
	for i := len(chain.middlewares) - 1; i >= 0; i-- {
		handler = chain.middlewares[i](handler)
	}
	return handler
}
