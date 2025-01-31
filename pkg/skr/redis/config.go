package redis

import kmccache "github.com/kyma-project/kyma-metrics-collector/pkg/cache"

type ConfigInf interface {
	NewClient(kmccache.Record) (*Client, error)
}

type Config struct{}
