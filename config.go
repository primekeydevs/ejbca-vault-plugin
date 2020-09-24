/*
 * Copyright (c) 2019 Entrust Datacard Corporation.
 * All rights reserved.
 */
/*
 * Modified by PrimeKey Solutions AB, Copyright (c) 2020
 */

package main

import (
	"log"
	"context"
	"github.com/hashicorp/vault/logical"
	"github.com/pkg/errors"
)

func getConfigCA(ctx context.Context, req *logical.Request, profileId string) (*EJBCAConfigProfile, error) {
	storageEntry, err := req.Storage.Get(ctx, "config/" + profileId)

	if (err != nil || storageEntry == nil) {
		log.Println("Missing configuration for profileId: " + profileId)
		return nil, errors.New("Configuration could not be loaded for profileId: " + profileId)
	}

	configCa := EJBCAConfigProfile{}
	err = storageEntry.DecodeJSON(&configCa)

	if err != nil {
		return nil, errors.Wrap(err, "Configuration could not be parsed")
	}

	return &configCa, nil
}

