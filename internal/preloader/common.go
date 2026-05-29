package preloader

type Job[T any] struct {
	el     *T // addr of element that we work with
	offset int
}
