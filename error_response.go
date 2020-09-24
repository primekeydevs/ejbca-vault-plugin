/*
 * Copyright (c) 2019 Entrust Datacard Corporation.
 * All rights reserved.
 */
/*
 * Modified by PrimeKey Solutions AB, Copyright (c) 2020
 */

package main

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
)

type ErrorResponse struct {
	Code      int `json:"error_code"`
	Message   string   `json:"error_message"`
}

func CheckForError(b *backend, body []byte, statusCode int) error {
	if (statusCode != 200 && statusCode != 201) {

		if b.Logger().IsDebug() {
			if statusCode >= 400 {
				b.Logger().Debug(fmt.Sprintf("Received failure response code: %d", statusCode))
			} else {
				b.Logger().Debug(fmt.Sprintf("Received response code: %d", statusCode))
			}
		}

		var errorResponse ErrorResponse
		err := json.Unmarshal(body, &errorResponse)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("CA error response could not be parsed (%d)", statusCode))
		}
		return errors.New(fmt.Sprintf("Error from server: %s (%d)", errorResponse.Message, statusCode))
	}

	return nil
}
