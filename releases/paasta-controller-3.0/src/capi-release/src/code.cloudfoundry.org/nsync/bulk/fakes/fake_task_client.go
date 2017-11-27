// This file was generated by counterfeiter
package fakes

import (
	"net/http"
	"sync"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/nsync/bulk"
	"code.cloudfoundry.org/runtimeschema/cc_messages"
)

type FakeTaskClient struct {
	FailTaskStub        func(logger lager.Logger, taskState *cc_messages.CCTaskState, httpClient *http.Client) error
	failTaskMutex       sync.RWMutex
	failTaskArgsForCall []struct {
		logger     lager.Logger
		taskState  *cc_messages.CCTaskState
		httpClient *http.Client
	}
	failTaskReturns struct {
		result1 error
	}
}

func (fake *FakeTaskClient) FailTask(logger lager.Logger, taskState *cc_messages.CCTaskState, httpClient *http.Client) error {
	fake.failTaskMutex.Lock()
	fake.failTaskArgsForCall = append(fake.failTaskArgsForCall, struct {
		logger     lager.Logger
		taskState  *cc_messages.CCTaskState
		httpClient *http.Client
	}{logger, taskState, httpClient})
	fake.failTaskMutex.Unlock()
	if fake.FailTaskStub != nil {
		return fake.FailTaskStub(logger, taskState, httpClient)
	} else {
		return fake.failTaskReturns.result1
	}
}

func (fake *FakeTaskClient) FailTaskCallCount() int {
	fake.failTaskMutex.RLock()
	defer fake.failTaskMutex.RUnlock()
	return len(fake.failTaskArgsForCall)
}

func (fake *FakeTaskClient) FailTaskArgsForCall(i int) (lager.Logger, *cc_messages.CCTaskState, *http.Client) {
	fake.failTaskMutex.RLock()
	defer fake.failTaskMutex.RUnlock()
	return fake.failTaskArgsForCall[i].logger, fake.failTaskArgsForCall[i].taskState, fake.failTaskArgsForCall[i].httpClient
}

func (fake *FakeTaskClient) FailTaskReturns(result1 error) {
	fake.FailTaskStub = nil
	fake.failTaskReturns = struct {
		result1 error
	}{result1}
}

var _ bulk.TaskClient = new(FakeTaskClient)
