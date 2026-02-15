package model

import "sync"

// Pool представляет собой пул объектов с возможностью переиспользования.
// T — тип объектов в пуле, которые должны иметь метод Reset().
// Обычно T — это *SomeStruct, где SomeStruct реализует Reset().
type Pool[T interface{ Reset() }] struct {
	inner sync.Pool
}

// New создаёт и возвращает указатель на новый пул объектов типа T.
// Конструктор newFunc должен возвращать новый экземпляр T (например, &MyStruct{}).
func New[T interface{ Reset() }](newFunc func() T) *Pool[T] {
	return &Pool[T]{
		inner: sync.Pool{
			New: func() interface{} {
				return newFunc()
			},
		},
	}
}

// Get возвращает объект из пула.
// Если пул пуст, создаётся новый объект с помощью функции, переданной в New.
func (p *Pool[T]) Get() T {
	return p.inner.Get().(T)
}

// Put помещает объект обратно в пул, предварительно сбрасывая его состояние.
func (p *Pool[T]) Put(obj T) {
	obj.Reset()
	p.inner.Put(obj)
}
