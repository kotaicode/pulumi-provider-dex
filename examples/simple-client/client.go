package main

import (
	"fmt"

	api "github.com/dexidp/dex/api/v2"
	"google.golang.org/grpc"
)

func newDexClient(dexAddr, caPath, clientCertPath, clientKeyPath string) (api.DexClient, error) {
	// // CA
	// caPEM, err := os.ReadFile(caPath)
	// if err != nil {
	// 	return nil, fmt.Errorf("read CA: %w", err)
	// }
	// roots := x509.NewCertPool()
	// if !roots.AppendCertsFromPEM(caPEM) {
	// 	return nil, fmt.Errorf("no certs in CA")
	// }

	// // mTLS client cert (optional but recommended)
	// cert, err := tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
	// if err != nil {
	// 	return nil, fmt.Errorf("load client cert: %w", err)
	// }

	// tlsCfg := &tls.Config{
	// 	RootCAs:      roots,
	// 	Certificates: []tls.Certificate{cert},
	// }

	//conn, err := grpc.Dial(dexAddr, grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg)))
	conn, err := grpc.Dial(dexAddr, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("dial dex: %w", err)
	}

	return api.NewDexClient(conn), nil
}
