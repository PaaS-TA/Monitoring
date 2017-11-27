// Code generated by counterfeiter. DO NOT EDIT.
package bundlerulesfakes

import (
	"sync"

	"code.cloudfoundry.org/guardian/rundmc/bundlerules"
)

type FakeChowner struct {
	ChownStub        func(path string, uid, gid int) error
	chownMutex       sync.RWMutex
	chownArgsForCall []struct {
		path string
		uid  int
		gid  int
	}
	chownReturns struct {
		result1 error
	}
	chownReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeChowner) Chown(path string, uid int, gid int) error {
	fake.chownMutex.Lock()
	ret, specificReturn := fake.chownReturnsOnCall[len(fake.chownArgsForCall)]
	fake.chownArgsForCall = append(fake.chownArgsForCall, struct {
		path string
		uid  int
		gid  int
	}{path, uid, gid})
	fake.recordInvocation("Chown", []interface{}{path, uid, gid})
	fake.chownMutex.Unlock()
	if fake.ChownStub != nil {
		return fake.ChownStub(path, uid, gid)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.chownReturns.result1
}

func (fake *FakeChowner) ChownCallCount() int {
	fake.chownMutex.RLock()
	defer fake.chownMutex.RUnlock()
	return len(fake.chownArgsForCall)
}

func (fake *FakeChowner) ChownArgsForCall(i int) (string, int, int) {
	fake.chownMutex.RLock()
	defer fake.chownMutex.RUnlock()
	return fake.chownArgsForCall[i].path, fake.chownArgsForCall[i].uid, fake.chownArgsForCall[i].gid
}

func (fake *FakeChowner) ChownReturns(result1 error) {
	fake.ChownStub = nil
	fake.chownReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeChowner) ChownReturnsOnCall(i int, result1 error) {
	fake.ChownStub = nil
	if fake.chownReturnsOnCall == nil {
		fake.chownReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.chownReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeChowner) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.chownMutex.RLock()
	defer fake.chownMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeChowner) recordInvocation(key string, args []interface{}) {
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

var _ bundlerules.Chowner = new(FakeChowner)
