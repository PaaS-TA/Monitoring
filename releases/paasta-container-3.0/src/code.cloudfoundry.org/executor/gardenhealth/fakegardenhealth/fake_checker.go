// Code generated by counterfeiter. DO NOT EDIT.
package fakegardenhealth

import (
	"sync"

	"code.cloudfoundry.org/executor/gardenhealth"
	"code.cloudfoundry.org/lager"
)

type FakeChecker struct {
	HealthcheckStub        func(lager.Logger) error
	healthcheckMutex       sync.RWMutex
	healthcheckArgsForCall []struct {
		arg1 lager.Logger
	}
	healthcheckReturns struct {
		result1 error
	}
	healthcheckReturnsOnCall map[int]struct {
		result1 error
	}
	CancelStub        func(lager.Logger)
	cancelMutex       sync.RWMutex
	cancelArgsForCall []struct {
		arg1 lager.Logger
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeChecker) Healthcheck(arg1 lager.Logger) error {
	fake.healthcheckMutex.Lock()
	ret, specificReturn := fake.healthcheckReturnsOnCall[len(fake.healthcheckArgsForCall)]
	fake.healthcheckArgsForCall = append(fake.healthcheckArgsForCall, struct {
		arg1 lager.Logger
	}{arg1})
	fake.recordInvocation("Healthcheck", []interface{}{arg1})
	fake.healthcheckMutex.Unlock()
	if fake.HealthcheckStub != nil {
		return fake.HealthcheckStub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.healthcheckReturns.result1
}

func (fake *FakeChecker) HealthcheckCallCount() int {
	fake.healthcheckMutex.RLock()
	defer fake.healthcheckMutex.RUnlock()
	return len(fake.healthcheckArgsForCall)
}

func (fake *FakeChecker) HealthcheckArgsForCall(i int) lager.Logger {
	fake.healthcheckMutex.RLock()
	defer fake.healthcheckMutex.RUnlock()
	return fake.healthcheckArgsForCall[i].arg1
}

func (fake *FakeChecker) HealthcheckReturns(result1 error) {
	fake.HealthcheckStub = nil
	fake.healthcheckReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeChecker) HealthcheckReturnsOnCall(i int, result1 error) {
	fake.HealthcheckStub = nil
	if fake.healthcheckReturnsOnCall == nil {
		fake.healthcheckReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.healthcheckReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeChecker) Cancel(arg1 lager.Logger) {
	fake.cancelMutex.Lock()
	fake.cancelArgsForCall = append(fake.cancelArgsForCall, struct {
		arg1 lager.Logger
	}{arg1})
	fake.recordInvocation("Cancel", []interface{}{arg1})
	fake.cancelMutex.Unlock()
	if fake.CancelStub != nil {
		fake.CancelStub(arg1)
	}
}

func (fake *FakeChecker) CancelCallCount() int {
	fake.cancelMutex.RLock()
	defer fake.cancelMutex.RUnlock()
	return len(fake.cancelArgsForCall)
}

func (fake *FakeChecker) CancelArgsForCall(i int) lager.Logger {
	fake.cancelMutex.RLock()
	defer fake.cancelMutex.RUnlock()
	return fake.cancelArgsForCall[i].arg1
}

func (fake *FakeChecker) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.healthcheckMutex.RLock()
	defer fake.healthcheckMutex.RUnlock()
	fake.cancelMutex.RLock()
	defer fake.cancelMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeChecker) recordInvocation(key string, args []interface{}) {
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

var _ gardenhealth.Checker = new(FakeChecker)
