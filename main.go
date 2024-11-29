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
