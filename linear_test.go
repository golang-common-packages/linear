package linear

import (
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

	// Setting up
	linearClient := New(1000000, true)
	New(1024, false).Push("1", "a")

	// Run the Read method b.N times
	for n := 0; n < b.N; n++ {
		linearClient.Read("1")
	}
}
