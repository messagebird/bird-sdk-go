package bird

import "iter"

// paginate turns a cursor-paged fetch into a lazy item iterator. fetchPage
// returns one page's items and the next cursor (nil or "" to stop). A fetch
// error is yielded once with a nil item, after which the sequence ends.
func paginate[T any](fetchPage func(cursor string) ([]T, *string, error)) iter.Seq2[*T, error] {
	return func(yield func(*T, error) bool) {
		cursor := ""
		for {
			data, next, err := fetchPage(cursor)
			if err != nil {
				yield(nil, err)
				return
			}
			for i := range data {
				if !yield(&data[i], nil) {
					return
				}
			}
			if next == nil || *next == "" {
				return
			}
			cursor = *next
		}
	}
}
