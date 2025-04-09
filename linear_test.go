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

	// Setup: add a valid key
	linearClient.Push("valid", "value")

	// Update with empty key
	err := linearClient.Update("", "new_value")
	assert.Error(err)

	// Update with nil value
	err = linearClient.Update("valid", nil)
	assert.Error(err)

	// Update with non-existent key
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

	assert.Equal(linearClient.GetNumberOfKeys(), 2)
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

func TestPushDuplicateKey(t *testing.T) {
	assert := assert.New(t)
	linearClient := New(1024, false)

	// Push a key-value pair
	assert.Nil(linearClient.Push("key1", "value1"))
	assert.Equal(1, linearClient.GetNumberOfKeys())

	// Push the same key with a different value
	assert.Nil(linearClient.Push("key1", "value2"))

	// Verify that the number of keys hasn't changed
	assert.Equal(1, linearClient.GetNumberOfKeys())

	// Verify that the value has been updated
	value, err := linearClient.Read("key1")
	assert.Nil(err)
	assert.Equal("value2", value)
}

func TestIsEmptyAndSize(t *testing.T) {
	assert := assert.New(t)
	linearClient := New(1024, false)

	// Initially empty
	assert.True(linearClient.IsEmpty())
	assert.Equal(int64(0), linearClient.GetLinearCurrentSize())

	// Add an item
	assert.Nil(linearClient.Push("key1", "value1"))
	assert.False(linearClient.IsEmpty())

	// Check size increased
	initialSize := linearClient.GetLinearCurrentSize()
	assert.Greater(initialSize, int64(0))

	// Add another item
	assert.Nil(linearClient.Push("key2", "value2"))
	assert.Greater(linearClient.GetLinearCurrentSize(), initialSize)

	// Remove items
	_, err := linearClient.Pop()
	assert.Nil(err)
	_, err = linearClient.Take()
	assert.Nil(err)

	// Should be empty again
	assert.True(linearClient.IsEmpty())
	assert.Equal(int64(0), linearClient.GetLinearCurrentSize())
}

func TestSetLinearSizes(t *testing.T) {
	assert := assert.New(t)
	linearClient := New(100, false)

	// Initial size
	assert.Equal(int64(100), linearClient.GetLinearSizes())

	// Set to a valid size
	err := linearClient.SetLinearSizes(200)
	assert.Nil(err)
	assert.Equal(int64(200), linearClient.GetLinearSizes())

	// Try to set to an invalid size
	err = linearClient.SetLinearSizes(0)
	assert.Error(err)
	assert.Equal("linearSizes must be higher than 0", err.Error())

	// Size should remain unchanged
	assert.Equal(int64(200), linearClient.GetLinearSizes())

	// Try to set to a negative size
	err = linearClient.SetLinearSizes(-100)
	assert.Error(err)
	assert.Equal("linearSizes must be higher than 0", err.Error())

	// Size should still remain unchanged
	assert.Equal(int64(200), linearClient.GetLinearSizes())
}

func TestRangeAndGetters(t *testing.T) {
	assert := assert.New(t)
	linearClient := New(1024, false)

	// Add some items
	assert.Nil(linearClient.Push("key1", "value1"))
	assert.Nil(linearClient.Push("key2", "value2"))
	assert.Nil(linearClient.Push("key3", "value3"))

	// Test GetItems
	items := linearClient.GetItems()
	assert.NotNil(items)

	// Test Getkeys
	keys := linearClient.Getkeys()
	assert.Equal(3, len(keys))
	assert.Contains(keys, "key1")
	assert.Contains(keys, "key2")
	assert.Contains(keys, "key3")

	// Test Range
	count := 0
	values := make(map[string]string)
	linearClient.Range(func(key, value interface{}) bool {
		count++
		values[key.(string)] = value.(string)
		return true
	})

	assert.Equal(3, count)
	assert.Equal("value1", values["key1"])
	assert.Equal("value2", values["key2"])
	assert.Equal("value3", values["key3"])

	// Test Range with early termination
	count = 0
	linearClient.Range(func(key, value interface{}) bool {
		count++
		return false // Stop after first item
	})

	assert.Equal(1, count)
}

func TestIsExists(t *testing.T) {
	assert := assert.New(t)
	linearClient := New(1024, false)

	// Check non-existent key
	size, exists := linearClient.IsExists("nonexistent")
	assert.Equal(int64(0), size)
	assert.False(exists)

	// Add an item
	assert.Nil(linearClient.Push("key1", "value1"))

	// Check existing key
	size, exists = linearClient.IsExists("key1")
	assert.True(exists)
	assert.Greater(size, int64(0))

	// Add a larger item
	assert.Nil(linearClient.Push("key2", strings.Repeat("x", 100)))

	// Check that size is proportional to content
	size1, _ := linearClient.IsExists("key1")
	size2, _ := linearClient.IsExists("key2")
	assert.Greater(size2, size1)

	// Remove an item
	_, err := linearClient.Get("key1")
	assert.Nil(err)

	// Check that removed key no longer exists
	_, exists = linearClient.IsExists("key1")
	assert.False(exists)
}
