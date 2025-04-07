package linear

import (
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPush(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		key      string
		value    string
		expected error
	}{
		{"1", "a", nil},
		{"2", "b", nil},
		{"3", "", nil},
		{"4", "c", nil},
		{"5", "d", nil},
	}

	linearClient := New(1024, false)

	for _, test := range tests {
		assert.Equal(linearClient.Push(test.key, test.value), test.expected)
	}

	assert.Equal(linearClient.GetNumberOfKeys(), 5)
}

func TestUpdateErrors(t *testing.T) {
	assert := assert.New(t)
	linearClient := New(1024, false)

	// Setup: thêm một key hợp lệ
	linearClient.Push("valid", "value")

	// Update với key rỗng
	err := linearClient.Update("", "new_value")
	assert.Error(err)

	// Update với value nil
	err = linearClient.Update("valid", nil)
	assert.Error(err)

	// Update với key không tồn tại
	err = linearClient.Update("not_exist", "value")
	assert.Error(err)
}

func TestPop(t *testing.T) {
	assert := assert.New(t)

	// Setting up
	datas := []struct {
		key   string
		value string
	}{
		{"1", "a"},
		{"2", "b"},
		{"3", "c"},
	}

	linearClient := New(1024, false)

	for _, data := range datas {
		linearClient.Push(data.key, data.value)
	}

	// Testing
	value, err := linearClient.Pop()
	if err != nil {
		t.Errorf("Pop failed, expected %v, got %v", "c", err.Error())
	}

	assert.Equal(value, "c")

	assert.Equal(linearClient.GetNumberOfKeys(), 2)
}

func TestTake(t *testing.T) {
	assert := assert.New(t)

	// Setting up
	datas := []struct {
		key   string
		value string
	}{
		{"1", "a"},
		{"2", "b"},
		{"3", "c"},
	}

	linearClient := New(1024, false)

	for _, data := range datas {
		linearClient.Push(data.key, data.value)
	}

	// Testing
	value, err := linearClient.Take()
	if err != nil {
		t.Errorf("Take failed, expected %v, got %v", "a", err)
	}

	assert.Equal(value, "a")

	assert.Equal(linearClient.GetNumberOfKeys(), 2)
}

func TestGet(t *testing.T) {
	assert := assert.New(t)

	// Setting up
	datas := []struct {
		key   string
		value string
	}{
		{"1", "a"},
		{"2", "b"},
		{"3", "c"},
	}

	linearClient := New(1024, false)

	for _, data := range datas {
		linearClient.Push(data.key, data.value)
	}

	// Testing
	value, err := linearClient.Get("2")
	if err != nil {
		t.Errorf("Get failed, expected %v, got %v", "b", err)
	}

	assert.Equal(value, "b")

	assert.Equal(linearClient.GetNumberOfKeys(), 2)
}

func TestRead(t *testing.T) {
	assert := assert.New(t)

	// Setting up
	datas := []struct {
		key   string
		value string
	}{
		{"1", "a"},
		{"2", "b"},
		{"3", "c"},
	}

	linearClient := New(1024, false)

	for _, data := range datas {
		linearClient.Push(data.key, data.value)
	}

	// Testing
	value, err := linearClient.Read("1")
	if err != nil {
		t.Errorf("Read failed, expected %v, got %v", "a", err)
	}

	assert.Equal(value, "a")

	assert.Equal(linearClient.GetNumberOfKeys(), 3)
}

func TestUpdate(t *testing.T) {
	assert := assert.New(t)

	// Setting up
	datas := []struct {
		key   string
		value string
	}{
		{"1", "a"},
		{"2", "b"},
		{"2", "c"},
	}

	linearClient := New(1024, false)

	for _, data := range datas {
		linearClient.Push(data.key, data.value)
	}

	// Testing
	err := linearClient.Update("2", "b2")
	if err != nil {
		t.Errorf("Update failed, expected %v, got %v", "b2", err)
	}

	value, err := linearClient.Read("2")
	if err != nil {
		t.Errorf("Update failed, expected %v, got %v", "b2", err)
	}

	assert.Equal(value, "b2")

	assert.Equal(linearClient.GetNumberOfKeys(), 3)
}

func BenchmarkPush(b *testing.B) {

	linearClient := New(1000000, true)

	// Run the Push method b.N times
	for n := 0; n < b.N; n++ {
		linearClient.Push("1", "a")
	}
}

func BenchmarkRead(b *testing.B) {
	linearClient := New(1, false)
	linearClient.Push("1", "a")

	for n := 0; n < b.N; n++ {
		linearClient.Read("1")
	}
}

func TestConcurrentAccess(t *testing.T) {
	assert := assert.New(t)
	linearClient := New(1000000, true)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := string(rune('a' + i))
			assert.Nil(linearClient.Push(key, "value"))
			_, err := linearClient.Pop()
			if err != nil && err.Error() != "linear is empty" && err.Error() != "index out of range" {
				t.Errorf("unexpected error: %v", err)
			}
		}(i)
	}
	wg.Wait()
}

func TestFullCapacity(t *testing.T) {
	assert := assert.New(t)
	linearClient := New(100, true) // Small capacity

	for i := 0; i < 10; i++ {
		assert.Nil(linearClient.Push(string(rune('a'+i)), strings.Repeat("x", 10)))
	}

	// Next push should trigger Take() automatically
	assert.Nil(linearClient.Push("k", "value"))
	assert.Equal(9, linearClient.GetNumberOfKeys())
}

func TestComplexData(t *testing.T) {
	assert := assert.New(t)
	linearClient := New(1000000, false)

	type complexStruct struct {
		ID    int
		Name  string
		Items []string
	}

	val := complexStruct{
		ID:    1,
		Name:  "test",
		Items: []string{"a", "b", "c"},
	}

	assert.Nil(linearClient.Push("complex", val))
	result, err := linearClient.Get("complex")
	assert.Nil(err)
	assert.Equal(val, result)
}

func TestErrorCases(t *testing.T) {
	assert := assert.New(t)
	linearClient := New(100, false)

	// Empty key
	err := linearClient.Push("", "value")
	assert.Error(err)

	// Nil value
	err = linearClient.Push("key", nil)
	assert.Error(err)

	// Get from empty
	_, err = linearClient.Pop()
	assert.Error(err)
}
