package rank

import "github.com/markphelps/optional"

type RankElem[T any] struct {
	sol T
	err error
}

func (re *RankElem[T]) IsError() bool {
	return re.err != nil
}

type Rank[T any] struct {
	ranking      map[int][]*RankElem[T]
	has_solution bool

	max_rank optional.Int
}

/*
func (r *Rank[T]) Iterator() uc.Iterater[T] {
	buckets := make(map[int][]T)

	for i := 0; i < len(r.solutions); i++ {
		prev, ok := buckets[r.rank[i]]
		if ok {
			buckets[r.rank[i]] = append(prev, r.solutions[i])
		} else {
			buckets[r.rank[i]] = []T{r.solutions[i]}
		}
	}

	var keys []int

	for k := range buckets {
		pos, ok := slices.BinarySearch(keys, k)
		uc.AssertOk(ok, "slices.BinarySearch(%d, %v)", keys, k)

		keys = slices.Insert(keys, pos, k)
	}

	return uc.NewDynamicIterator(
		uc.NewSimpleIterator(keys),
		func(rank int) uc.Iterater[T] {
			values, ok := buckets[rank]
			uc.AssertF(!ok, "rank %d not found", rank)

			return uc.NewSimpleIterator(values)
		},
	)
}
*/

func NewRank[T comparable]() *Rank[T] {
	return &Rank[T]{
		ranking:      make(map[int][]*RankElem[T]),
		has_solution: false,
		max_rank:     optional.Int{},
	}
}

func (r *Rank[T]) trim_err_elems() {
	for rank, values := range r.ranking {
		var top int

		for i := 0; i < len(values); i++ {
			elem := values[i]

			if !elem.IsError() {
				values[top] = values[i]
				top++
			}
		}

		r.ranking[rank] = values[:top]
	}

	for rank, values := range r.ranking {
		if len(values) == 0 {
			delete(r.ranking, rank)
		}
	}
}

func (r *Rank[T]) AddSol(sol T, rank int) {
	if !r.has_solution {
		r.has_solution = true

		r.trim_err_elems()
	}

	elem := &RankElem[T]{
		sol: sol,
		err: nil,
	}

	prev, ok := r.ranking[rank]
	if ok {
		r.ranking[rank] = append(prev, elem)
	} else {
		r.ranking[rank] = []*RankElem[T]{elem}
	}
}

func (r *Rank[T]) AddErr(err error, rank int) {
	if r.has_solution {
		return
	}

	elem := &RankElem[T]{
		sol: *new(T),
		err: err,
	}

	prev, ok := r.ranking[rank]
	if ok {
		r.ranking[rank] = append(prev, elem)
	} else {
		r.ranking[rank] = []*RankElem[T]{elem}
	}
}

func (r *Rank[T]) HasSolution() bool {
	return r.has_solution
}

func (r *Rank[T]) GetErrors() []error {
	max_rank := 0

	for rank, values := range r.ranking {
		var errs []error

		for _, value := range values {
			if value.IsError() {
				errs = append(errs, value.err)
			}
		}

		if len(errs) == 0 {
			continue
		}

	}
}
