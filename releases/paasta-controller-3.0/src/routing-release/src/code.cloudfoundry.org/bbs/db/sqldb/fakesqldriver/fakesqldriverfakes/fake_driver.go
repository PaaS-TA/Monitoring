// This file was generated by counterfeiter
package fakesqldriverfakes

import (
	"database/sql/driver"
	"sync"

	"code.cloudfoundry.org/bbs/db/sqldb/fakesqldriver"
)

type FakeDriver struct {
	OpenStub        func(name string) (driver.Conn, error)
	openMutex       sync.RWMutex
	openArgsForCall []struct {
		name string
	}
	openReturns struct {
		result1 driver.Conn
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeDriver) Open(name string) (driver.Conn, error) {
	fake.openMutex.Lock()
	fake.openArgsForCall = append(fake.openArgsForCall, struct {
		name string
	}{name})
	fake.recordInvocation("Open", []interface{}{name})
	fake.openMutex.Unlock()
	if fake.OpenStub != nil {
		return fake.OpenStub(name)
	} else {
		return fake.openReturns.result1, fake.openReturns.result2
	}
}

func (fake *FakeDriver) OpenCallCount() int {
	fake.openMutex.RLock()
	defer fake.openMutex.RUnlock()
	return len(fake.openArgsForCall)
}

func (fake *FakeDriver) OpenArgsForCall(i int) string {
	fake.openMutex.RLock()
	defer fake.openMutex.RUnlock()
	return fake.openArgsForCall[i].name
}

func (fake *FakeDriver) OpenReturns(result1 driver.Conn, result2 error) {
	fake.OpenStub = nil
	fake.openReturns = struct {
		result1 driver.Conn
		result2 error
	}{result1, result2}
}

func (fake *FakeDriver) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.openMutex.RLock()
	defer fake.openMutex.RUnlock()
	return fake.invocations
}

func (fake *FakeDriver) recordInvocation(key string, args []interface{}) {
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

var _ fakesqldriver.Driver = new(FakeDriver)
