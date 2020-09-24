/*
 * PrimeKey Solutions AB, Copyright (c) 2020
 */

package main

import (
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

func pathRevokeCert(b *backend) *framework.Path {
	ret := &framework.Path{
		Pattern: "revokeCert/" + framework.GenericNameRegex("profileId") + "/?$",

		Operations: map[logical.Operation]framework.OperationHandler{
			logical.UpdateOperation: &framework.PathOperation{Callback: b.opWriteRevokeCert},
		},

		HelpSynopsis:    "Certificate revocation",
		HelpDescription: "Revoke a certificate",
		Fields:          addCommonFields(map[string]*framework.FieldSchema{}),
	}

	ret.Fields["serial"] = &framework.FieldSchema{
		Type:        framework.TypeString,
		Default:     "",
		Description: "Hex encoded serial number of the certificate to revoke.",
		Required:    true,
	}

	ret.Fields["reason"] = &framework.FieldSchema{
		Type:        framework.TypeString,
		Default:     "",
		Description: "Revocation reason, one of: NOT_REVOKED, UNSPECIFIED, KEY_COMPROMISE, CA_COMPROMISE, AFFILIATION_CHANGED, SUPERSEDED, CESSATION_OF_OPERATION, CERTIFICATE_HOLD, REMOVE_FROM_CRL, PRIVILEGES_WITHDRAWN, AA_COMPROMISE",
		Required:    true,
	}

	return ret
}
