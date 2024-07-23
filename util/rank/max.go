package rank

type Max struct {
	value  int
	is_set bool
}

func NewMax() *Max {
	return &Max{
		value:  0,
		is_set: false,
	}
}

func (m *Max) Set(value int) {
	m.value = value
	m.is_set = true
}

func (m *Max) Get() (int, bool) {
	return m.value, m.is_set
}
