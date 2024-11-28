package main

import (
	"errors"
	"fmt"
	"github.com/aws/smithy-go"
	"log"
	"strings"
)

// imdsRequestError to provide the caller on the request status
type imdsRequestError struct {
	requestKey string
	err        error
	code       string            // Added to support SDK V2 APIError interface
	fault      smithy.ErrorFault // Added to support SDK V2 APIError interface
}

func (e *imdsRequestError) Error() string {
	return fmt.Sprintf("failed to retrieve %s from instance metadata %v", e.requestKey, e.err)
}

func (e *imdsRequestError) Unwrap() error {
	return e.err
}

// Implement smithy.APIError interface
func (e *imdsRequestError) ErrorCode() string {
	var apiErr smithy.APIError
	if errors.As(e.err, &apiErr) {
		return apiErr.ErrorCode()
	}
	return e.code
}

func (e *imdsRequestError) ErrorMessage() string {
	return e.Error()
}

func (e *imdsRequestError) ErrorFault() smithy.ErrorFault {
	var apiErr smithy.APIError
	if errors.As(e.err, &apiErr) {
		return apiErr.ErrorFault()
	}
	return e.fault
}

// Constructor for imdsRequestError
func newIMDSRequestError(requestKey string, err error) *imdsRequestError {
	return &imdsRequestError{
		requestKey: requestKey,
		err:        err,
		code:       "IMDSRequestError",
		fault:      smithy.FaultUnknown,
	}
}

// Simulated function that returns a smithy.OperationError
func simulateEC2IMDSCall() error {
	return &smithy.OperationError{
		ServiceID:     "ec2imds",
		OperationName: "GetMetadata",
		Err:           fmt.Errorf("http response error StatusCode: 404, request to EC2 IMDS failed"),
	}
}

// IsNotFound checks if the error indicates a "not found" condition
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}

	// Check for OperationError
	var oe *smithy.OperationError
	if errors.As(err, &oe) {
		return strings.Contains(oe.Error(), "StatusCode: 404")
	}

	// Check for any APIError (including imdsRequestError)
	var ae smithy.APIError
	if errors.As(err, &ae) {
		if imdsErr, ok := ae.(*imdsRequestError); ok {
			return IsNotFound(imdsErr.err)
		}
		return ae.ErrorCode() == "NotFound"
	}

	return false
}

// Function that makes use of the IMDS call and handles errors
func getMetadata() error {
	// Make the IMDS call
	err := simulateEC2IMDSCall()
	if err != nil {
		// Wrap the error in imdsRequestError
		imdsErr := newIMDSRequestError("metadata", err)

		// Now try to handle the wrapped error
		var checkErr *imdsRequestError
		if errors.As(imdsErr, &checkErr) {
			if IsNotFound(checkErr.err) {
				log.Printf("Not found error detected")
				return nil
			}
			log.Printf("Other IMDS error: %v", checkErr)
			return checkErr.err
		}

		return err
	}
	return nil
}

func main() {
	err := getMetadata()
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		log.Printf("Success or not found")
	}
}
