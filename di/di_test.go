package di_test

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/seinra/axiom/di"
)

type ledgerService struct {
	Balance int
}

type auditService struct {
	Trail string
}

var errBoomFailure = errors.New("di test: boom")

const hammerWorkerCount = 32

func TestProvideInvokeRoundTrip(t *testing.T) {
	injector := di.New()
	var builds atomic.Int64
	provideFailure := di.Provide(injector, "ledger", func(active *di.Injector) (ledgerService, error) {
		builds.Add(1)
		return ledgerService{Balance: 42}, nil
	})
	if provideFailure != nil {
		t.Fatalf("Provide failure = %v, want nil", provideFailure)
	}
	fetched, invokeFailure := di.Invoke[ledgerService](injector, "ledger")
	if invokeFailure != nil {
		t.Fatalf("Invoke failure = %v, want nil", invokeFailure)
	}
	if fetched.Balance != 42 {
		t.Fatalf("Invoke Balance = %d, want 42", fetched.Balance)
	}
	mustFetched := di.MustInvoke[ledgerService](injector, "ledger")
	if mustFetched.Balance != 42 {
		t.Fatalf("MustInvoke Balance = %d, want 42", mustFetched.Balance)
	}
	if builds.Load() != 1 {
		t.Fatalf("constructor calls = %d, want 1", builds.Load())
	}
}

func TestConstructorRunsOnceUnderHammer(t *testing.T) {
	injector := di.New()
	var builds atomic.Int64
	provideFailure := di.Provide(injector, "hammer", func(active *di.Injector) (ledgerService, error) {
		builds.Add(1)
		return ledgerService{Balance: 7}, nil
	})
	if provideFailure != nil {
		t.Fatalf("Provide failure = %v, want nil", provideFailure)
	}
	outcomes := make([]ledgerService, hammerWorkerCount)
	failures := make([]error, hammerWorkerCount)
	var waitGroup sync.WaitGroup
	startSignal := make(chan struct{})
	for workerIdx := 0; workerIdx < hammerWorkerCount; workerIdx++ {
		waitGroup.Add(1)
		go func(workerSlot int) {
			defer waitGroup.Done()
			<-startSignal
			fetched, invokeFailure := di.Invoke[ledgerService](injector, "hammer")
			outcomes[workerSlot] = fetched
			failures[workerSlot] = invokeFailure
		}(workerIdx)
	}
	close(startSignal)
	waitGroup.Wait()
	for workerIdx := 0; workerIdx < hammerWorkerCount; workerIdx++ {
		if failures[workerIdx] != nil {
			t.Fatalf("worker %d failure = %v, want nil", workerIdx, failures[workerIdx])
		}
		if outcomes[workerIdx].Balance != 7 {
			t.Fatalf("worker %d Balance = %d, want 7", workerIdx, outcomes[workerIdx].Balance)
		}
	}
	if builds.Load() != 1 {
		t.Fatalf("constructor calls = %d, want 1", builds.Load())
	}
}

func TestMissingNameReportsNotFound(t *testing.T) {
	injector := di.New()
	_, invokeFailure := di.Invoke[ledgerService](injector, "missing")
	if invokeFailure == nil {
		t.Fatal("Invoke failure = nil, want ErrServiceNotFound")
	}
	if !errors.Is(invokeFailure, di.ErrServiceNotFound) {
		t.Fatalf("Invoke failure = %v, want ErrServiceNotFound", invokeFailure)
	}
}

func TestDuplicateProvideReportsDuplicate(t *testing.T) {
	injector := di.New()
	firstFailure := di.Provide(injector, "ledger", func(active *di.Injector) (ledgerService, error) {
		return ledgerService{Balance: 1}, nil
	})
	if firstFailure != nil {
		t.Fatalf("first Provide failure = %v, want nil", firstFailure)
	}
	secondFailure := di.Provide(injector, "ledger", func(active *di.Injector) (ledgerService, error) {
		return ledgerService{Balance: 2}, nil
	})
	if secondFailure == nil {
		t.Fatal("second Provide failure = nil, want ErrDuplicateService")
	}
	if !errors.Is(secondFailure, di.ErrDuplicateService) {
		t.Fatalf("second Provide failure = %v, want ErrDuplicateService", secondFailure)
	}
	fetched, invokeFailure := di.Invoke[ledgerService](injector, "ledger")
	if invokeFailure != nil {
		t.Fatalf("Invoke failure = %v, want nil", invokeFailure)
	}
	if fetched.Balance != 1 {
		t.Fatalf("Invoke Balance = %d, want 1 (first registration wins)", fetched.Balance)
	}
}

func TestFailingConstructorRetriesUntilSuccess(t *testing.T) {
	injector := di.New()
	var builds atomic.Int64
	provideFailure := di.Provide(injector, "flaky", func(active *di.Injector) (ledgerService, error) {
		attempt := builds.Add(1)
		if attempt <= 2 {
			return ledgerService{}, errBoomFailure
		}
		return ledgerService{Balance: 99}, nil
	})
	if provideFailure != nil {
		t.Fatalf("Provide failure = %v, want nil", provideFailure)
	}
	_, firstFailure := di.Invoke[ledgerService](injector, "flaky")
	if !errors.Is(firstFailure, errBoomFailure) {
		t.Fatalf("first Invoke failure = %v, want boom", firstFailure)
	}
	_, secondFailure := di.Invoke[ledgerService](injector, "flaky")
	if !errors.Is(secondFailure, errBoomFailure) {
		t.Fatalf("second Invoke failure = %v, want boom", secondFailure)
	}
	fetched, thirdFailure := di.Invoke[ledgerService](injector, "flaky")
	if thirdFailure != nil {
		t.Fatalf("third Invoke failure = %v, want nil", thirdFailure)
	}
	if fetched.Balance != 99 {
		t.Fatalf("third Invoke Balance = %d, want 99", fetched.Balance)
	}
	cached, fourthFailure := di.Invoke[ledgerService](injector, "flaky")
	if fourthFailure != nil {
		t.Fatalf("fourth Invoke failure = %v, want nil", fourthFailure)
	}
	if cached.Balance != 99 {
		t.Fatalf("fourth Invoke Balance = %d, want 99", cached.Balance)
	}
	if builds.Load() != 3 {
		t.Fatalf("constructor calls = %d, want 3 (two failures then cached success)", builds.Load())
	}
}

func TestProvideIsLazyUntilFirstInvoke(t *testing.T) {
	injector := di.New()
	var builds atomic.Int64
	provideFailure := di.Provide(injector, "lazy", func(active *di.Injector) (ledgerService, error) {
		builds.Add(1)
		return ledgerService{Balance: 5}, nil
	})
	if provideFailure != nil {
		t.Fatalf("Provide failure = %v, want nil", provideFailure)
	}
	if builds.Load() != 0 {
		t.Fatalf("constructor calls before Invoke = %d, want 0", builds.Load())
	}
	_, invokeFailure := di.Invoke[ledgerService](injector, "lazy")
	if invokeFailure != nil {
		t.Fatalf("Invoke failure = %v, want nil", invokeFailure)
	}
	if builds.Load() != 1 {
		t.Fatalf("constructor calls after Invoke = %d, want 1", builds.Load())
	}
}

func TestMustInvokePanicsOnMissing(t *testing.T) {
	injector := di.New()
	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("MustInvoke did not panic on missing service, want panic")
		}
	}()
	_ = di.MustInvoke[ledgerService](injector, "missing")
}

func TestTwoServiceTypesCoexist(t *testing.T) {
	injector := di.New()
	ledgerFailure := di.Provide(injector, "ledger", func(active *di.Injector) (ledgerService, error) {
		return ledgerService{Balance: 11}, nil
	})
	if ledgerFailure != nil {
		t.Fatalf("Provide ledger failure = %v, want nil", ledgerFailure)
	}
	auditFailure := di.Provide(injector, "audit", func(active *di.Injector) (auditService, error) {
		return auditService{Trail: "kept"}, nil
	})
	if auditFailure != nil {
		t.Fatalf("Provide audit failure = %v, want nil", auditFailure)
	}
	fetchedLedger, ledgerInvokeFailure := di.Invoke[ledgerService](injector, "ledger")
	if ledgerInvokeFailure != nil {
		t.Fatalf("Invoke ledger failure = %v, want nil", ledgerInvokeFailure)
	}
	if fetchedLedger.Balance != 11 {
		t.Fatalf("Invoke ledger Balance = %d, want 11", fetchedLedger.Balance)
	}
	fetchedAudit, auditInvokeFailure := di.Invoke[auditService](injector, "audit")
	if auditInvokeFailure != nil {
		t.Fatalf("Invoke audit failure = %v, want nil", auditInvokeFailure)
	}
	if fetchedAudit.Trail != "kept" {
		t.Fatalf("Invoke audit Trail = %q, want %q", fetchedAudit.Trail, "kept")
	}
}

func TestPanickingConstructorPanicsAllWaiters(t *testing.T) {
	// No rendezvous needed: the panic reset guarantees every caller panics
	// with the payload whether it led, waited, or arrived late and retried.
	boomInjector := di.New()
	provideFailure := di.Provide(boomInjector, "boom", func(active *di.Injector) (auditService, error) {
		panic("boom")
	})
	if provideFailure != nil {
		t.Fatalf("Provide boom failure = %v, want nil", provideFailure)
	}
	const waiterCount = 8
	var waiterGroup sync.WaitGroup
	panics := make([]any, waiterCount+1)
	for position := range waiterCount + 1 {
		waiterGroup.Add(1)
		go func() {
			defer waiterGroup.Done()
			defer func() {
				panics[position] = recover()
			}()
			_ = di.MustInvoke[auditService](boomInjector, "boom")
		}()
	}
	waiterGroup.Wait()
	for position, payload := range panics {
		if payload != "boom" {
			t.Fatalf("caller %d panic payload = %v, want boom", position, payload)
		}
	}
}
