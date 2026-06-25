package configs

import (
	"github.com/elastic/go-elasticsearch/v8"
)

func ConnectElasticsearch(env *Env) (*elasticsearch.Client, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{env.ElasticURL},
	}
	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return client, nil
}
