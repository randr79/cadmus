package cadmus

import "context"

type key struct{}

func ContextWithDispatcher(ctx context.Context, dispatcher *dispatcher) context.Context {
	return context.WithValue(ctx, key{}, dispatcher)
}

func DispatcherFromContext(ctx context.Context) (*dispatcher, bool) {
	v := ctx.Value(key{})
	m, ok := v.(*dispatcher)
	return m, ok
}
