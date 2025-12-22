package main

import "context"

func main() {
	dexClient, err := newDexClient(
		"localhost:5557", //TODO: load from config/envvar
		"/etc/dex/grpc/ca.crt",
		"/etc/dex/grpc/client.crt",
		"/etc/dex/grpc/client.key",
	)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	// these three you’d probably read from env or a secret
	if err := ensureMsEntraConnector(
		ctx,
		dexClient,
		"tenantId",     //TODO: load from config/envvar
		"clientId",     //TODO: load from config/envvar
		"clientSecret", //TODO: load from config/envvar
	); err != nil {
		panic(err)
	}
}
