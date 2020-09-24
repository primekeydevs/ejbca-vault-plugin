/*
 * Copyright (c) 2019 Entrust Datacard Corporation.
 * All rights reserved.
 */

package main

import (
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

func pathIssued(b *backend) *framework.Path {
	ret := &framework.Path{
		Pattern: "issued/" + framework.GenericNameRegex("profileId") + "/?$",

		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ReadOperation:   &framework.PathOperation{Callback: b.opReadCert},
			logical.ListOperation:   &framework.PathOperation{Callback: b.opListCerts},
		},

		HelpSynopsis:    "List issued",
		HelpDescription: "List and read issued certificates",
		Fields:          addCommonFields(map[string]*framework.FieldSchema{}),
	}

	return ret
}
