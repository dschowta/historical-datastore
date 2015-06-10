package run

import (
	"errors"
	"fmt"
	"os/user"
	"path/filepath"

	"linksmart.eu/services/historical-datastore/Godeps/_workspace/src/github.com/influxdb/influxdb/cluster"
	"linksmart.eu/services/historical-datastore/Godeps/_workspace/src/github.com/influxdb/influxdb/meta"
	"linksmart.eu/services/historical-datastore/Godeps/_workspace/src/github.com/influxdb/influxdb/services/admin"
	"linksmart.eu/services/historical-datastore/Godeps/_workspace/src/github.com/influxdb/influxdb/services/collectd"
	"linksmart.eu/services/historical-datastore/Godeps/_workspace/src/github.com/influxdb/influxdb/services/continuous_querier"
	"linksmart.eu/services/historical-datastore/Godeps/_workspace/src/github.com/influxdb/influxdb/services/graphite"
	"linksmart.eu/services/historical-datastore/Godeps/_workspace/src/github.com/influxdb/influxdb/services/hh"
	"linksmart.eu/services/historical-datastore/Godeps/_workspace/src/github.com/influxdb/influxdb/services/httpd"
	"linksmart.eu/services/historical-datastore/Godeps/_workspace/src/github.com/influxdb/influxdb/services/monitor"
	"linksmart.eu/services/historical-datastore/Godeps/_workspace/src/github.com/influxdb/influxdb/services/opentsdb"
	"linksmart.eu/services/historical-datastore/Godeps/_workspace/src/github.com/influxdb/influxdb/services/retention"
	"linksmart.eu/services/historical-datastore/Godeps/_workspace/src/github.com/influxdb/influxdb/services/udp"
	"linksmart.eu/services/historical-datastore/Godeps/_workspace/src/github.com/influxdb/influxdb/tsdb"
)

// Config represents the configuration format for the influxd binary.
type Config struct {
	Meta      meta.Config      `toml:"meta"`
	Data      tsdb.Config      `toml:"data"`
	Cluster   cluster.Config   `toml:"cluster"`
	Retention retention.Config `toml:"retention"`

	Admin     admin.Config      `toml:"admin"`
	HTTPD     httpd.Config      `toml:"http"`
	Graphites []graphite.Config `toml:"graphite"`
	Collectd  collectd.Config   `toml:"collectd"`
	OpenTSDB  opentsdb.Config   `toml:"opentsdb"`
	UDP       udp.Config        `toml:"udp"`

	// Snapshot SnapshotConfig `toml:"snapshot"`
	Monitoring      monitor.Config            `toml:"monitoring"`
	ContinuousQuery continuous_querier.Config `toml:"continuous_queries"`

	HintedHandoff hh.Config `toml:"hinted-handoff"`
}

// NewConfig returns an instance of Config with reasonable defaults.
func NewConfig() *Config {
	c := &Config{}
	c.Meta = meta.NewConfig()
	c.Data = tsdb.NewConfig()
	c.Cluster = cluster.NewConfig()

	c.Admin = admin.NewConfig()
	c.HTTPD = httpd.NewConfig()
	c.Collectd = collectd.NewConfig()
	c.OpenTSDB = opentsdb.NewConfig()

	c.Monitoring = monitor.NewConfig()
	c.ContinuousQuery = continuous_querier.NewConfig()
	c.Retention = retention.NewConfig()
	c.HintedHandoff = hh.NewConfig()

	return c
}

// NewDemoConfig returns the config that runs when no config is specified.
func NewDemoConfig() (*Config, error) {
	c := NewConfig()

	// By default, store meta and data files in current users home directory
	u, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("failed to determine current user for storage")
	}

	c.Meta.Dir = filepath.Join(u.HomeDir, ".influxdb/meta")
	c.Data.Dir = filepath.Join(u.HomeDir, ".influxdb/data")
	c.HintedHandoff.Dir = filepath.Join(u.HomeDir, ".influxdb/hh")

	c.Admin.Enabled = true
	c.Monitoring.Enabled = false

	return c, nil
}

// Validate returns an error if the config is invalid.
func (c *Config) Validate() error {
	if c.Meta.Dir == "" {
		return errors.New("Meta.Dir must be specified")
	} else if c.Data.Dir == "" {
		return errors.New("Data.Dir must be specified")
	} else if c.HintedHandoff.Dir == "" {
		return errors.New("HintedHandoff.Dir must be specified")
	}
	return nil
}
