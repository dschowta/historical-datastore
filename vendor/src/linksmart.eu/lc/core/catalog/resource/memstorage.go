package resource

import (
	"fmt"
	"sync"

	avl "github.com/ancientlore/go-avltree"

	"linksmart.eu/lc/core/catalog"
)

// In-memory storage
type MemoryStorage struct {
	sync.RWMutex
	devices *avl.Tree
}

func NewMemoryStorage() *MemoryStorage {
	storage := &MemoryStorage{
		devices: avl.New(operator, 0),
	}

	return storage
}

// Comparison operator for AVL Tree
func operator(a interface{}, b interface{}) int {
	if a.(Device).Id < b.(Device).Id {
		return -1
	} else if a.(Device).Id > b.(Device).Id {
		return 1
	}
	return 0
}

// CRUD
func (s *MemoryStorage) add(d *Device) error {
	s.Lock()
	defer s.Unlock()

	_, duplicate := s.devices.Add(*d)
	if duplicate {
		return &ConflictError{fmt.Sprintf("Device id %s is not unique", d.Id)}
	}

	return nil
}

func (s *MemoryStorage) get(id string) (*Device, error) {
	s.RLock()
	defer s.RUnlock()

	d := s.devices.Find(Device{Id: id})
	if d == nil {
		return nil, &NotFoundError{fmt.Sprintf("Device with id %s is not found", id)}
	}
	device := d.(Device)

	return &device, nil
}

func (s *MemoryStorage) update(id string, d *Device) error {
	s.Lock()
	defer s.Unlock()

	r := s.devices.Remove(Device{Id: id})
	if r == nil {
		return &NotFoundError{fmt.Sprintf("Device with id %s is not found", id)}
	}

	s.devices.Add(*d)

	return nil
}

func (s *MemoryStorage) delete(id string) error {
	s.Lock()
	defer s.Unlock()

	r := s.devices.Remove(Device{Id: id})
	if r == nil {
		return &NotFoundError{fmt.Sprintf("Device with id %s is not found", id)}
	}

	return nil
}

func (s *MemoryStorage) list(page int, perPage int) (Devices, int, error) {
	s.RLock()
	defer s.RUnlock()

	total := s.devices.Len()
	offset, limit := catalog.GetPagingAttr(total, page, perPage, MaxPerPage)

	// page/registry is empty
	if limit == 0 {
		return []Device{}, 0, nil
	}

	devices := make([]Device, 0)
	data := s.devices.Data()
	for i := offset; i < offset+limit; i++ {
		devices = append(devices, data[i].(Device))
	}

	return devices, total, nil
}

func (s *MemoryStorage) total() (int, error) {
	s.RLock()
	defer s.RUnlock()

	return s.devices.Len(), nil
}

func (s *MemoryStorage) Close() error {
	return nil
}
