package services

import "context"

// PingDB checks if the database connection is alive and returns true if successful
func (v *Vault) PingServer(ctx context.Context) bool {
	return v.grpcclient.PingServer(ctx)
}
