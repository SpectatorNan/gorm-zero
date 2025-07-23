package helper

import "github.com/SpectatorNan/gorm-zero/v2/pagex"

// Execute executes a single result query and assigns to target
func Execute[T any](execFn func() (*T, error), target *T) error {
	res, err := execFn()
	if err != nil {
		return err
	}
	if res != nil {
		*target = *res
	}
	return nil
}

// ExecuteSlice executes a slice result query and assigns to target
func ExecuteSlice[T any](execFn func() ([]*T, error), target *[]*T) error {
	res, err := execFn()
	if err != nil {
		return err
	}
	*target = res
	return nil
}

// ExecutePage executes a paginated query and assigns results
func ExecutePage[T any](target *[]*T, count *int64, page *pagex.PagePrams, execFn func() ([]*T, int64, error)) error {
	res, total, err := execFn()
	if err != nil {
		return err
	}
	*target = res
	*count = total
	return nil
}
