package model

import (
	"testing"
)

// Тестовая структура для проверки пула
type TestStruct struct {
	Value   int
	Message string
	Data    *int
	Slice   []int
	Map     map[string]int
}

// Реализация метода Reset для TestStruct
func (t *TestStruct) Reset() {
	if t == nil {
		return
	}
	t.Value = 0
	t.Message = ""
	if t.Data != nil {
		*t.Data = 0
	}
	t.Slice = t.Slice[:0] // Очищаем слайс
	if t.Map != nil {
		clear(t.Map) // Очищаем мапу
	}
}

// TestPool_Get проверяет, что метод Get возвращает инициализированный объект
func TestPool_Get(t *testing.T) {
	pool := New(func() *TestStruct { return &TestStruct{} })

	// Получаем объект из пула
	obj := pool.Get()
	if obj == nil {
		t.Fatal("Get returned nil")
	}

	// Проверяем, что поля в нулевых значениях после Reset (вызывается при Put)
	// или при инициализации (в New)
	if obj.Value != 0 {
		t.Errorf("Expected Value to be 0, got %d", obj.Value)
	}
	if obj.Message != "" {
		t.Errorf("Expected Message to be empty, got %s", obj.Message)
	}
	if obj.Data != nil {
		t.Errorf("Expected Data to be nil, got %v", obj.Data)
	}
	if len(obj.Slice) != 0 {
		t.Errorf("Expected Slice to be empty, got %v", obj.Slice)
	}
	if obj.Map != nil {
		t.Errorf("Expected Map to be nil, got %v", obj.Map)
	}
}

// TestPool_Put проверяет, что метод Put корректно сбрасывает состояние и возвращает объект в пул
func TestPool_Put(t *testing.T) {
	pool := New(func() *TestStruct { return &TestStruct{} })

	// Получаем объект
	obj := pool.Get()
	if obj == nil {
		t.Fatal("Get returned nil")
	}

	// Проверяем начальное состояние (нулевые значения)
	if obj.Value != 0 || obj.Message != "" || obj.Data != nil || len(obj.Slice) != 0 || obj.Map != nil {
		t.Errorf("Get returned non-zero object: %+v", obj)
	}

	// Модифицируем объект
	val := 42
	obj.Value = 100
	obj.Message = "test"
	obj.Data = &val
	obj.Slice = append(obj.Slice, 1, 2, 3)
	obj.Map = map[string]int{"key": 42}

	// Проверяем, что изменения применились
	if obj.Value != 100 {
		t.Errorf("Expected Value to be 100, got %d", obj.Value)
	}
	if obj.Message != "test" {
		t.Errorf("Expected Message to be 'test', got %s", obj.Message)
	}
	if obj.Data == nil || *obj.Data != 42 {
		t.Errorf("Expected Data to point to 42, got %v", obj.Data)
	}
	if len(obj.Slice) != 3 {
		t.Errorf("Expected Slice to have 3 elements, got %d", len(obj.Slice))
	}
	if obj.Map == nil || obj.Map["key"] != 42 {
		t.Errorf("Expected Map to contain key=42, got %v", obj.Map)
	}

	// Возвращаем объект в пул
	pool.Put(obj)

	// Получаем объект снова
	obj2 := pool.Get()
	if obj2 == nil {
		t.Fatal("Get returned nil after Put")
	}

	// Проверяем, что объект был сброшен до нулевых значений
	if obj2.Value != 0 {
		t.Errorf("Expected Value to be 0 after Put, got %d", obj2.Value)
	}
	if obj2.Message != "" {
		t.Errorf("Expected Message to be empty after Put, got %s", obj2.Message)
	}
	if len(obj2.Slice) != 0 {
		t.Errorf("Expected Slice to be empty after Put, got %v", obj2.Slice)
	}
	if obj2.Data != nil && *obj.Data != 0 {
		t.Errorf("Expected Data to be nil or 0 after Put, got %v", obj2.Data)
	}
}

// TestPool_Concurrent проверяет безопасность пула при многопоточном доступе
func TestPool_Concurrent(t *testing.T) {
	pool := New(func() *TestStruct { return &TestStruct{} })
	const numGoroutines = 1000
	done := make(chan bool)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			obj := pool.Get()
			obj.Value++
			obj.Message = "updated"
			pool.Put(obj)
			done <- true
		}()
	}

	// Ожидаем завершения всех горутин
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

// TestPool_Metrics проверяет работу пула с объектами Metrics
func TestPool_Metrics(t *testing.T) {
	pool := New(func() *Metrics { return &Metrics{} })

	// Получаем объект Metrics из пула
	obj := pool.Get()
	if obj == nil {
		t.Fatal("Get returned nil")
	}

	// Проверяем начальное состояние (нулевые значения)
	if obj.ID != "" || obj.MType != "" || obj.Delta != nil || obj.Value != nil || obj.Hash != "" {
		t.Errorf("Get returned non-zero Metrics object: %+v", obj)
	}

	// Модифицируем объект
	value := int64(42)
	obj.ID = "test_counter"
	obj.MType = Counter
	obj.Delta = &value
	obj.Hash = "test_hash"

	// Проверяем, что изменения применились
	if obj.ID != "test_counter" {
		t.Errorf("Expected ID to be 'test_counter', got %s", obj.ID)
	}
	if obj.MType != Counter {
		t.Errorf("Expected MType to be 'counter', got %s", obj.MType)
	}
	if obj.Delta == nil || *obj.Delta != 42 {
		t.Errorf("Expected Delta to point to 42, got %v", obj.Delta)
	}
	if obj.Hash != "test_hash" {
		t.Errorf("Expected Hash to be 'test_hash', got %s", obj.Hash)
	}

	// Возвращаем объект в пул
	pool.Put(obj)

	// Получаем объект снова
	obj2 := pool.Get()
	if obj2 == nil {
		t.Fatal("Get returned nil after Put")
	}

	// Проверяем, что объект был сброшен до нулевых значений
	if obj2.ID != "" {
		t.Errorf("Expected ID to be empty after Put, got %s", obj2.ID)
	}
	if obj2.MType != "" {
		t.Errorf("Expected MType to be empty after Put, got %s", obj2.MType)
	}
	if obj2.Delta != nil && *obj2.Delta != 0 {
		t.Errorf("Expected Delta value to be 0, got %v", *obj2.Delta)
	}
	if obj2.Value != nil {
		t.Errorf("Expected Value to be nil after Put, got %v", obj2.Value)
	} else if obj2.Value != nil && *obj2.Value != 0.0 {
		t.Errorf("Expected Value value to be 0.0, got %v", *obj2.Value)
	}
	if obj2.Hash != "" {
		t.Errorf("Expected Hash to be empty after Put, got %s", obj2.Hash)
	}
}
