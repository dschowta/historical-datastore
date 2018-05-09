// Copyright 2016 Fraunhofer Institute for Applied Information Technology FIT

package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"code.linksmart.eu/com/go-sec/auth/obtainer"
	"code.linksmart.eu/hds/historical-datastore/common"
	"code.linksmart.eu/sc/service-catalog/utils"
)

type RemoteClient struct {
	serverEndpoint *url.URL
	ticket         *obtainer.Client
}

func NewRemoteClient(serverEndpoint string, ticket *obtainer.Client) (*RemoteClient, error) {
	// Check if serverEndpoint is a correct URL
	endpointUrl, err := url.Parse(serverEndpoint)
	if err != nil {
		return nil, logger.Errorf("%s", err)
	}

	return &RemoteClient{
		serverEndpoint: endpointUrl,
		ticket:         ticket,
	}, nil
}

func (c *RemoteClient) Index(page int, perPage int) (*Registry, error) {
	res, err := utils.HTTPRequest("GET",
		fmt.Sprintf("%v?%v=%v&%v=%v", c.serverEndpoint, common.ParamPage, page, common.ParamPerPage, perPage),
		nil,
		nil,
		c.ticket,
	)
	if err != nil {
		return nil, logger.Errorf("%s", err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, logger.Errorf("Unable to read body of response: %v", err.Error())
	}

	if res.StatusCode == http.StatusOK {
		var reg Registry
		err = json.Unmarshal(body, &reg)
		if err != nil {
			return nil, logger.Errorf("%s", err)
		}
		return &reg, nil
	}

	return nil, logger.Errorf("%v: %v", res.StatusCode, string(body))
}

func (c *RemoteClient) Add(d *DataSource) (string, error) {
	b, _ := json.Marshal(d)
	res, err := utils.HTTPRequest("POST",
		c.serverEndpoint.String()+"/",
		map[string][]string{"Content-Type": []string{"application/ld+json"}},
		bytes.NewReader(b),
		c.ticket,
	)
	if err != nil {
		return "", logger.Errorf("%s", err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusCreated {
		// retrieve ID from the header
		loc := res.Header.Get("Location")
		tkz := strings.Split(loc, "/")
		return tkz[len(tkz)-1], nil
	}

	// Get body of error
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", logger.Errorf("Unable to read body of error: %v", err.Error())
	}

	return "", logger.Errorf("%v: %v", res.StatusCode, string(body))
}

func (c *RemoteClient) Get(id string) (*DataSource, error) {
	res, err := utils.HTTPRequest("GET",
		fmt.Sprintf("%v/%v", c.serverEndpoint, id),
		nil,
		nil,
		c.ticket,
	)
	if err != nil {
		return nil, logger.Errorf("%s", err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	} else if res.StatusCode != http.StatusOK {
		return nil, logger.Errorf("%v", res.StatusCode)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, logger.Errorf("%s", err)
	}

	var ds DataSource
	err = json.Unmarshal(body, &ds)
	if err != nil {
		return nil, logger.Errorf("%s", err)
	}

	return &ds, nil
}

func (c *RemoteClient) Update(id string, d *DataSource) error {
	b, _ := json.Marshal(d)
	res, err := utils.HTTPRequest("PUT",
		fmt.Sprintf("%v/%v", c.serverEndpoint, id),
		nil,
		bytes.NewReader(b),
		c.ticket,
	)
	if err != nil {
		return logger.Errorf("%s", err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return logger.Errorf("%s", err)
	}

	if res.StatusCode == http.StatusNotFound {
		return ErrNotFound
	} else if res.StatusCode != http.StatusOK {
		return logger.Errorf("%v: %v", res.StatusCode, string(body))
	}
	return nil
}

func (c *RemoteClient) Delete(id string) error {
	res, err := utils.HTTPRequest("DELETE",
		fmt.Sprintf("%v/%v", c.serverEndpoint, id),
		nil,
		bytes.NewReader([]byte{}),
		c.ticket,
	)
	if err != nil {
		return logger.Errorf("%s", err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return logger.Errorf("%s", err)
	}

	if res.StatusCode == http.StatusNotFound {
		return ErrNotFound
	} else if res.StatusCode != http.StatusOK {
		return logger.Errorf("%v: %v", res.StatusCode, string(body))
	}

	return nil
}

func (c *RemoteClient) FilterOne(path, op, value string) (*DataSource, error) {
	res, err := utils.HTTPRequest("GET",
		fmt.Sprintf("%v/%v/%v/%v/%v", c.serverEndpoint, FTypeOne, path, op, value),
		nil,
		nil,
		c.ticket,
	)
	if err != nil {
		return nil, logger.Errorf("%s", err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, logger.Errorf("%s", err)
	}

	if res.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	} else if res.StatusCode != http.StatusOK {
		return nil, logger.Errorf("%v: %v", res.StatusCode, string(body))
	}

	var ds DataSource
	err = json.Unmarshal(body, &ds)
	if err != nil {
		return nil, logger.Errorf("%s", err)
	}

	return &ds, nil
}

func (c *RemoteClient) FilterMany(path, op, value string) ([]DataSource, error) {
	res, err := utils.HTTPRequest("GET",
		fmt.Sprintf("%v/%v/%v/%v/%v", c.serverEndpoint, FTypeMany, path, op, value),
		nil,
		nil,
		c.ticket,
	)
	if err != nil {
		return nil, logger.Errorf("%s", err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, logger.Errorf("%s", err)
	}

	if res.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	} else if res.StatusCode != http.StatusOK {
		return nil, logger.Errorf("%v: %v", res.StatusCode, string(body))
	}

	var reg Registry
	err = json.Unmarshal(body, &reg)
	if err != nil {
		return nil, logger.Errorf("%s", err)
	}

	return reg.Entries, nil
}
