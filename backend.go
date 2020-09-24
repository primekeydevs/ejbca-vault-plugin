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

	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

// Factory returns a new backend as logical.Backend.
func Factory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {
	b := Backend(conf)
	if err := b.Setup(ctx, conf); err != nil {
		return nil, err
	}
	return b, nil
}

// Backend returns a private embedded struct of framework.Backend.
func Backend(conf *logical.BackendConfig) *backend {
	var b backend
	b.Backend = &framework.Backend{
		Help: "Vault EJBCA Plugin",
		Paths: []*framework.Path{
			pathConfig(&b),
			pathEnrollCSR(&b),
			pathRevokeCert(&b),
			pathIssued(&b),
		},
		PathsSpecial: &logical.Paths{},
		Secrets:     []*framework.Secret{},
		BackendType: logical.TypeLogical,
	}
	return &b
}

type backend struct {
	*framework.Backend
}
