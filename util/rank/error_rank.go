package rank

import (
	"errors"
	"slices"
)

type ErrorRank struct {
	elems []error
	ranks []int

	max_rank *Max
}

func NewErrorRank() *ErrorRank {
	return &ErrorRank{
		elems:    make([]error, 0),
		ranks:    make([]int, 0),
		max_rank: NewMax(),
	}
}

func (r *ErrorRank) AddError(err error, rank int) {
	idx := slices.IndexFunc(r.elems, func(e error) bool {
		return errors.Is(e, err)
	})

	// INFO: This function ignores rank changing when it is lower than an existing rank.
	// So, if I ever need to change this behavior, modify this function.

	if idx != -1 {
		if r.ranks[idx] < rank {
			r.ranks[idx] = rank
		}
	} else {
		r.elems = append(r.elems, err)
		r.ranks = append(r.ranks, rank)
	}

	if idx == -1 || r.ranks[idx] < rank {
		prev_rank, ok := r.max_rank.Get()

		if !ok || rank > prev_rank {
			r.max_rank.Set(rank)
		}
	}

}

func (r *ErrorRank) MaxRank() (int, bool) {
	return r.max_rank.Get()
}
