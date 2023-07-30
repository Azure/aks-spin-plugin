package prompt

import (
	"fmt"

	"github.com/manifoldco/promptui"
)

type SelectOpt[T any] struct {
	// Field returns the name to use for each select item.
	Field func(t T) string
}

func Select[T any](label string, items []T, opt *SelectOpt[T]) (T, error) {
	var selections interface{} = items

	if opt != nil && opt.Field != nil {
		new := make([]string, len(items))
		for i, item := range items {
			new[i] = opt.Field(item)
		}

		selections = new
	}

	p := promptui.Select{
		Label: label,
		Items: selections,
	}

	idx, _, err := p.Run()
	if err != nil {
		return *new(T), fmt.Errorf("running select: %w", err)
	}

	return items[idx], nil
}
