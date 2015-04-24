package registry

import (
	"fmt"
	"sync"
	//"time"
	"errors"
	"sort"

	"linksmart.eu/services/historical-datastore/Godeps/_workspace/src/code.google.com/p/go-uuid/uuid"
	"linksmart.eu/services/historical-datastore/common"
)

var ErrorNotFound = errors.New("Data source is not found!")

// In-memory storage
type MemoryStorage struct {
	data map[string]DataSource
	//index []string
	mutex sync.RWMutex
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make(map[string]DataSource),
	}
}

func (ms *MemoryStorage) add(ds *DataSource) error {

	// Get a new UUID and convert it to string (UUID type can't be used as map-key)
	newUUID := fmt.Sprint(uuid.NewRandom())
	fmt.Println("New unique id: ", newUUID)

	// Initialize read-only fields
	ds.ID = newUUID
	ds.URL = fmt.Sprintf("%s/%s", common.RegistryAPILoc, ds.ID)
	ds.Data = fmt.Sprintf("%s/%s", common.DataAPILoc, ds.ID)

	ms.data[newUUID] = *ds
	fmt.Println("Added DS: ", ms.data[newUUID])

	return nil
}

func (ms *MemoryStorage) update(id string, ds *DataSource) error {
	ms.mutex.Lock()

	_, ok := ms.data[id]
	if !ok {
		ms.mutex.Unlock()
		return ErrorNotFound
	}

	tempDS := ms.data[id]

	// Update writable elements
	ms.data[id] = DataSource{
		ID:          tempDS.ID,
		URL:         tempDS.URL,
		Data:        tempDS.Data,
		Resource:    ds.Resource,
		Meta:        ds.Meta,
		Retention:   ds.Retention,
		Aggregation: ds.Aggregation,
		Type:        ds.Type,
		Format:      ds.Format,
	}

	ms.mutex.Unlock()

	return nil
}

func (ms *MemoryStorage) delete(id string) error {
	ms.mutex.Lock()

	_, ok := ms.data[id]
	if !ok {
		ms.mutex.Unlock()
		return ErrorNotFound
	}

	delete(ms.data, id)
	ms.mutex.Unlock()

	return nil
}

func (ms *MemoryStorage) get(id string) (DataSource, error) {
	fmt.Println("Getting ds with id: ", id)
	fmt.Println("Content: ", ms.data[id])

	ms.mutex.RLock()
	ds, ok := ms.data[id]
	if !ok {
		ms.mutex.RUnlock()
		return ds, ErrorNotFound
	}
	ms.mutex.RUnlock()

	return ds, nil
}

func (ms *MemoryStorage) getMany(page, perPage int) ([]DataSource, int, error) {
	ms.mutex.RLock()
	total := len(ms.data)

	// Extract keys out of maps
	allKeys := make([]string, 0, total)
	for k := range ms.data {
		allKeys = append(allKeys, k)
	}
	// Sort keys
	sort.Strings(allKeys)

	// Get the queried page
	pagedKeys := getPageOfSlice(allKeys, page, perPage, MaxPerPage)

	// Empty registry
	if len(pagedKeys) == 0 {
		ms.mutex.RUnlock()
		return []DataSource{}, total, nil
	}

	datasources := make([]DataSource, 0, len(pagedKeys))
	for _, k := range pagedKeys {
		datasources = append(datasources, ms.data[k])
	}
	ms.mutex.RUnlock()

	return datasources, total, nil
}

func getCount() int {
	// TODO
	return 0
}

func pathFilterOne(path, op, value string) (DataSource, error) {
	// TODO
	return DataSource{}, nil
}

func pathFilter(path, op, value string, page, perPage int) ([]DataSource, int, error) {
	// TODO
	return []DataSource{}, 0, nil
}

// Utilities from LSLC

// Returns a 'slice' of the given slice based on the requested 'page'
func getPageOfSlice(slice []string, page, perPage, maxPerPage int) []string {
	keys := []string{}
	page, perPage = validatePagingParams(page, perPage, maxPerPage)

	// Never return more than the defined maximum
	if perPage > maxPerPage || perPage == 0 {
		perPage = maxPerPage
	}

	// if 1, not specified or negative - return the first page
	if page < 2 {
		// first page
		if perPage > len(slice) {
			keys = slice
		} else {
			keys = slice[:perPage]
		}
	} else if page == int(len(slice)/perPage)+1 {
		// last page
		keys = slice[perPage*(page-1):]

	} else if page <= len(slice)/perPage && page*perPage <= len(slice) {
		// slice
		r := page * perPage
		l := r - perPage
		keys = slice[l:r]
	}
	return keys
}

func validatePagingParams(page, perPage, maxPerPage int) (int, int) {
	// use defaults if not specified
	if page == 0 {
		page = 1
	}
	if perPage == 0 || perPage > maxPerPage {
		perPage = maxPerPage
	}

	return page, perPage
}
