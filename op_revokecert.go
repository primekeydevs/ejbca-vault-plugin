/*
 * PrimeKey Solutions AB, Copyright (c) 2020
 */

package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

type RevocationResponse struct {
	IssuerDN    string `json:"issuer_dn"`
	SerialNo       string `json:"serial_number"`
	RevocationReason string `json:"revocation_reason"`
	RevocationDate string `json:"revocation_date"`
	Message string `json:"message"`
	Revoked bool `json:"revoked"`
}

func (b *backend) opWriteRevokeCert(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {

	profileId := data.Get("profileId").(string)
	if b.Logger().IsDebug() {
        	b.Logger().Debug("profileId: " + profileId)
        }
	if len(profileId) <= 0 {
		return logical.ErrorResponse("profile is empty"), nil
	}

	serial := data.Get("serial").(string)
	if b.Logger().IsTrace() {
        	b.Logger().Trace("serial: " + serial)
        }
	if len(serial) <= 0 {
		return logical.ErrorResponse("serial is empty"), nil
	}

        // NOT_REVOKED, UNSPECIFIED, KEY_COMPROMISE, CA_COMPROMISE, AFFILIATION_CHANGED, SUPERSEDED, CESSATION_OF_OPERATION, CERTIFICATE_HOLD, REMOVE_FROM_CRL, PRIVILEGES_WITHDRAWN, AA_COMPROMISE
	reason := data.Get("reason").(string)
	if b.Logger().IsTrace() {
        	b.Logger().Trace("reason: " + reason)
        }
	if len(reason) <= 0 {
		return logical.ErrorResponse("reason is empty"), nil
	}

	var err error
	configCa, err := getConfigCA(ctx, req, profileId)
	if err != nil  {
		return logical.ErrorResponse("Error getting CA configuration", err), nil
	}
	// Construct revocation request URL
	tlsClientConfig, err := getTLSConfig(ctx, req, configCa)
	if err != nil {
		return logical.ErrorResponse("Error retrieving TLS configuration: %v", err), err
	}

	tr := &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: tlsClientConfig,
	}

        revokeUrl := configCa.URL+"/certificate/"+configCa.SubjectDn+"/"+serial+"/revoke?reason="+reason
        b.Logger().Info("revokeUrl: " + revokeUrl)
	client := &http.Client{Transport: tr}
        // HTTP PUT requires different handling from GET and POST
        request, err := http.NewRequest(http.MethodPut, revokeUrl, nil)
        resp, err := client.Do(request)
	if err != nil {
		return logical.ErrorResponse("Error response: %v", err), err
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return logical.ErrorResponse("CA response could not be read: %v", err), err
	}

	if b.Logger().IsDebug() {
		b.Logger().Debug("response body: " + string(responseBody))
	}

	err = CheckForError(b, responseBody, resp.StatusCode)
	if err != nil {
		return logical.ErrorResponse("Error response received from server: %v", err), err
	}

	var revocationResponse RevocationResponse
	err = json.Unmarshal(responseBody, &revocationResponse)
	if err != nil {
		return logical.ErrorResponse("CA revocation response could not be parsed: %v", err), err
	}

	var respData map[string]interface{}
	if len(revocationResponse.IssuerDN) <= 0 {
		return logical.ErrorResponse("Revocation response issuer DN is empty"), nil
	}
	if len(revocationResponse.SerialNo) <= 0 {
		return logical.ErrorResponse("Revoction response serial number is empty"), nil
	}

	// Remove stored issued certificate from Vault, so it does not show up in the list any longer

        if (revocationResponse.Message == "Successfully revoked") {
                // StorageEntry is constructed this way when storing issued certificates
		err = req.Storage.Delete(ctx, "issued/"+profileId+"/"+revocationResponse.SerialNo)
		if err != nil {
			return logical.ErrorResponse("Could not remove certificate"), err
		}
                b.Logger().Info("Removed certificate from storage: " + revocationResponse.SerialNo)
        } else {
        	return logical.ErrorResponse("Revoking certificate does ot return success: "+revocationResponse.Message), err
        }

	return &logical.Response{
		Data: respData,
	}, nil

}

