// This file was generated by counterfeiter
package fake_evacuation_context

import (
	"sync"

	"code.cloudfoundry.org/rep/evacuation/evacuation_context"
)

type FakeEvacuationNotifier struct {
	EvacuateNotifyStub        func() <-chan struct{}
	evacuateNotifyMutex       sync.RWMutex
	evacuateNotifyArgsForCall []struct{}
	evacuateNotifyReturns     struct {
		result1 <-chan struct{}
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeEvacuationNotifier) EvacuateNotify() <-chan struct{} {
	fake.evacuateNotifyMutex.Lock()
	fake.evacuateNotifyArgsForCall = append(fake.evacuateNotifyArgsForCall, struct{}{})
	fake.recordInvocation("EvacuateNotify", []interface{}{})
	fake.evacuateNotifyMutex.Unlock()
	if fake.EvacuateNotifyStub != nil {
		return fake.EvacuateNotifyStub()
	} else {
		return fake.evacuateNotifyReturns.result1
	}
}

func (fake *FakeEvacuationNotifier) EvacuateNotifyCallCount() int {
	fake.evacuateNotifyMutex.RLock()
	defer fake.evacuateNotifyMutex.RUnlock()
	return len(fake.evacuateNotifyArgsForCall)
}

func (fake *FakeEvacuationNotifier) EvacuateNotifyReturns(result1 <-chan struct{}) {
	fake.EvacuateNotifyStub = nil
	fake.evacuateNotifyReturns = struct {
		result1 <-chan struct{}
	}{result1}
}

func (fake *FakeEvacuationNotifier) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.evacuateNotifyMutex.RLock()
	defer fake.evacuateNotifyMutex.RUnlock()
	return fake.invocations
}

func (fake *FakeEvacuationNotifier) recordInvocation(key string, args []interface{}) {
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

var _ evacuation_context.EvacuationNotifier = new(FakeEvacuationNotifier)
