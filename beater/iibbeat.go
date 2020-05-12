package beater

import (
	"fmt"
	"time"
	"github.com/mygitlab/iibbeat/mqsi"
	"github.com/gofrs/uuid"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/logp"

	"github.com/mygitlab/iibbeat/config"
)

// iibbeat configuration.
type iibbeat struct {
	done   chan struct{}
	config config.Config
	client beat.Client
}

// New creates an instance of iibbeat.
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	c := config.DefaultConfig
	if err := cfg.Unpack(&c); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &iibbeat{
		done:   make(chan struct{}),
		config: c,
	}
	return bt, nil
}

// Run starts iibbeat.
func (bt *iibbeat) Run(b *beat.Beat) error {
	logp.Info("iibbeat is running! Hit CTRL-C to stop it.")

	var err error
	bt.client, err = b.Publisher.Connect()
	if err != nil {
		return err
	}

	ticker := time.NewTicker(bt.config.Period)
	mqsiscriptdir := string(bt.config.Path)
	for {
		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
		}

		bytebuf := mqsi.PullnodeCollectionInfo(mqsiscriptdir)
		uuid, err := uuid.NewV4()
		if err != nil {
			logp.Err("Error unable to generate UUID")
		}

		for _, node := range bytebuf.Nodes {
			for _, isname := range node.IntegrationServers {
				for _, app := range isname.Components {
					event := beat.Event{
						Timestamp: time.Now(),
						Fields: common.MapStr{
							"type":    b.Info.Name,
							// "counter": counter,
							"mqsiresult": common.MapStr{
								"nodeName":						node.Name,
								"nodeStatus":					node.Status,
								"integrationServerName":		isname.Name,
								"integrationServerStatus":		isname.Status,
								"appName":						app.Name,
								"appStatus":					app.Status,
								"appType":						app.Type,
								"appDeployedTime":				app.DeployedTime,
								"appBarFilename":				app.BarFileName,
								"appBarFileLastModifiedTime":	app.BarFileLastModifiedTime,
								"correlationID":				uuid,					
							},
						},
					}
					bt.client.Publish(event)
				}
			}			
		}
		logp.Info("IIB Event sent")
	}
}

// Stop stops iibbeat.
func (bt *iibbeat) Stop() {
	bt.client.Close()
	close(bt.done)
}
