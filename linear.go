package linear

import (
	"errors"
	"log"
	"sync"
	"unsafe"
)

// Linear contains all the private properties
type Linear struct {
	items             *sync.Map
	keys              []string
	sizeChecker       bool
	linearSizes       int64
	linearCurrentSize int64
	rwMutex           *sync.RWMutex
}

// New return new linear instance
func New(maxSize int64, sizeChecker bool) *Linear {

	// Argument validator
	if maxSize <= 0 {
		log.Fatalln("linearSizes much higher than 0")
	}

	currentLinear := Linear{
		keys:              []string{},
		items:             &sync.Map{},
		sizeChecker:       sizeChecker,
		linearSizes:       maxSize,
		linearCurrentSize: 0,
		rwMutex:           &sync.RWMutex{},
	}

	return &currentLinear
}

// Push item to the linear with key
func (l *Linear) Push(key string, value interface{}) error {

	// Argument validator
	if key == "" && value == nil {
		return errors.New("Key and value should not be empty")
	}

	itemSize := int64(unsafe.Sizeof(key)) + int64(unsafe.Sizeof(value))
	if itemSize > l.linearSizes {
		return errors.New("Linear doesn't have enough memory space")
	}

	// Clean space for new item
	if l.sizeChecker {
		for l.linearCurrentSize+itemSize > l.linearSizes {
			if _, err := l.Take(); err != nil {
				return err
			}
		}
	}

	l.rwMutex.Lock()
	l.items.LoadOrStore(key, value)
	l.linearCurrentSize += itemSize
	l.keys = append(l.keys, key)
	l.rwMutex.Unlock()

	return nil
}

// Pop return the last item from the linear and remove it out of Linear
func (l *Linear) Pop() (interface{}, error) {

	// Execution conditions
	if l.IsEmpty() {
		return nil, errors.New("Linear is empty")
	}

	lastItemIndex := len(l.keys) - 1
	item, ok := l.items.Load(l.keys[lastItemIndex])
	if !ok {
		return nil, nil
	}

	itemSize := int64(unsafe.Sizeof(item)) + int64(unsafe.Sizeof(l.keys[lastItemIndex]))

	l.rwMutex.Lock()
	l.items.Delete(l.keys[lastItemIndex])
	l.linearCurrentSize -= itemSize
	l.keys = removeItemByIndex(l.keys, lastItemIndex) //Update keys slice after remove that key/value from items map
	l.rwMutex.Unlock()

	return item, nil
}

// Take return the first item from the linear and remove it
func (l *Linear) Take() (interface{}, error) {

	// Execution conditions
	if l.IsEmpty() {
		return nil, errors.New("Can't Take, because Linear is empty")
	}

	item, ok := l.items.Load(l.keys[0])
	if !ok {
		return nil, nil
	}

	itemSize := int64(unsafe.Sizeof(item)) + int64(unsafe.Sizeof(l.keys[0]))

	l.rwMutex.Lock()
	l.items.Delete(l.keys[0])
	l.linearCurrentSize -= itemSize
	l.keys = removeItemByIndex(l.keys, 0) //Update keys slice after remove that key/value from items map
	l.rwMutex.Unlock()

	return item, nil
}

// Get method return the item by key from linear and remove it
func (l *Linear) Get(key string) (interface{}, error) {

	// Execution conditions
	if l.IsEmpty() {
		return nil, errors.New("Linear is empty")
	}

	var (
		wg             sync.WaitGroup
		item           interface{}
		itemExits      bool
		itemIndex      int
		itemIndexExits bool
	)

	wg.Add(2)
	go func() {
		item, itemExits = l.items.Load(key)
		wg.Done()
	}()

	go func() {
		itemIndex, itemIndexExits = findIndexByItem(key, l.keys)
		wg.Done()
	}()
	wg.Wait()

	if !itemExits || !itemIndexExits {
		return nil, nil
	}

	l.rwMutex.Lock()
	l.items.Delete(key)
	l.linearCurrentSize -= int64(unsafe.Sizeof(item))
	l.keys = removeItemByIndex(l.keys, itemIndex) //Update keys slice after remove that key from items map
	l.rwMutex.Unlock()

	return item, nil
}

// Read method return the item by key from linear without remove it
func (l *Linear) Read(key string) (interface{}, error) {

	// Execution conditions
	if l.IsEmpty() {
		return nil, errors.New("Linear is empty")
	}

	item, exits := l.items.Load(key)
	if !exits {
		return nil, nil
	}

	return item, nil
}

// Update reassign value to the key
func (l *Linear) Update(key string, value interface{}) error {

	// Argument validator
	if key == "" && value == nil {
		return errors.New("key and value should not be empty")
	}

	// Execution conditions
	if l.IsEmpty() {
		return errors.New("Linear is empty")
	}

	newItemSize := int64(unsafe.Sizeof(key)) + int64(unsafe.Sizeof(value))
	if newItemSize > l.linearSizes || l.IsEmpty() {
		return errors.New("Linear is empty or not enough space")
	}

	l.rwMutex.Lock()
	currentSize, exits := l.IsExits(key)
	if !exits {
		l.rwMutex.Unlock()
		return errors.New("Key does not exit")
	}
	l.items.Store(key, value)
	l.linearCurrentSize += currentSize - newItemSize
	l.rwMutex.Unlock()

	return nil
}

// Range the LinearClient
func (l *Linear) Range(fn func(key, value interface{}) bool) {
	l.items.Range(fn)
}

// IsExits check key exits or not
func (l *Linear) IsExits(key string) (int64, bool) {

	value, exits := l.items.Load(key)
	if !exits {
		return 0, false
	}

	return int64(unsafe.Sizeof(key)) + int64(unsafe.Sizeof(value)), true
}

// IsEmpty check linear size
func (l *Linear) IsEmpty() bool {
	return len(l.keys) == 0
}

// GetItems return the map contain items
func (l *Linear) GetItems() *sync.Map {
	return l.items
}

// Getkeys return the list of key
func (l *Linear) Getkeys() []string {
	return l.keys
}

// GetNumberOfKeys return the number of keys
func (l *Linear) GetNumberOfKeys() int {
	return len(l.keys)
}

// GetLinearSizes return the linear size
func (l *Linear) GetLinearSizes() int64 {
	return l.linearSizes
}

// SetLinearSizes change the linear size with new value
func (l *Linear) SetLinearSizes(linearSizes int64) error {

	// Argument validator
	if linearSizes <= 0 {
		return errors.New("linearSizes much higher than 0")
	}

	l.rwMutex.RLock()
	l.linearSizes = linearSizes
	l.rwMutex.RUnlock()

	return nil
}

// GetLinearCurrentSize return the current linear size
func (l *Linear) GetLinearCurrentSize() int64 {
	return l.linearCurrentSize
}

// removeItemByIndex remove item out of []string by index but maintains order, and return the new one
// Source: https://yourbasic.org/golang/delete-element-slice/
func removeItemByIndex(slice []string, idx int) []string {

	copy(slice[idx:], slice[idx+1:]) // Shift slice[idx+1:] left one index.
	slice[len(slice)-1] = ""         // Erase last element (write zero value).
	return slice[:len(slice)-1]      // Truncate slice.
}

// findIndexByItem return index belong to the key
// Source: https://stackoverflow.com/questions/46745043/performance-of-for-range-in-go
func findIndexByItem(keyName string, items []string) (int, bool) {

	for index := range items {
		if keyName == items[index] {
			return index, true
		}
	}

	return -1, false
}
