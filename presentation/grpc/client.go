package grpc

import (
	authinternal "github.com/hiamthach108/dreon-auth/presentation/grpc/gen/proto"
	"google.golang.org/grpc"
)

// AuthInternalClient wraps the gRPC connection and generated AuthInternalService client
// for internal auth RPCs (relation tuples and permission checks).
// Call Close() when done to release the connection.
type AuthInternalClient struct {
	conn   *grpc.ClientConn
	client authinternal.AuthInternalServiceClient
}

// NewAuthInternalClientFromConn builds an AuthInternalClient from an existing gRPC connection.
// The client does not take ownership of conn; the caller is responsible for closing it
// via AuthInternalClient.Close() when using this constructor after a dial.
func NewAuthInternalClientFromConn(conn *grpc.ClientConn) *AuthInternalClient {
	return &AuthInternalClient{
		conn:   conn,
		client: authinternal.NewAuthInternalServiceClient(conn),
	}
}

// Client returns the generated AuthInternalServiceClient for making RPCs.
func (c *AuthInternalClient) Client() authinternal.AuthInternalServiceClient {
	return c.client
}

// Close closes the underlying gRPC connection. No-op if already closed.
func (c *AuthInternalClient) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}
