/*
 * Copyright (c) 2019 Entrust Datacard Corporation.
 * All rights reserved.
 */
/*
 * Modified by PrimeKey Solutions AB, Copyright (c) 2020
 */

package main

import (
	"bytes"
	"context"
//	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
//	"math/big"
	"net/http"
        "github.com/sethvargo/go-password/password"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

type EnrollmentRequest struct {
	CSR                               string                    `json:"certificate_request,omitempty"`
	CertificateProfile                string                    `json:"certificate_profile_name,omitempty"`
	EndEntityProfile                  string                    `json:"end_entity_profile_name,omitempty"`
	CAName                            string                    `json:"certificate_authority_name,omitempty"`
	Username                          string                    `json:"username,omitempty"`
	Password                          string                    `json:"password,omitempty"`
	IncludeChain                      string                    `json:"include_chain,omitempty"`
}

type EnrollmentResponse struct {
	Certificate    string `json:"certificate"`
	SerialNo       string `json:"serial_number"`
	ResponseFormat string `json:"response_format"`
}

func (b *backend) opWriteEnrollCSR(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {

	profileId := data.Get("profileId").(string)
	if b.Logger().IsDebug() {
        	b.Logger().Debug("profileId: " + profileId)
        }
	if len(profileId) <= 0 {
		return logical.ErrorResponse("profile is empty"), nil
	}

	var err error

	username := data.Get("username").(string)
	if b.Logger().IsTrace() {
        	b.Logger().Trace("username: " + username)
        }
	if len(username) <= 0 {
		return logical.ErrorResponse("username is empty"), nil
	}

	csrPem := data.Get("csr").(string)
	// Verify that we can decode it, for user friendlieness
	// Just decode a single block, omit any subsequent blocks
	csrBlock, _ := pem.Decode([]byte(csrPem))
	if csrBlock == nil {
		return logical.ErrorResponse("CSR could not be decoded"), nil
	}

	enrollmentcode, err := password.Generate(32, 20, 0, false, true)
	if err != nil  {
		return logical.ErrorResponse("Random password could not be generated", err), nil
	}

	configCa, err := getConfigCA(ctx, req, profileId)
	if err != nil  {
		return logical.ErrorResponse("Error getting CA configuration", err), nil
	}
	// Construct enrollment request JSON
	enrollmentRequest := EnrollmentRequest{
		CSR:              csrPem,
		CertificateProfile: configCa.CertProfile,
		EndEntityProfile: configCa.EEProfile,
		CAName: configCa.CAName,
		Username: username,
		Password:  enrollmentcode,
		IncludeChain:  "false",
	}

	body, err := json.Marshal(enrollmentRequest)
	if err != nil {
		return logical.ErrorResponse("Error constructing enrollment request: %v", err), err
	}

	if b.Logger().IsDebug() {
		b.Logger().Debug(fmt.Sprintf("Enrollment request body: %v", string(body)))
	}

	tlsClientConfig, err := getTLSConfig(ctx, req, configCa)
	if err != nil {
		return logical.ErrorResponse("Error retrieving TLS configuration: %v", err), err
	}

	tr := &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: tlsClientConfig,
	}

	client := &http.Client{Transport: tr}
	resp, err := client.Post(configCa.URL+"/certificate/pkcs10enroll", "application/json", bytes.NewReader(body))
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

	var enrollmentResponse EnrollmentResponse
	err = json.Unmarshal(responseBody, &enrollmentResponse)
	if err != nil {
		return logical.ErrorResponse("CA enrollment response could not be parsed: %v", err), err
	}

	var respData map[string]interface{}
	respData = map[string]interface{}{
		"certificate": enrollmentResponse.Certificate,
		"serial_number": enrollmentResponse.SerialNo,
	}
	if len(enrollmentResponse.Certificate) <= 0 {
		return logical.ErrorResponse("Certificate response is empty"), nil
	}
	if len(enrollmentResponse.SerialNo) <= 0 {
		return logical.ErrorResponse("Certificate response serial number is empty"), nil
	}

	// Store issued certificate into Vault, so we can list it in Vault as well, not only on the CA
	// Note that this is not like CA storage, when the plugin is unloaded, these certificates are gone
	storageEntry, err := logical.StorageEntryJSON("issued/"+profileId+"/"+respData["serial_number"].(string), respData)
	if err != nil {
		return logical.ErrorResponse("error creating certificate storage entry"), err
	}
	err = req.Storage.Put(ctx, storageEntry)
	if err != nil {
		return logical.ErrorResponse("could not store certificate"), err
	}
        b.Logger().Info("Added certificate to storage: " + enrollmentResponse.SerialNo)

	return &logical.Response{
		Data: respData,
	}, nil

}

