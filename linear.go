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
	sizeMux sync.RWMutex
	keysMux sync.RWMutex
	linearCurrentSize int64
}

func (l *Linear) calculateItemSize(key string, value interface{}) int64 {
	size := int64(len(key))
	
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		size += int64(v.Len())
	case reflect.Slice, reflect.Array:
		size += int64(v.Len())
	case reflect.Map, reflect.Struct:
		size += 64
	default:
		size += int64(reflect.TypeOf(value).Size())
	}
	return size
}

func New(maxSize int64, sizeChecker bool) *Linear {
	if maxSize <= 0 {
		log.Fatalln("linearSizes much higher than 0")
	}

	return &Linear{
		keys:        make([]string, 0, 16), // Pre-allocate initial capacity
		items:       &sync.Map{},
		sizeChecker: sizeChecker,
		linearSizes: maxSize,
		linearCurrentSize: 0,
	}
}

func (l *Linear) Push(key string, value interface{}) error {
	if key == "" && value == nil {
		return errors.New("key and value should not be empty")
	}

	itemSize := l.calculateItemSize(key, value)
	if itemSize > l.linearSizes {
		return errors.New("linear doesn't have enough memory space")
	}

	if l.sizeChecker {
		for l.linearCurrentSize+itemSize > l.linearSizes {
			if _, err := l.Take(); err != nil {
				return err
			}
		}
	}

	l.items.LoadOrStore(key, value)
	l.sizeMux.Lock()
	l.linearCurrentSize += itemSize
	l.sizeMux.Unlock()

	l.keysMux.Lock()
	l.keys = append(l.keys, key)
	l.keysMux.Unlock()

	return nil
}

func (l *Linear) removeByIndex(index int) (interface{}, int64, error) {
	if l.IsEmpty() {
		return nil, 0, errors.New("linear is empty")
	}

	key := l.keys[index]
	item, ok := l.items.Load(key)
	if !ok {
		return nil, 0, nil
	}

	itemSize := l.calculateItemSize(key, item)
	l.items.Delete(key)

	l.keysMux.Lock()
	l.keys = removeItemByIndex(l.keys, index)
	l.keysMux.Unlock()

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
		return nil, nil
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
		return nil, nil
	}

	return item, nil
}

func (l *Linear) Update(key string, value interface{}) error {
	if key == "" && value == nil {
		return errors.New("key and value should not be empty")
	}

	if l.IsEmpty() {
		return errors.New("linear is empty")
	}

	newItemSize := l.calculateItemSize(key, value)
	if newItemSize > l.linearSizes || l.IsEmpty() {
		return errors.New("linear is empty or not enough space")
	}

	currentSize, exits := l.IsExits(key)
	if !exits {
		return errors.New("key does not exit")
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

func (l *Linear) IsExits(key string) (int64, bool) {
	value, exits := l.items.Load(key)
	if !exits {
		return 0, false
	}
	return l.calculateItemSize(key, value), true
}

func (l *Linear) IsEmpty() bool {
	return len(l.keys) == 0
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
		return errors.New("linearSizes much higher than 0")
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
