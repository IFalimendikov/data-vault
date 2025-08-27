package services

import "context"

// PingServer checks if the server connection is alive
func (v *Vault) PingServer(ctx context.Context) bool {
	return v.grpcclient.PingServer(ctx)
}
