package custinflux

import (
	"context"
	"fmt"
	"labs/local-transcoder/internal/configs"
	"labs/local-transcoder/internal/logger"
	"sync"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"go.uber.org/zap"
)

var once sync.Once

var client influxdb2.Client

func Init(ctx context.Context, options ...Optioner) {
	once.Do(func() {
		opts := &Options{}
		for _, o := range options {
			o(opts)
		}

		influxConfigs := configs.Get().InfluxConfigs
		clientOptions := buildInfluxOptions(&influxConfigs)
		c := influxdb2.NewClientWithOptions(
			buildInfluxAddress(&influxConfigs),
			influxConfigs.Token,
			clientOptions,
		)

		client = c

		_, err := client.Health(context.Background())
		if err != nil {
			logger.SFatal("custinflux.Init: check health of the server with client.Health",
				zap.Error(err))
		}

		go opts.registerFunc(ctx, client)
	})
}

func buildInfluxAddress(c *configs.InfluxConfigs) string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func buildInfluxOptions(c *configs.InfluxConfigs) *influxdb2.Options {
	clientOptions := &influxdb2.Options{}

	clientOptions.SetApplicationName(c.Name)
	clientOptions.SetFlushInterval(5000)
	clientOptions.SetLogLevel(2)
	clientOptions.SetUseGZip(true)
	clientOptions.SetMaxRetries(3)
	clientOptions.SetMaxRetryInterval(3000)
	clientOptions.SetBatchSize(20)

	return clientOptions
}

func Client() influxdb2.Client {
	return client
}
