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

// Client represents a Dgraph database client with configuration.
type Client struct {
	Cli    *dgo.Dgraph
	config DgraphConfig
}

// DgraphConfig contains configuration for connecting to a Dgraph server.
type DgraphConfig struct {
	Credentials config.Credentials
	GrpcServer  config.Server
}

// NewClientCtx creates a new Dgraph client with the given context and configuration.
// It establishes a connection to the Dgraph server using the provided credentials
// and gRPC server details. The connection uses insecure transport credentials.
// Returns the client or an error if connection fails.
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

// Migrate applies database schema migrations from the provided filesystem.
// The migrations should be located in "migrations/dgraph/schema.gql".
// Returns an error if migration fails.

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

// Close terminates the connection to the Dgraph server.
// The provided context is currently unused but maintained for future compatibility.
// Always returns nil error.
func (cli *Client) Close(_ context.Context) error {
	cli.Cli.Close()
	return nil
}
