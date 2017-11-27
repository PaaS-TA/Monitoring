// This file was generated by counterfeiter
package eventfakes

import (
	"sync"

	"code.cloudfoundry.org/bbs/events"
	"code.cloudfoundry.org/bbs/models"
)

type FakeHub struct {
	SubscribeStub        func() (events.EventSource, error)
	subscribeMutex       sync.RWMutex
	subscribeArgsForCall []struct{}
	subscribeReturns     struct {
		result1 events.EventSource
		result2 error
	}
	EmitStub        func(models.Event)
	emitMutex       sync.RWMutex
	emitArgsForCall []struct {
		arg1 models.Event
	}
	CloseStub        func() error
	closeMutex       sync.RWMutex
	closeArgsForCall []struct{}
	closeReturns     struct {
		result1 error
	}
	RegisterCallbackStub        func(func(count int))
	registerCallbackMutex       sync.RWMutex
	registerCallbackArgsForCall []struct {
		arg1 func(count int)
	}
	UnregisterCallbackStub        func()
	unregisterCallbackMutex       sync.RWMutex
	unregisterCallbackArgsForCall []struct{}
	invocations                   map[string][][]interface{}
	invocationsMutex              sync.RWMutex
}

func (fake *FakeHub) Subscribe() (events.EventSource, error) {
	fake.subscribeMutex.Lock()
	fake.subscribeArgsForCall = append(fake.subscribeArgsForCall, struct{}{})
	fake.recordInvocation("Subscribe", []interface{}{})
	fake.subscribeMutex.Unlock()
	if fake.SubscribeStub != nil {
		return fake.SubscribeStub()
	} else {
		return fake.subscribeReturns.result1, fake.subscribeReturns.result2
	}
}

func (fake *FakeHub) SubscribeCallCount() int {
	fake.subscribeMutex.RLock()
	defer fake.subscribeMutex.RUnlock()
	return len(fake.subscribeArgsForCall)
}

func (fake *FakeHub) SubscribeReturns(result1 events.EventSource, result2 error) {
	fake.SubscribeStub = nil
	fake.subscribeReturns = struct {
		result1 events.EventSource
		result2 error
	}{result1, result2}
}

func (fake *FakeHub) Emit(arg1 models.Event) {
	fake.emitMutex.Lock()
	fake.emitArgsForCall = append(fake.emitArgsForCall, struct {
		arg1 models.Event
	}{arg1})
	fake.recordInvocation("Emit", []interface{}{arg1})
	fake.emitMutex.Unlock()
	if fake.EmitStub != nil {
		fake.EmitStub(arg1)
	}
}

func (fake *FakeHub) EmitCallCount() int {
	fake.emitMutex.RLock()
	defer fake.emitMutex.RUnlock()
	return len(fake.emitArgsForCall)
}

func (fake *FakeHub) EmitArgsForCall(i int) models.Event {
	fake.emitMutex.RLock()
	defer fake.emitMutex.RUnlock()
	return fake.emitArgsForCall[i].arg1
}

func (fake *FakeHub) Close() error {
	fake.closeMutex.Lock()
	fake.closeArgsForCall = append(fake.closeArgsForCall, struct{}{})
	fake.recordInvocation("Close", []interface{}{})
	fake.closeMutex.Unlock()
	if fake.CloseStub != nil {
		return fake.CloseStub()
	} else {
		return fake.closeReturns.result1
	}
}

func (fake *FakeHub) CloseCallCount() int {
	fake.closeMutex.RLock()
	defer fake.closeMutex.RUnlock()
	return len(fake.closeArgsForCall)
}

func (fake *FakeHub) CloseReturns(result1 error) {
	fake.CloseStub = nil
	fake.closeReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeHub) RegisterCallback(arg1 func(count int)) {
	fake.registerCallbackMutex.Lock()
	fake.registerCallbackArgsForCall = append(fake.registerCallbackArgsForCall, struct {
		arg1 func(count int)
	}{arg1})
	fake.recordInvocation("RegisterCallback", []interface{}{arg1})
	fake.registerCallbackMutex.Unlock()
	if fake.RegisterCallbackStub != nil {
		fake.RegisterCallbackStub(arg1)
	}
}

func (fake *FakeHub) RegisterCallbackCallCount() int {
	fake.registerCallbackMutex.RLock()
	defer fake.registerCallbackMutex.RUnlock()
	return len(fake.registerCallbackArgsForCall)
}

func (fake *FakeHub) RegisterCallbackArgsForCall(i int) func(count int) {
	fake.registerCallbackMutex.RLock()
	defer fake.registerCallbackMutex.RUnlock()
	return fake.registerCallbackArgsForCall[i].arg1
}

func (fake *FakeHub) UnregisterCallback() {
	fake.unregisterCallbackMutex.Lock()
	fake.unregisterCallbackArgsForCall = append(fake.unregisterCallbackArgsForCall, struct{}{})
	fake.recordInvocation("UnregisterCallback", []interface{}{})
	fake.unregisterCallbackMutex.Unlock()
	if fake.UnregisterCallbackStub != nil {
		fake.UnregisterCallbackStub()
	}
}

func (fake *FakeHub) UnregisterCallbackCallCount() int {
	fake.unregisterCallbackMutex.RLock()
	defer fake.unregisterCallbackMutex.RUnlock()
	return len(fake.unregisterCallbackArgsForCall)
}

func (fake *FakeHub) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.subscribeMutex.RLock()
	defer fake.subscribeMutex.RUnlock()
	fake.emitMutex.RLock()
	defer fake.emitMutex.RUnlock()
	fake.closeMutex.RLock()
	defer fake.closeMutex.RUnlock()
	fake.registerCallbackMutex.RLock()
	defer fake.registerCallbackMutex.RUnlock()
	fake.unregisterCallbackMutex.RLock()
	defer fake.unregisterCallbackMutex.RUnlock()
	return fake.invocations
}

func (fake *FakeHub) recordInvocation(key string, args []interface{}) {
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

var _ events.Hub = new(FakeHub)
