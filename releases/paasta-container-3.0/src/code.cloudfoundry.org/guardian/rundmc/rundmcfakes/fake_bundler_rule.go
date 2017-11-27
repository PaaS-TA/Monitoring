// Code generated by counterfeiter. DO NOT EDIT.
package rundmcfakes

import (
	"sync"

	"code.cloudfoundry.org/guardian/gardener"
	"code.cloudfoundry.org/guardian/rundmc"
	"code.cloudfoundry.org/guardian/rundmc/goci"
)

type FakeBundlerRule struct {
	ApplyStub        func(bndle goci.Bndl, spec gardener.DesiredContainerSpec, containerDir string) (goci.Bndl, error)
	applyMutex       sync.RWMutex
	applyArgsForCall []struct {
		bndle        goci.Bndl
		spec         gardener.DesiredContainerSpec
		containerDir string
	}
	applyReturns struct {
		result1 goci.Bndl
		result2 error
	}
	applyReturnsOnCall map[int]struct {
		result1 goci.Bndl
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeBundlerRule) Apply(bndle goci.Bndl, spec gardener.DesiredContainerSpec, containerDir string) (goci.Bndl, error) {
	fake.applyMutex.Lock()
	ret, specificReturn := fake.applyReturnsOnCall[len(fake.applyArgsForCall)]
	fake.applyArgsForCall = append(fake.applyArgsForCall, struct {
		bndle        goci.Bndl
		spec         gardener.DesiredContainerSpec
		containerDir string
	}{bndle, spec, containerDir})
	fake.recordInvocation("Apply", []interface{}{bndle, spec, containerDir})
	fake.applyMutex.Unlock()
	if fake.ApplyStub != nil {
		return fake.ApplyStub(bndle, spec, containerDir)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fake.applyReturns.result1, fake.applyReturns.result2
}

func (fake *FakeBundlerRule) ApplyCallCount() int {
	fake.applyMutex.RLock()
	defer fake.applyMutex.RUnlock()
	return len(fake.applyArgsForCall)
}

func (fake *FakeBundlerRule) ApplyArgsForCall(i int) (goci.Bndl, gardener.DesiredContainerSpec, string) {
	fake.applyMutex.RLock()
	defer fake.applyMutex.RUnlock()
	return fake.applyArgsForCall[i].bndle, fake.applyArgsForCall[i].spec, fake.applyArgsForCall[i].containerDir
}

func (fake *FakeBundlerRule) ApplyReturns(result1 goci.Bndl, result2 error) {
	fake.ApplyStub = nil
	fake.applyReturns = struct {
		result1 goci.Bndl
		result2 error
	}{result1, result2}
}

func (fake *FakeBundlerRule) ApplyReturnsOnCall(i int, result1 goci.Bndl, result2 error) {
	fake.ApplyStub = nil
	if fake.applyReturnsOnCall == nil {
		fake.applyReturnsOnCall = make(map[int]struct {
			result1 goci.Bndl
			result2 error
		})
	}
	fake.applyReturnsOnCall[i] = struct {
		result1 goci.Bndl
		result2 error
	}{result1, result2}
}

func (fake *FakeBundlerRule) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.applyMutex.RLock()
	defer fake.applyMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeBundlerRule) recordInvocation(key string, args []interface{}) {
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

var _ rundmc.BundlerRule = new(FakeBundlerRule)