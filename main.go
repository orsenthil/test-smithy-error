package main

import (
	"errors"
	"fmt"
	"github.com/aws/smithy-go"
	"log"
	"strings"
)

// Define the interface that represents what we consider an IMDS error
type IMDSError interface {
	error
	ErrorContainer
}

type ErrorContainer interface {
	Error() string
	// Add any other common methods both types share
}

// imdsRequestError implementing the IMDSError interface
type imdsRequestError struct {
	requestKey string
	err        error
	code       string
	fault      smithy.ErrorFault
}

func (e *imdsRequestError) Error() string {
	return fmt.Sprintf("failed to retrieve %s from instance metadata %v", e.requestKey, e.err)
}

func (e *imdsRequestError) Unwrap() error {
	return e.err
}

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

	// Check for any APIError
	var ae smithy.APIError
	if errors.As(err, &ae) {
		return ae.ErrorCode() == "NotFound"
	}

	return false
}

// Simulated function that returns a smithy.OperationError
func simulateEC2IMDSCall() error {
	return &smithy.OperationError{
		ServiceID:     "ec2imds",
		OperationName: "GetMetadata",
		Err:           fmt.Errorf("http response error StatusCode: 404, request to EC2 IMDS failed"),
	}
}

// Function that makes use of the IMDS call and handles errors
func getMetadata() error {
	err := simulateEC2IMDSCall()
	if err != nil {
		log.Printf("Original error type: %T", err)

		var imdsErr IMDSError
		var oe *smithy.OperationError
		if errors.As(err, &oe) {
			log.Printf("Recognized as OperationError interface")
		}

		if errors.As(err, &imdsErr) {
			log.Printf("Recognized as IMDSError interface")
			if IsNotFound(err) {
				log.Printf("Not found error detected")
				return nil
			}
			log.Printf("Other IMDS error: %v", imdsErr)
			return err
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
