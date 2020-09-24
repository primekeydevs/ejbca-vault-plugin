/*
 * Copyright (c) 2019 Entrust Datacard Corporation.
 * All rights reserved.
 */
/*
 * Modified by PrimeKey Solutions AB, Copyright (c) 2020
 */

package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"

	"github.com/hashicorp/vault/logical"
	"github.com/pkg/errors"
)

func getTLSConfig(ctx context.Context, req *logical.Request, configCa *EJBCAConfigProfile) (*tls.Config, error) {
	certificate, err := tls.X509KeyPair([]byte(configCa.PEMBundle), []byte(configCa.PEMBundle))
	if err != nil {
		return nil, errors.Wrap(err, "Error parsing client certificate and key")
	}
	// Don't trust any built in (Public CA?) certificates, use only the one passes to our plugin
	//certPool, _ := x509.SystemCertPool()
	//if certPool == nil {
	//	certPool = x509.NewCertPool()
	//}
	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM([]byte(configCa.CACerts)); !ok {
		return nil, errors.New("Error appending CA certs.")
	}

	tlsClientConfig := tls.Config{
		MaxVersion: tls.VersionTLS12,
        Renegotiation: tls.RenegotiateFreelyAsClient,
		Certificates: []tls.Certificate{
			certificate,
		},
		RootCAs: certPool,
	}

	tlsClientConfig.BuildNameToCertificate()

	return &tlsClientConfig, nil
}
