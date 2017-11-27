// This file was generated by counterfeiter
package maintainfakes

import (
	"sync"
	"time"

	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/rep/maintain"
	"github.com/tedsuo/ifrit"
)

type FakeCellPresenceClient struct {
	NewCellPresenceRunnerStub        func(logger lager.Logger, cellPresence *models.CellPresence, retryInterval, lockTTL time.Duration) ifrit.Runner
	newCellPresenceRunnerMutex       sync.RWMutex
	newCellPresenceRunnerArgsForCall []struct {
		logger        lager.Logger
		cellPresence  *models.CellPresence
		retryInterval time.Duration
		lockTTL       time.Duration
	}
	newCellPresenceRunnerReturns struct {
		result1 ifrit.Runner
	}
	newCellPresenceRunnerReturnsOnCall map[int]struct {
		result1 ifrit.Runner
	}
	CellByIdStub        func(logger lager.Logger, cellId string) (*models.CellPresence, error)
	cellByIdMutex       sync.RWMutex
	cellByIdArgsForCall []struct {
		logger lager.Logger
		cellId string
	}
	cellByIdReturns struct {
		result1 *models.CellPresence
		result2 error
	}
	cellByIdReturnsOnCall map[int]struct {
		result1 *models.CellPresence
		result2 error
	}
	CellsStub        func(logger lager.Logger) (models.CellSet, error)
	cellsMutex       sync.RWMutex
	cellsArgsForCall []struct {
		logger lager.Logger
	}
	cellsReturns struct {
		result1 models.CellSet
		result2 error
	}
	cellsReturnsOnCall map[int]struct {
		result1 models.CellSet
		result2 error
	}
	CellEventsStub        func(logger lager.Logger) <-chan models.CellEvent
	cellEventsMutex       sync.RWMutex
	cellEventsArgsForCall []struct {
		logger lager.Logger
	}
	cellEventsReturns struct {
		result1 <-chan models.CellEvent
	}
	cellEventsReturnsOnCall map[int]struct {
		result1 <-chan models.CellEvent
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeCellPresenceClient) NewCellPresenceRunner(logger lager.Logger, cellPresence *models.CellPresence, retryInterval time.Duration, lockTTL time.Duration) ifrit.Runner {
	fake.newCellPresenceRunnerMutex.Lock()
	ret, specificReturn := fake.newCellPresenceRunnerReturnsOnCall[len(fake.newCellPresenceRunnerArgsForCall)]
	fake.newCellPresenceRunnerArgsForCall = append(fake.newCellPresenceRunnerArgsForCall, struct {
		logger        lager.Logger
		cellPresence  *models.CellPresence
		retryInterval time.Duration
		lockTTL       time.Duration
	}{logger, cellPresence, retryInterval, lockTTL})
	fake.recordInvocation("NewCellPresenceRunner", []interface{}{logger, cellPresence, retryInterval, lockTTL})
	fake.newCellPresenceRunnerMutex.Unlock()
	if fake.NewCellPresenceRunnerStub != nil {
		return fake.NewCellPresenceRunnerStub(logger, cellPresence, retryInterval, lockTTL)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.newCellPresenceRunnerReturns.result1
}

func (fake *FakeCellPresenceClient) NewCellPresenceRunnerCallCount() int {
	fake.newCellPresenceRunnerMutex.RLock()
	defer fake.newCellPresenceRunnerMutex.RUnlock()
	return len(fake.newCellPresenceRunnerArgsForCall)
}

func (fake *FakeCellPresenceClient) NewCellPresenceRunnerArgsForCall(i int) (lager.Logger, *models.CellPresence, time.Duration, time.Duration) {
	fake.newCellPresenceRunnerMutex.RLock()
	defer fake.newCellPresenceRunnerMutex.RUnlock()
	return fake.newCellPresenceRunnerArgsForCall[i].logger, fake.newCellPresenceRunnerArgsForCall[i].cellPresence, fake.newCellPresenceRunnerArgsForCall[i].retryInterval, fake.newCellPresenceRunnerArgsForCall[i].lockTTL
}

func (fake *FakeCellPresenceClient) NewCellPresenceRunnerReturns(result1 ifrit.Runner) {
	fake.NewCellPresenceRunnerStub = nil
	fake.newCellPresenceRunnerReturns = struct {
		result1 ifrit.Runner
	}{result1}
}

func (fake *FakeCellPresenceClient) NewCellPresenceRunnerReturnsOnCall(i int, result1 ifrit.Runner) {
	fake.NewCellPresenceRunnerStub = nil
	if fake.newCellPresenceRunnerReturnsOnCall == nil {
		fake.newCellPresenceRunnerReturnsOnCall = make(map[int]struct {
			result1 ifrit.Runner
		})
	}
	fake.newCellPresenceRunnerReturnsOnCall[i] = struct {
		result1 ifrit.Runner
	}{result1}
}

func (fake *FakeCellPresenceClient) CellById(logger lager.Logger, cellId string) (*models.CellPresence, error) {
	fake.cellByIdMutex.Lock()
	ret, specificReturn := fake.cellByIdReturnsOnCall[len(fake.cellByIdArgsForCall)]
	fake.cellByIdArgsForCall = append(fake.cellByIdArgsForCall, struct {
		logger lager.Logger
		cellId string
	}{logger, cellId})
	fake.recordInvocation("CellById", []interface{}{logger, cellId})
	fake.cellByIdMutex.Unlock()
	if fake.CellByIdStub != nil {
		return fake.CellByIdStub(logger, cellId)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fake.cellByIdReturns.result1, fake.cellByIdReturns.result2
}

func (fake *FakeCellPresenceClient) CellByIdCallCount() int {
	fake.cellByIdMutex.RLock()
	defer fake.cellByIdMutex.RUnlock()
	return len(fake.cellByIdArgsForCall)
}

func (fake *FakeCellPresenceClient) CellByIdArgsForCall(i int) (lager.Logger, string) {
	fake.cellByIdMutex.RLock()
	defer fake.cellByIdMutex.RUnlock()
	return fake.cellByIdArgsForCall[i].logger, fake.cellByIdArgsForCall[i].cellId
}

func (fake *FakeCellPresenceClient) CellByIdReturns(result1 *models.CellPresence, result2 error) {
	fake.CellByIdStub = nil
	fake.cellByIdReturns = struct {
		result1 *models.CellPresence
		result2 error
	}{result1, result2}
}

func (fake *FakeCellPresenceClient) CellByIdReturnsOnCall(i int, result1 *models.CellPresence, result2 error) {
	fake.CellByIdStub = nil
	if fake.cellByIdReturnsOnCall == nil {
		fake.cellByIdReturnsOnCall = make(map[int]struct {
			result1 *models.CellPresence
			result2 error
		})
	}
	fake.cellByIdReturnsOnCall[i] = struct {
		result1 *models.CellPresence
		result2 error
	}{result1, result2}
}

func (fake *FakeCellPresenceClient) Cells(logger lager.Logger) (models.CellSet, error) {
	fake.cellsMutex.Lock()
	ret, specificReturn := fake.cellsReturnsOnCall[len(fake.cellsArgsForCall)]
	fake.cellsArgsForCall = append(fake.cellsArgsForCall, struct {
		logger lager.Logger
	}{logger})
	fake.recordInvocation("Cells", []interface{}{logger})
	fake.cellsMutex.Unlock()
	if fake.CellsStub != nil {
		return fake.CellsStub(logger)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fake.cellsReturns.result1, fake.cellsReturns.result2
}

func (fake *FakeCellPresenceClient) CellsCallCount() int {
	fake.cellsMutex.RLock()
	defer fake.cellsMutex.RUnlock()
	return len(fake.cellsArgsForCall)
}

func (fake *FakeCellPresenceClient) CellsArgsForCall(i int) lager.Logger {
	fake.cellsMutex.RLock()
	defer fake.cellsMutex.RUnlock()
	return fake.cellsArgsForCall[i].logger
}

func (fake *FakeCellPresenceClient) CellsReturns(result1 models.CellSet, result2 error) {
	fake.CellsStub = nil
	fake.cellsReturns = struct {
		result1 models.CellSet
		result2 error
	}{result1, result2}
}

func (fake *FakeCellPresenceClient) CellsReturnsOnCall(i int, result1 models.CellSet, result2 error) {
	fake.CellsStub = nil
	if fake.cellsReturnsOnCall == nil {
		fake.cellsReturnsOnCall = make(map[int]struct {
			result1 models.CellSet
			result2 error
		})
	}
	fake.cellsReturnsOnCall[i] = struct {
		result1 models.CellSet
		result2 error
	}{result1, result2}
}

func (fake *FakeCellPresenceClient) CellEvents(logger lager.Logger) <-chan models.CellEvent {
	fake.cellEventsMutex.Lock()
	ret, specificReturn := fake.cellEventsReturnsOnCall[len(fake.cellEventsArgsForCall)]
	fake.cellEventsArgsForCall = append(fake.cellEventsArgsForCall, struct {
		logger lager.Logger
	}{logger})
	fake.recordInvocation("CellEvents", []interface{}{logger})
	fake.cellEventsMutex.Unlock()
	if fake.CellEventsStub != nil {
		return fake.CellEventsStub(logger)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.cellEventsReturns.result1
}

func (fake *FakeCellPresenceClient) CellEventsCallCount() int {
	fake.cellEventsMutex.RLock()
	defer fake.cellEventsMutex.RUnlock()
	return len(fake.cellEventsArgsForCall)
}

func (fake *FakeCellPresenceClient) CellEventsArgsForCall(i int) lager.Logger {
	fake.cellEventsMutex.RLock()
	defer fake.cellEventsMutex.RUnlock()
	return fake.cellEventsArgsForCall[i].logger
}

func (fake *FakeCellPresenceClient) CellEventsReturns(result1 <-chan models.CellEvent) {
	fake.CellEventsStub = nil
	fake.cellEventsReturns = struct {
		result1 <-chan models.CellEvent
	}{result1}
}

func (fake *FakeCellPresenceClient) CellEventsReturnsOnCall(i int, result1 <-chan models.CellEvent) {
	fake.CellEventsStub = nil
	if fake.cellEventsReturnsOnCall == nil {
		fake.cellEventsReturnsOnCall = make(map[int]struct {
			result1 <-chan models.CellEvent
		})
	}
	fake.cellEventsReturnsOnCall[i] = struct {
		result1 <-chan models.CellEvent
	}{result1}
}

func (fake *FakeCellPresenceClient) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.newCellPresenceRunnerMutex.RLock()
	defer fake.newCellPresenceRunnerMutex.RUnlock()
	fake.cellByIdMutex.RLock()
	defer fake.cellByIdMutex.RUnlock()
	fake.cellsMutex.RLock()
	defer fake.cellsMutex.RUnlock()
	fake.cellEventsMutex.RLock()
	defer fake.cellEventsMutex.RUnlock()
	return fake.invocations
}

func (fake *FakeCellPresenceClient) recordInvocation(key string, args []interface{}) {
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

var _ maintain.CellPresenceClient = new(FakeCellPresenceClient)