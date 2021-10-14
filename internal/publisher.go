package internal

import "context"

type Publisher interface {
	Publish(context.Context, Item) error
	Skip(context.Context, Item)
}
