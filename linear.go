package linear

import (
	"errors"
	"reflect"
	"sync"
)

// Linear contains all attributes
type Linear struct {
	items              sync.Map
	keys               []string
	sizeChecker        bool
	linearSizes        int64
	linearCurrentSizes int64
	rwMutex            sync.RWMutex
}

// New return new instance
func New(linearSizes int64, sizeChecker bool) *Linear {
	currentLinear := Linear{
		items:              sync.Map{},
		keys:               []string{},
		sizeChecker:        sizeChecker,
		linearSizes:        linearSizes,
		linearCurrentSizes: 0,
		rwMutex:            sync.RWMutex{},
	}

	return &currentLinear
}

// Push item to the linear by key
func (l *Linear) Push(key string, value interface{}) error {
	itemSize := int64(reflect.Type.Size(key) + reflect.Type.Size(value))
	if itemSize > l.linearSizes || key == "" {
		return errors.New("key is empty or linear not enough space")
	}

	// Clean space for new item
	if l.sizeChecker {
		for l.linearCurrentSizes+itemSize > l.linearSizes {
			l.Take()
		}
	}

	l.rwMutex.Lock()
	l.linearCurrentSizes += int64(reflect.Type.Size(value))
	l.keys = append(l.keys, key)
	l.items.LoadOrStore(key, value)
	l.rwMutex.Unlock()

	return nil
}

// Pop return the last item from the linear and remove it
func (l *Linear) Pop() (interface{}, error) {
	if l.IsEmpty() {
		return nil, errors.New("the linear is empty")
	}

	lastItemIndex := len(l.keys) - 1
	item, exits := l.items.Load(l.keys[lastItemIndex])
	if !exits {
		return nil, errors.New("this key does not exits")
	}

	l.rwMutex.Lock()
	l.linearCurrentSizes = l.linearCurrentSizes - int64(reflect.Type.Size(item))
	l.items.Delete(l.keys[lastItemIndex])
	l.keys = removeItemByIndex(l.keys, lastItemIndex) //Update keys slice after remove that key from items map
	l.rwMutex.Unlock()

	return item, nil
}

// Take return the first item from the linear and remove it
func (l *Linear) Take() (interface{}, error) {
	if l.IsEmpty() {
		return nil, errors.New("the linear is empty")
	}

	l.rwMutex.Lock()
	item, exits := l.items.Load(l.keys[0])
	if !exits {
		l.rwMutex.Unlock()
		return nil, errors.New("that key does not exits")
	}

	l.linearCurrentSizes -= int64(reflect.Type.Size(item))
	l.items.Delete(l.keys[0])
	l.keys = removeItemByIndex(l.keys, 0) //Update keys slice after remove that key from items map
	l.rwMutex.Unlock()

	return item, nil
}

// Get method return the item by key from linear and remove it
// Goroutine: https://stackoverflow.com/questions/20945069/catching-return-values-from-goroutines
func (l *Linear) Get(key string) (interface{}, error) {
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

	l.rwMutex.Lock()
	l.items.Delete(key)
	l.keys = removeItemByIndex(l.keys, itemIndex) //Update keys slice after remove that key from items map
	l.rwMutex.Unlock()

	return item, nil
}

// Read method return the item by key from linear without remove it
func (l *Linear) Read(key string) (interface{}, error) {
	if l.IsEmpty() {
		return nil, errors.New("linear is empty")
	}

	item, exits := l.items.Load(key)
	if !exits {
		return nil, errors.New("that key does not exits")
	}

	return item, nil
}

// Range the Linear
func (l *Linear) Range(fn func(key, value interface{}) bool) {
	l.items.Range(fn)
}

// IsEmpty check linear size
func (l *Linear) IsEmpty() bool {
	return len(l.keys) == 0
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
func (l *Linear) SetLinearSizes(linearSizes int64) {
	l.rwMutex.RLock()
	l.linearSizes = linearSizes
	l.rwMutex.RUnlock()
}

// GetLinearCurrentSize return the current linear size
func (l *Linear) GetLinearCurrentSize() int64 {
	return l.linearCurrentSizes
}

// removeItemByIndex remove item out of []string by index but maintains order, and return the new one
// Source: https://yourbasic.org/golang/delete-element-slice/
func removeItemByIndex(s []string, idx int) []string {
	copy(s[idx:], s[idx+1:]) // Shift s[idx+1:] left one index.
	s[len(s)-1] = ""         // Erase last element (write zero value).
	return s[:len(s)-1]      // Truncate s.
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
