package data

import (
	"fmt"
	"time"

	"code.linksmart.eu/hds/historical-datastore/common"
	"code.linksmart.eu/hds/historical-datastore/registry"
	datastore "github.com/dschowta/senml.datastore"
	"github.com/farshidtz/senml"
)

type LightdbStorage struct {
	storage *datastore.SenmlDataStore
}

func NewSenmlStorage(conf common.DataConf) (storage *LightdbStorage, disconnect_func func() error, err error) {
	datastore := new(datastore.SenmlDataStore)
	err = datastore.Connect(conf.Backend.DSN)
	if err != nil {
		return nil, nil, err
	}
	storage = new(LightdbStorage)
	storage.storage = datastore
	return storage, storage.Disconnect, nil
}

func (s *LightdbStorage) Submit(data map[string]senml.Pack, sources map[string]*registry.DataStream) error {
	for _, dps := range data {
		err := s.storage.Add(dps)
		if err != nil {
			return fmt.Errorf("error creating batch points: %s", err)
		}
	}
	return nil
}

func (s *LightdbStorage) Query(q Query, sources ...*registry.DataStream) (senml.Pack, int, *time.Time, error) {
	//TODO: Support multidimensional queries
	/*Multi dimensional queries have problems with pagination:

	1. Multinextlinks (each dimension in a multidimensional time series gives a next link)
		1. +handles all the multidimensional scenarios
		2. - overhead on client to keep track of time series
	2. Combined response: (a combined list is given with a single next link)
		1. +handles all the multidimensional scnarios
		2. - overhead on server to combine the results and deduce the nextlink

	*/
	//TODO: Is this a right place to decide the maxentries? Should be at API level
	maxEntries := q.perPage
	if q.Limit > 0 && q.perPage > q.Limit { //if limit is provided by the user and it is less than perPage, then use the limit
		maxEntries = q.Limit
	}

	senmlQuery := datastore.Query{
		Start:      datastore.ToSenmlTime(q.Start),
		End:        datastore.ToSenmlTime(q.End),
		MaxEntries: maxEntries,
		Series:     sources[0].Name,
		Sort:       q.Sort,
	}
	retPack, nextlink, err := s.storage.Query(senmlQuery)
	if err != nil {
		return nil, 0, nil, err
	}

	var nextLinkTime *time.Time

	if nextlink != nil {
		t := datastore.FromSenmlTime(*nextlink)
		nextLinkTime = &t
	}

	return retPack, len(retPack), nextLinkTime, nil
}

func (s *LightdbStorage) Disconnect() error {
	return s.storage.Disconnect()
}

// CreateHandler handles the creation of a new data source
func (s *LightdbStorage) CreateHandler(ds registry.DataStream) error {
	return nil
}

// UpdateHandler handles updates of a data source
func (s *LightdbStorage) UpdateHandler(oldDS registry.DataStream, newDS registry.DataStream) error {
	//TODO supporting retetion

	return nil
}

// DeleteHandler handles deletion of a data source
func (s *LightdbStorage) DeleteHandler(ds registry.DataStream) error {
	err := s.storage.Delete(ds.Name)
	if err != nil && err != datastore.ErrSeriesNotFound {
		return err
	}
	//log.Println("LightdbStorage: dropped measurements for", ds.Name)
	return nil
}
