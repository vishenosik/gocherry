package graph

import (
	"context"
	"io/fs"
	"path"

	"github.com/dgraph-io/dgo/v240"
	migrate "github.com/vishenosik/dmigrate"
	"github.com/vishenosik/gocherry/pkg/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	Cli    *dgo.Dgraph
	config DgraphConfig
}

type DgraphConfig struct {
	Credentials config.Credentials
	GrpcServer  config.Server
}

func NewClientCtx(ctx context.Context, config DgraphConfig) (*Client, error) {

	client, err := dgo.NewClient(
		config.GrpcServer.String(),
		// add Dgraph ACL credentials
		dgo.WithACLCreds(config.Credentials.User, config.Credentials.Password),
		// add insecure transport credentials
		dgo.WithGrpcOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		Cli:    client,
		config: config,
	}, nil
}

func (cli *Client) Migrate(migrations fs.FS) error {

	migrator, err := migrate.NewDgraphMigratorContext(
		context.TODO(),
		migrate.Config{
			User:     cli.config.Credentials.User,
			Password: cli.config.Credentials.Password,
			Host:     cli.config.GrpcServer.Host,
			Port:     cli.config.GrpcServer.Port,
			Timeout:  cli.config.GrpcServer.Timeout,
		},
		migrations,
	)
	if err != nil {
		return err
	}

	if err := migrator.Up(path.Join("migrations", "dgraph", "schema.gql")); err != nil {
		return err
	}

	return nil
}
