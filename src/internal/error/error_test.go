package custerror_test

import (
	"errors"
	custerror "labs/local-transcoder/internal/error"
	"testing"
)

func TestError_ErrorIsFalse(t *testing.T) {
	returnedError := custerror.ErrorNotFound
	compareTo := custerror.ErrorFailedPrecondition

	result := errors.Is(returnedError, compareTo)

	expectedResult := false
	if result != expectedResult {
		t.Fatalf("result should be %t, got %t", expectedResult, result)
	}

}

func TestError_ErrorIsTrue(t *testing.T) {
	returnedError := custerror.ErrorNotFound
	compareTo := custerror.ErrorNotFound

	result := errors.Is(returnedError, compareTo)

	expectedResult := true
	if result != expectedResult {
		t.Fatalf("result should be %t, got %t", expectedResult, result)
	}

}

func TestError_ErrorIsFalseWrapped(t *testing.T) {
	returnedError := custerror.ErrorNotFound
	compareTo := custerror.FormatAlreadyExists("ErrorIsFalseWrapped custom message")

	result := errors.Is(returnedError, compareTo)

	expectedResult := false
	if result != expectedResult {
		t.Fatalf("result should be %t, got %t", expectedResult, result)
	}

}

func TestError_ErrorIsTrueWrapped(t *testing.T) {
	returnedError := custerror.ErrorNotFound
	compareTo := custerror.FormatNotFound("ErrorIsFalseWrapped custom message")

	result := errors.Is(returnedError, compareTo)

	expectedResult := true
	if result != expectedResult {
		t.Fatalf("result should be %t, got %t", expectedResult, result)
	}

}
