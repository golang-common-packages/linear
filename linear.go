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
	linearSizes       int64 // bytes
	linearCurrentSize int64 // bytes
	mux               *sync.RWMutex
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
		mux:               &sync.RWMutex{},
	}

	return &currentLinear
}

// Push item to the linear with key
func (l *Linear) Push(key string, value interface{}) error {

	// Argument validator
	if key == "" && value == nil {
		return errors.New("key and value should not be empty")
	}

	itemSize := int64(unsafe.Sizeof(key)) + int64(unsafe.Sizeof(value))
	if itemSize > l.linearSizes {
		return errors.New("linear doesn't have enough memory space")
	}

	// Clean space for new item
	if l.sizeChecker {
		for l.linearCurrentSize+itemSize > l.linearSizes {
			if _, err := l.Take(); err != nil {
				return err
			}
		}
	}

	l.items.LoadOrStore(key, value)
	l.mux.Lock()
	l.linearCurrentSize += itemSize
	l.keys = append(l.keys, key)
	l.mux.Unlock()

	return nil
}

// Pop return and remove the last item out of the linear
func (l *Linear) Pop() (interface{}, error) {

	// Execution conditions
	if l.IsEmpty() {
		return nil, errors.New("linear is empty")
	}

	lastItemIndex := len(l.keys) - 1
	item, ok := l.items.Load(l.keys[lastItemIndex])
	if !ok {
		return nil, nil
	}

	itemSize := int64(unsafe.Sizeof(item)) + int64(unsafe.Sizeof(l.keys[lastItemIndex]))

	l.items.Delete(l.keys[lastItemIndex])
	l.mux.Lock()
	l.linearCurrentSize -= itemSize
	l.keys = removeItemByIndex(l.keys, lastItemIndex) //Update keys slice after remove that key out of the map
	l.mux.Unlock()

	return item, nil
}

// Take return and remove the first item out of the linear
func (l *Linear) Take() (interface{}, error) {

	// Execution conditions
	if l.IsEmpty() {
		return nil, errors.New("can't take, because linear is empty")
	}

	item, ok := l.items.Load(l.keys[0])
	if !ok {
		return nil, nil
	}

	itemSize := int64(unsafe.Sizeof(item)) + int64(unsafe.Sizeof(l.keys[0]))

	l.items.Delete(l.keys[0])
	l.mux.Lock()
	l.linearCurrentSize -= itemSize
	l.keys = removeItemByIndex(l.keys, 0) //Update keys slice after remove that key/value from items map
	l.mux.Unlock()

	return item, nil
}

// Get method return and remove the item by key out of the linear
func (l *Linear) Get(key string) (interface{}, error) {

	// Execution conditions
	if l.IsEmpty() {
		return nil, errors.New("linear is empty")
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

	l.items.Delete(key)
	l.mux.Lock()
	l.linearCurrentSize -= int64(unsafe.Sizeof(item))
	l.keys = removeItemByIndex(l.keys, itemIndex) //Update keys slice after remove that key from items map
	l.mux.Unlock()

	return item, nil
}

// Read method return the item by key from linear without remove it
func (l *Linear) Read(key string) (interface{}, error) {

	// Execution conditions
	if l.IsEmpty() {
		return nil, errors.New("linear is empty")
	}

	item, ok := l.items.Load(key)
	if !ok {
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
		return errors.New("linear is empty")
	}

	newItemSize := int64(unsafe.Sizeof(key)) + int64(unsafe.Sizeof(value))
	if newItemSize > l.linearSizes || l.IsEmpty() {
		return errors.New("linear is empty or not enough space")
	}

	currentSize, exits := l.IsExits(key)
	if !exits {
		return errors.New("key does not exit")
	}

	l.items.Store(key, value)
	l.mux.Lock()
	l.linearCurrentSize -= currentSize
	l.linearCurrentSize += newItemSize
	l.mux.Unlock()

	return nil
}

// Range the LinearClient
func (l *Linear) Range(fn func(key, value interface{}) bool) {
	l.items.Range(fn)
}

// IsExits check key exits or not and return size and status
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

	l.mux.Lock()
	l.linearSizes = linearSizes
	l.mux.Unlock()

	return nil
}

// GetLinearCurrentSize return the current linear size
func (l *Linear) GetLinearCurrentSize() int64 {

	l.mux.RLock()
	currentSize := l.linearCurrentSize
	l.mux.RUnlock()

	return currentSize
}
