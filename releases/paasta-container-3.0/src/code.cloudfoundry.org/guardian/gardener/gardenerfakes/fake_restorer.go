// Code generated by counterfeiter. DO NOT EDIT.
package gardenerfakes

import (
	"sync"

	"code.cloudfoundry.org/guardian/gardener"
	"code.cloudfoundry.org/lager"
)

type FakeRestorer struct {
	RestoreStub        func(logger lager.Logger, handles []string) []string
	restoreMutex       sync.RWMutex
	restoreArgsForCall []struct {
		logger  lager.Logger
		handles []string
	}
	restoreReturns struct {
		result1 []string
	}
	restoreReturnsOnCall map[int]struct {
		result1 []string
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeRestorer) Restore(logger lager.Logger, handles []string) []string {
	var handlesCopy []string
	if handles != nil {
		handlesCopy = make([]string, len(handles))
		copy(handlesCopy, handles)
	}
	fake.restoreMutex.Lock()
	ret, specificReturn := fake.restoreReturnsOnCall[len(fake.restoreArgsForCall)]
	fake.restoreArgsForCall = append(fake.restoreArgsForCall, struct {
		logger  lager.Logger
		handles []string
	}{logger, handlesCopy})
	fake.recordInvocation("Restore", []interface{}{logger, handlesCopy})
	fake.restoreMutex.Unlock()
	if fake.RestoreStub != nil {
		return fake.RestoreStub(logger, handles)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.restoreReturns.result1
}

func (fake *FakeRestorer) RestoreCallCount() int {
	fake.restoreMutex.RLock()
	defer fake.restoreMutex.RUnlock()
	return len(fake.restoreArgsForCall)
}

func (fake *FakeRestorer) RestoreArgsForCall(i int) (lager.Logger, []string) {
	fake.restoreMutex.RLock()
	defer fake.restoreMutex.RUnlock()
	return fake.restoreArgsForCall[i].logger, fake.restoreArgsForCall[i].handles
}

func (fake *FakeRestorer) RestoreReturns(result1 []string) {
	fake.RestoreStub = nil
	fake.restoreReturns = struct {
		result1 []string
	}{result1}
}

func (fake *FakeRestorer) RestoreReturnsOnCall(i int, result1 []string) {
	fake.RestoreStub = nil
	if fake.restoreReturnsOnCall == nil {
		fake.restoreReturnsOnCall = make(map[int]struct {
			result1 []string
		})
	}
	fake.restoreReturnsOnCall[i] = struct {
		result1 []string
	}{result1}
}

func (fake *FakeRestorer) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.restoreMutex.RLock()
	defer fake.restoreMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeRestorer) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ gardener.Restorer = new(FakeRestorer)
