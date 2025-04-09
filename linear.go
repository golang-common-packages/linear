package linear

import (
	"errors"
	"log"
	"reflect"
	"sync"
)

type Linear struct {
	items       *sync.Map
	keys        []string
	sizeChecker bool
	linearSizes int64

	// Separate locks for different resources
	sizeMux           sync.RWMutex
	keysMux           sync.RWMutex
	linearCurrentSize int64
}

func (l *Linear) calculateItemSize(key string, value interface{}) int64 {
	size := int64(len(key))

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		size += int64(v.Len())
	case reflect.Slice, reflect.Array:
		size += int64(v.Len()) * int64(v.Type().Elem().Size())
	case reflect.Map, reflect.Struct:
		size += 64 // byte
	default:
		size += int64(reflect.TypeOf(value).Size())
	}
	return size
}

func New(maxSize int64, sizeChecker bool) *Linear {
	if maxSize <= 0 {
		log.Fatalln("linearSizes must be higher than 0")
	}

	return &Linear{
		keys:              make([]string, 0, 16), // Pre-allocate initial capacity
		items:             &sync.Map{},
		sizeChecker:       sizeChecker,
		linearSizes:       maxSize,
		linearCurrentSize: 0,
	}
}

func (l *Linear) Push(key string, value interface{}) error {
	if key == "" || value == nil {
		return errors.New("key and value should not be empty")
	}

	// Check if the key already exists
	oldSize := int64(0)
	oldValue, exists := l.items.Load(key)
	if exists {
		oldSize = l.calculateItemSize(key, oldValue)
	}

	// Only calculate the new size when necessary
	itemSize := l.calculateItemSize(key, value)
	if itemSize > l.linearSizes {
		return errors.New("linear doesn't have enough memory space")
	}

	// Calculate the actual size change needed
	actualSizeChange := itemSize - oldSize

	if l.sizeChecker && actualSizeChange > 0 {
		// Only remove old elements if more space is needed
		for l.linearCurrentSize+actualSizeChange > l.linearSizes {
			if _, err := l.Take(); err != nil {
				// If no more elements can be removed
				return err
			}
		}
	}

	l.items.Store(key, value)
	l.sizeMux.Lock()
	l.linearCurrentSize = l.linearCurrentSize - oldSize + itemSize
	l.sizeMux.Unlock()

	// Only add the key to the keys array if it doesn't already exist
	if !exists {
		l.keysMux.Lock()
		l.keys = append(l.keys, key)
		l.keysMux.Unlock()
	}

	return nil
}

func (l *Linear) removeByIndex(index int) (interface{}, int64, error) {
	// Check before locking to avoid deadlock
	if l.IsEmpty() {
		return nil, 0, errors.New("linear is empty")
	}

	l.keysMux.Lock()
	// Check again after locking as it might have changed
	if len(l.keys) == 0 {
		l.keysMux.Unlock()
		return nil, 0, errors.New("linear is empty")
	}

	if index < 0 || index >= len(l.keys) {
		l.keysMux.Unlock()
		return nil, 0, errors.New("index out of range")
	}

	key := l.keys[index]
	l.keys = removeItemByIndex(l.keys, index)
	l.keysMux.Unlock()

	item, ok := l.items.Load(key)
	if !ok {
		return nil, 0, errors.New("key exists in keys array but not in items map")
	}

	itemSize := l.calculateItemSize(key, item)
	l.items.Delete(key)

	l.sizeMux.Lock()
	l.linearCurrentSize -= itemSize
	l.sizeMux.Unlock()

	return item, itemSize, nil
}

func (l *Linear) Pop() (interface{}, error) {
	item, _, err := l.removeByIndex(len(l.keys) - 1)
	return item, err
}

func (l *Linear) Take() (interface{}, error) {
	item, _, err := l.removeByIndex(0)
	return item, err
}

func (l *Linear) Get(key string) (interface{}, error) {
	if l.IsEmpty() {
		return nil, errors.New("linear is empty")
	}

	index, exists := findIndexByItem(key, l.keys)
	if !exists {
		return nil, errors.New("key does not exist")
	}

	item, _, err := l.removeByIndex(index)
	return item, err
}

func (l *Linear) Read(key string) (interface{}, error) {
	if l.IsEmpty() {
		return nil, errors.New("linear is empty")
	}

	item, ok := l.items.Load(key)
	if !ok {
		return nil, errors.New("key does not exist")
	}

	return item, nil
}

func (l *Linear) Update(key string, value interface{}) error {
	if key == "" || value == nil {
		return errors.New("key and value should not be empty")
	}

	if l.IsEmpty() {
		return errors.New("linear is empty")
	}

	newItemSize := l.calculateItemSize(key, value)
	if newItemSize > l.linearSizes {
		return errors.New("not enough space for the new value")
	}

	currentSize, exists := l.IsExists(key)
	if !exists {
		return errors.New("key does not exist")
	}

	l.items.Store(key, value)
	l.sizeMux.Lock()
	l.linearCurrentSize -= currentSize
	l.linearCurrentSize += newItemSize
	l.sizeMux.Unlock()

	return nil
}

func (l *Linear) Range(fn func(key, value interface{}) bool) {
	l.items.Range(fn)
}

func (l *Linear) IsExists(key string) (int64, bool) {
	value, exists := l.items.Load(key)
	if !exists {
		return 0, false
	}
	return l.calculateItemSize(key, value), true
}

func (l *Linear) IsEmpty() bool {
	l.keysMux.RLock()
	isEmpty := len(l.keys) == 0
	l.keysMux.RUnlock()
	return isEmpty
}

func (l *Linear) GetItems() *sync.Map {
	return l.items
}

func (l *Linear) Getkeys() []string {
	return l.keys
}

func (l *Linear) GetNumberOfKeys() int {
	return len(l.keys)
}

func (l *Linear) GetLinearSizes() int64 {
	return l.linearSizes
}

func (l *Linear) SetLinearSizes(linearSizes int64) error {
	if linearSizes <= 0 {
		return errors.New("linearSizes must be higher than 0")
	}

	l.sizeMux.Lock()
	l.linearSizes = linearSizes
	l.sizeMux.Unlock()

	return nil
}

func (l *Linear) GetLinearCurrentSize() int64 {
	l.sizeMux.RLock()
	currentSize := l.linearCurrentSize
	l.sizeMux.RUnlock()
	return currentSize
}
