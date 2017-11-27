// Code generated by counterfeiter. DO NOT EDIT.
package kawasakifakes

import (
	"sync"

	"code.cloudfoundry.org/guardian/kawasaki"
)

type FakeIDGenerator struct {
	GenerateStub        func() string
	generateMutex       sync.RWMutex
	generateArgsForCall []struct{}
	generateReturns     struct {
		result1 string
	}
	generateReturnsOnCall map[int]struct {
		result1 string
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeIDGenerator) Generate() string {
	fake.generateMutex.Lock()
	ret, specificReturn := fake.generateReturnsOnCall[len(fake.generateArgsForCall)]
	fake.generateArgsForCall = append(fake.generateArgsForCall, struct{}{})
	fake.recordInvocation("Generate", []interface{}{})
	fake.generateMutex.Unlock()
	if fake.GenerateStub != nil {
		return fake.GenerateStub()
	}
	if specificReturn {
		return ret.result1
	}
	return fake.generateReturns.result1
}

func (fake *FakeIDGenerator) GenerateCallCount() int {
	fake.generateMutex.RLock()
	defer fake.generateMutex.RUnlock()
	return len(fake.generateArgsForCall)
}

func (fake *FakeIDGenerator) GenerateReturns(result1 string) {
	fake.GenerateStub = nil
	fake.generateReturns = struct {
		result1 string
	}{result1}
}

func (fake *FakeIDGenerator) GenerateReturnsOnCall(i int, result1 string) {
	fake.GenerateStub = nil
	if fake.generateReturnsOnCall == nil {
		fake.generateReturnsOnCall = make(map[int]struct {
			result1 string
		})
	}
	fake.generateReturnsOnCall[i] = struct {
		result1 string
	}{result1}
}

func (fake *FakeIDGenerator) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.generateMutex.RLock()
	defer fake.generateMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeIDGenerator) recordInvocation(key string, args []interface{}) {
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

var _ kawasaki.IDGenerator = new(FakeIDGenerator)
