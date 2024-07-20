package pkg

type DataType struct {
	// Name is the name of the data type.
	Name string

	// Sig is the type signature of the data type.
	Sig string

	// Generics is the full generics signature of the data type.
	Generics string

	// ZeroValue is the zero value of the data type.
	ZeroValue string
}

const templ = `

// {{ .Name }} is a stack of {{ .DataType }} values implemented without a maximum capacity
// and using a linked list.
type {{ .Name }}{{ .Generics }} struct {
	front *{{ .HelperSig }}
	size int
}

// New{{ .Name }} creates a new linked stack.
//
// Returns:
//   - *{{ .TypeSig }}: A pointer to the newly created stack. Never returns nil.
func New{{ .Name }}{{ .Generics }}() *{{ .TypeSig }} {
	return &{{ .TypeSig }}{
		size: 0,
	}
}

// Push implements the stack.Stacker interface.
//
// Always returns true.
func (s *{{ .TypeSig }}) Push(value {{ .DataType }}) bool {
	node := &{{ .HelperSig }}{
		value: value,
	}

	if s.front != nil {
		node.next = s.front
	}

	s.front = node
	s.size++

	return true
}

// PushMany implements the stack.Stacker interface.
//
// Always returns the number of values pushed onto the stack.
func (s *{{ .TypeSig }}) PushMany(values []{{ .DataType }}) int {
	if len(values) == 0 {
		return 0
	}

	node := &{{ .HelperSig }}{
		value: values[0],
	}

	if s.front != nil {
		node.next = s.front
	}

	s.front = node

	for i := 1; i < len(values); i++ {
		node := &{{ .HelperSig }}{
			value: values[i],
			next:  s.front,
		}

		s.front = node
	}

	s.size += len(values)
	
	return len(values)
}

// Pop implements the stack.Stacker interface.
func (s *{{ .TypeSig }}) Pop() ({{ .DataType }}, bool) {
	if s.front == nil {
		return {{ .ZeroValue }}, false
	}

	to_remove := s.front
	s.front = s.front.next

	s.size--
	to_remove.next = nil

	return to_remove.value, true
}

// Peek implements the stack.Stacker interface.
func (s *{{ .TypeSig }}) Peek() ({{ .DataType }}, bool) {
	if s.front == nil {
		return {{ .ZeroValue }}, false
	}

	return s.front.value, true
}

// IsEmpty implements the stack.Stacker interface.
func (s *{{ .TypeSig }}) IsEmpty() bool {
	return s.front == nil
}

// Size implements the stack.Stacker interface.
func (s *{{ .TypeSig }}) Size() int {
	return s.size
}

// Iterator implements the stack.Stacker interface.
func (s *{{ .TypeSig }}) Iterator() common.Iterater[{{ .DataType }}] {
	var builder common.Builder[{{ .DataType }}]

	for node := s.front; node != nil; node = node.next {
		builder.Add(node.value)
	}

	return builder.Build()
}

// Clear implements the stack.Stacker interface.
func (s *{{ .TypeSig }}) Clear() {
	if s.front == nil {
		return
	}

	prev := s.front

	for node := s.front.next; node != nil; node = node.next {
		prev = node
		prev.next = nil
	}

	prev.next = nil

	s.front = nil
	s.size = 0
}

// GoString implements the stack.Stacker interface.
func (s *{{ .TypeSig }}) GoString() string {
	values := make([]string, 0, s.size)
	for node := s.front; node != nil; node = node.next {
		values = append(values, common.StringOf(node.value))
	}

	var builder strings.Builder

	builder.WriteString("{{ .TypeSig }}[size=")
	builder.WriteString(strconv.Itoa(s.size))
	builder.WriteString(", values=[")
	builder.WriteString(strings.Join(values, ", "))
	builder.WriteString(" →]]")

	return builder.String()
}

// Slice implements the stack.Stacker interface.
//
// The 0th element is the top of the stack.
func (s *{{ .TypeSig }}) Slice() []{{ .DataType }} {
	slice := make([]{{ .DataType }}, 0, s.size)

	for node := s.front; node != nil; node = node.next {
		slice = append(slice, node.value)
	}

	return slice
}

// Copy implements the stack.Stacker interface.
//
// The copy is a shallow copy.
func (s *{{ .TypeSig }}) Copy() common.Copier {
	if s.front == nil {
		return &{{ .TypeSig }}{}
	}

	s_copy := &{{ .TypeSig }}{
		size: s.size,
	}

	node_copy := &{{ .HelperSig }}{
		value: s.front.value,
	}

	s_copy.front = node_copy

	prev := node_copy

	for node := s.front.next; node != nil; node = node.next {
		node_copy := &{{ .HelperSig }}{
			value: node.value,
		}

		prev.next = node_copy

		prev = node_copy
	}

	return s_copy
}

// Capacity implements the stack.Stacker interface.
//
// Always returns -1.
func (s *{{ .TypeSig }}) Capacity() int {
	return -1
}

// IsFull implements the stack.Stacker interface.
//
// Always returns false.
func (s *{{ .TypeSig }}) IsFull() bool {
	return false
}
`
