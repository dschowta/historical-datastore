// Copyright 2016 Fraunhofer Institute for Applied Information Technology FIT

package aggregation

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"code.linksmart.eu/hds/historical-datastore/common"
	"code.linksmart.eu/hds/historical-datastore/data"
	"code.linksmart.eu/hds/historical-datastore/registry"
	"code.linksmart.eu/sc/service-catalog/utils"
	"github.com/gorilla/mux"
)

const (
	MaxPerPage = 1000
)

var (
	ErrNotImplemented = errors.New("API not implemented")
)

type API struct {
	registryClient registry.Client
	storage        Storage
}

func NewAPI(registryClient registry.Client, storage Storage) *API {
	return &API{registryClient, storage}
}

func (api *API) Index(w http.ResponseWriter, r *http.Request) {

	aggrs, err := api.Aggregations()
	if err != nil {
		common.ErrorResponse(http.StatusInternalServerError, "Error reading registry: "+err.Error(), w)
		return
	}

	var index Index
	index.Aggrs = make([]Aggregation, 0, len(aggrs))
	for _, v := range aggrs {
		index.Aggrs = append(index.Aggrs, v)
	}

	b, err := json.Marshal(&index)
	if err != nil {
		common.ErrorResponse(http.StatusInternalServerError, "Error Marshalling: "+err.Error(), w)
		return
	}

	w.Header().Add("Content-Type", common.DefaultMIMEType)
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func (api *API) Filter(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	fpath := params["path"]
	fop := params["op"]
	fvalue := params["value"]
	pathTknz := strings.Split(fpath, ".")

	aggrs, err := api.Aggregations()
	if err != nil {
		common.ErrorResponse(http.StatusInternalServerError, "Error reading registry: "+err.Error(), w)
		return
	}

	var index Index
	index.Aggrs = make([]Aggregation, 0, len(aggrs))
	for _, aggr := range aggrs {
		matched, err := utils.MatchObject(aggr, pathTknz, fop, fvalue)
		if err != nil {
			common.ErrorResponse(http.StatusInternalServerError, "Error matching aggregation: "+err.Error(), w)
			return
		}
		if matched {
			index.Aggrs = append(index.Aggrs, aggr)
		}
	}

	b, err := json.Marshal(&index)
	if err != nil {
		common.ErrorResponse(http.StatusInternalServerError, "Error marshalling: "+err.Error(), w)
		return
	}

	w.Header().Set("Content-Type", common.DefaultMIMEType)
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func (api *API) Query(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	timeStart := time.Now()
	params := mux.Vars(r)
	aggrID := params["aggrid"]

	page, perPage, err := common.ParsePagingParams(r.Form.Get(common.ParamPage), r.Form.Get(common.ParamPerPage), MaxPerPage)
	if err != nil {
		common.ErrorResponse(http.StatusBadRequest, err.Error(), w)
		return
	}

	ids := strings.Split(params["uuid"], common.IDSeparator)
	if len(ids) == 0 {
		common.ErrorResponse(http.StatusBadRequest,
			"Source IDs not specified.", w)
		return
	}

	q, err := data.ParseQueryParameters(r.Form)
	if err != nil {
		common.ErrorResponse(http.StatusBadRequest, err.Error(), w)
		return
	}

	err = common.ValidatePerItemLimit(q.Limit, perPage, len(ids))
	if err != nil {
		common.ErrorResponse(http.StatusBadRequest, err.Error(), w)
		return
	}

	// Query sources from registry
	var sources []registry.DataSource

OUTERLOOP:
	for _, id := range ids {
		ds, err := api.registryClient.Get(id)
		if err != nil {
			common.ErrorResponse(http.StatusNotFound,
				fmt.Sprintf("Error retrieving data source %v from the registry: %v", id, err.Error()), w)
			return
		}
		sources = append(sources, ds)

		// Check if ds has the aggregation
		for _, dsa := range ds.Aggregation {
			if dsa.ID == aggrID {
				continue OUTERLOOP
			}
		}
		common.ErrorResponse(http.StatusNotFound, fmt.Sprintf("Data source %v does not have aggregation %v", id, aggrID), w)
		return
	}

	// Retrieve the aggregation object
	var aggr registry.Aggregation
	for _, dsa := range sources[0].Aggregation {
		if dsa.ID == aggrID {
			aggr = dsa
			break
		}
	}

	// Query aggregated data from storage
	dataset, total, err := api.storage.Query(aggr, q, page, perPage, sources...)
	if err == ErrNotImplemented {
		common.ErrorResponse(http.StatusNotImplemented, err.Error(), w)
		return
	} else if err != nil {
		common.ErrorResponse(http.StatusInternalServerError, "Error retrieving data from the database: "+err.Error(), w)
		return
	}

	v := url.Values{}
	v.Add(common.ParamStart, q.Start.Format(time.RFC3339))
	// Omit end in open-ended queries
	if q.End.After(q.Start) {
		v.Add(common.ParamEnd, q.End.Format(time.RFC3339))
	}
	v.Add(common.ParamSort, q.Sort)
	if q.Limit > 0 { // non-positive limit is ignored
		v.Add(common.ParamLimit, fmt.Sprintf("%d", q.Limit))
	}
	v.Add(common.ParamPage, fmt.Sprintf("%d", page))
	v.Add(common.ParamPerPage, fmt.Sprintf("%d", perPage))
	recordSet := RecordSet{
		URL:     fmt.Sprintf("%s?%s", r.URL.Path, v.Encode()),
		Data:    dataset,
		Time:    time.Since(timeStart).Seconds() * 1000,
		Page:    page,
		PerPage: perPage,
		Total:   total,
	}

	b, err := json.Marshal(recordSet)
	if err != nil {
		common.ErrorResponse(http.StatusInternalServerError, "Error marshalling recordset: "+err.Error(), w)
		return
	}

	w.Header().Set("Content-Type", common.DefaultMIMEType)
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

// Utility functions

// Retrieve aggregations from registry api
func (api *API) Aggregations() (map[string]Aggregation, error) {
	aggrs := make(map[string]Aggregation)
	perPage := 100
	for page := 1; ; page++ {
		datasources, total, err := api.registryClient.GetDataSources(page, perPage)
		if err != nil {
			return aggrs, err
		}

		for _, ds := range datasources {
			for _, dsa := range ds.Aggregation {
				var aggr Aggregation
				aggr.ID = dsa.ID
				aggr.Interval = dsa.Interval
				aggr.Aggregates = dsa.Aggregates
				aggr.Retention = dsa.Retention
				var sources []string
				a, found := aggrs[dsa.ID]
				if found {
					sources = a.Sources
				}
				aggr.Sources = append(sources, ds.ID)
				aggrs[dsa.ID] = aggr
			}
		}

		if page*perPage >= total {
			break
		}
	}

	return aggrs, nil
}