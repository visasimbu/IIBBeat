package beater

import (
	"fmt"
	"time"
	"encoding/json"
	"github.com/mygitlab/iibbeat/mqsi"

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

		bytebuf := mqsi.PullnodeInfoJson(mqsiscriptdir)
		
		var jsonevent common.MapStr
		if err := json.Unmarshal(bytebuf, &jsonevent); err != nil {
			logp.Error(err)
		}		
		
		event := beat.Event{			
			Timestamp: time.Now(),
			Fields: common.MapStr{
				"type":    b.Info.Name,
				"event": jsonevent,
			},
		}
		bt.client.Publish(event)

		logp.Info("IIB Event sent")
	}
}

// Stop stops iibbeat.
func (bt *iibbeat) Stop() {
	bt.client.Close()
	close(bt.done)
}
