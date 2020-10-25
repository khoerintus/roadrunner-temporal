package temporal

import (
	"github.com/spiral/roadrunner/v2"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/temporalio/roadrunner-temporal"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

const temporalSection = "temporal"

type Config struct {
	Address    string
	Namespace  string
	Activities *roadrunner.Config
}

type Temporal interface {
	GetClient() (client.Client, error)
	GetConfig() Config
	CreateWorker(taskQueue string, options worker.Options) (worker.Worker, error)
}

// inherit roadrunner.rpc.Plugin interface
type Provider struct {
	configProvider config.Provider
	// Temporal config from .rr.yaml
	config Config
	// Temporal connection
	serviceClient client.Client
}

// logger dep also
func (p *Provider) Init(provider config.Provider) error {
	p.configProvider = provider
	return nil
}

func (p *Provider) GetConfig() Config {
	return p.config
}

func (p *Provider) Configure() error {
	return p.configProvider.UnmarshalKey(temporalSection, &p.config)
}

func (p *Provider) Close() error {
	return nil
}

func (p *Provider) Serve() chan error {
	errCh := make(chan error, 1)
	var err error
	p.serviceClient, err = client.NewClient(client.Options{
		HostPort:      p.config.Address,
		Namespace:     p.config.Namespace,
		DataConverter: roadrunner_temporal.NewRRDataConverter(),
	})
	if err != nil {
		errCh <- err
	}

	return errCh
}

func (p *Provider) Stop() error {
	p.serviceClient.Close()
	return nil
}

func (p *Provider) GetClient() (client.Client, error) {
	return p.serviceClient, nil
}

func (p *Provider) CreateWorker(tq string, options worker.Options) (worker.Worker, error) {
	w := worker.New(p.serviceClient, tq, options)
	return w, nil
}

func (p *Provider) Name() string {
	return name
}

//func (t *T) MethodName(argType T1, replyType *T2) error
func (p *Provider) RpcService() (interface{}, error) {
	c, err := p.GetClient()
	if err != nil {
		return nil, err
	}

	return &Rpc{
		client: c,
	}, nil
}