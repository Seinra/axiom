package di_test

import (
	"errors"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

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

func TestInvokeOnNilInjectorReportsNotFound(t *testing.T) {
	// Contract: no container holds no service, so a nil Injector reports the
	// same ErrServiceNotFound sentinel as an unknown name. The message text
	// ("nil Injector" versus "unknown service") tells the two apart.
	var nilInjector *di.Injector
	_, invokeFailure := di.Invoke[ledgerService](nilInjector, "ledger")
	if invokeFailure == nil {
		t.Fatal("Invoke on nil failure = nil, want ErrServiceNotFound")
	}
	if !errors.Is(invokeFailure, di.ErrServiceNotFound) {
		t.Fatalf("Invoke on nil failure = %v, want ErrServiceNotFound", invokeFailure)
	}
	if got := invokeFailure.Error(); !strings.Contains(got, "nil Injector") {
		t.Fatalf("Invoke on nil failure = %q, want message to name nil Injector", got)
	}
	missingInjector := di.New()
	_, missingFailure := di.Invoke[ledgerService](missingInjector, "ledger")
	if !errors.Is(missingFailure, di.ErrServiceNotFound) {
		t.Fatalf("Invoke missing failure = %v, want ErrServiceNotFound", missingFailure)
	}
	if got := missingFailure.Error(); !strings.Contains(got, "unknown service") {
		t.Fatalf("Invoke missing failure = %q, want message to name unknown service", got)
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

func TestPanickingConstructorRunsOncePerCaller(t *testing.T) {
	// Retry-then-panic contract: waiters woken by the panic reset retry and
	// re-run the constructor instead of receiving the payload directly, so
	// with callerTotal concurrent callers the constructor runs callerTotal
	// times and every caller still panics with the same payload.
	const callerTotal = 8
	boomInjector := di.New()
	var builds atomic.Int64
	provideFailure := di.Provide(boomInjector, "boom", func(active *di.Injector) (auditService, error) {
		builds.Add(1)
		panic("boom")
	})
	if provideFailure != nil {
		t.Fatalf("Provide boom failure = %v, want nil", provideFailure)
	}
	var waiterGroup sync.WaitGroup
	startSignal := make(chan struct{})
	panics := make([]any, callerTotal)
	for position := range callerTotal {
		waiterGroup.Add(1)
		go func() {
			defer waiterGroup.Done()
			defer func() {
				panics[position] = recover()
			}()
			<-startSignal
			_ = di.MustInvoke[auditService](boomInjector, "boom")
		}()
	}
	close(startSignal)
	waiterGroup.Wait()
	for position, payload := range panics {
		if payload != "boom" {
			t.Fatalf("caller %d panic payload = %v, want boom", position, payload)
		}
	}
	if builds.Load() != callerTotal {
		t.Fatalf("constructor calls = %d, want %d (once per caller)", builds.Load(), callerTotal)
	}
}

func TestPanickingConstructorRetriesUntilSuccess(t *testing.T) {
	// The panic reset leaves the entry reusable: after panics stop, the next
	// Invoke builds and caches normally instead of sticking.
	flakyInjector := di.New()
	var builds atomic.Int64
	provideFailure := di.Provide(flakyInjector, "flaky", func(active *di.Injector) (ledgerService, error) {
		if builds.Add(1) <= 2 {
			panic("boom")
		}
		return ledgerService{Balance: 99}, nil
	})
	if provideFailure != nil {
		t.Fatalf("Provide flaky failure = %v, want nil", provideFailure)
	}
	for attempt := 1; attempt <= 2; attempt++ {
		panicked := func() (didPanic any) {
			defer func() { didPanic = recover() }()
			_, _ = di.Invoke[ledgerService](flakyInjector, "flaky")
			return nil
		}()
		if panicked != "boom" {
			t.Fatalf("attempt %d panic payload = %v, want boom", attempt, panicked)
		}
	}
	fetched, invokeFailure := di.Invoke[ledgerService](flakyInjector, "flaky")
	if invokeFailure != nil {
		t.Fatalf("third Invoke failure = %v, want nil", invokeFailure)
	}
	if fetched.Balance != 99 {
		t.Fatalf("third Invoke Balance = %d, want 99", fetched.Balance)
	}
	if builds.Load() != 3 {
		t.Fatalf("constructor calls = %d, want 3 (two panics then cached success)", builds.Load())
	}
}

func TestReplaceThenInvokeRebuilds(t *testing.T) {
	replaceInjector := di.New()
	var firstBuilds atomic.Int64
	var secondBuilds atomic.Int64
	provideFailure := di.Provide(replaceInjector, "ledger", func(active *di.Injector) (ledgerService, error) {
		firstBuilds.Add(1)
		return ledgerService{Balance: 1}, nil
	})
	if provideFailure != nil {
		t.Fatalf("Provide failure = %v, want nil", provideFailure)
	}
	firstFetched, firstFailure := di.Invoke[ledgerService](replaceInjector, "ledger")
	if firstFailure != nil {
		t.Fatalf("first Invoke failure = %v, want nil", firstFailure)
	}
	if firstFetched.Balance != 1 {
		t.Fatalf("first Invoke Balance = %d, want 1", firstFetched.Balance)
	}
	replaceFailure := di.Replace(replaceInjector, "ledger", func(active *di.Injector) (ledgerService, error) {
		secondBuilds.Add(1)
		return ledgerService{Balance: 2}, nil
	})
	if replaceFailure != nil {
		t.Fatalf("Replace failure = %v, want nil", replaceFailure)
	}
	secondFetched, secondFailure := di.Invoke[ledgerService](replaceInjector, "ledger")
	if secondFailure != nil {
		t.Fatalf("second Invoke failure = %v, want nil", secondFailure)
	}
	if secondFetched.Balance != 2 {
		t.Fatalf("second Invoke Balance = %d, want 2", secondFetched.Balance)
	}
	if firstBuilds.Load() != 1 {
		t.Fatalf("first constructor calls = %d, want 1", firstBuilds.Load())
	}
	if secondBuilds.Load() != 1 {
		t.Fatalf("second constructor calls = %d, want 1 (rebuild after Replace)", secondBuilds.Load())
	}
	cached, thirdFailure := di.Invoke[ledgerService](replaceInjector, "ledger")
	if thirdFailure != nil {
		t.Fatalf("third Invoke failure = %v, want nil", thirdFailure)
	}
	if cached.Balance != 2 {
		t.Fatalf("third Invoke Balance = %d, want 2", cached.Balance)
	}
	if secondBuilds.Load() != 1 {
		t.Fatalf("second constructor calls after cache = %d, want 1", secondBuilds.Load())
	}
}

func TestReplaceUnknownReportsNotFound(t *testing.T) {
	missingInjector := di.New()
	replaceFailure := di.Replace(missingInjector, "missing", func(active *di.Injector) (ledgerService, error) {
		return ledgerService{Balance: 1}, nil
	})
	if replaceFailure == nil {
		t.Fatal("Replace failure = nil, want ErrServiceNotFound")
	}
	if !errors.Is(replaceFailure, di.ErrServiceNotFound) {
		t.Fatalf("Replace failure = %v, want ErrServiceNotFound", replaceFailure)
	}
}

func TestReplaceDuringBlockedBuildReportsInFlight(t *testing.T) {
	blockedInjector := di.New()
	var executions atomic.Int64
	releaseBuild := make(chan struct{})
	provideFailure := di.Provide(blockedInjector, "blocked", func(active *di.Injector) (ledgerService, error) {
		executions.Add(1)
		<-releaseBuild
		return ledgerService{Balance: 21}, nil
	})
	if provideFailure != nil {
		t.Fatalf("Provide failure = %v, want nil", provideFailure)
	}
	waiterDone := make(chan struct{})
	var waiterFetched ledgerService
	var waiterFailure error
	go func() {
		defer close(waiterDone)
		waiterFetched, waiterFailure = di.Invoke[ledgerService](blockedInjector, "blocked")
	}()
	deadline := time.Now().Add(2 * time.Second)
	for executions.Load() < 1 {
		if time.Now().After(deadline) {
			t.Fatal("blocked build did not start before deadline")
		}
		runtime.Gosched()
	}
	replaceFailure := di.Replace(blockedInjector, "blocked", func(active *di.Injector) (ledgerService, error) {
		return ledgerService{Balance: 99}, nil
	})
	if !errors.Is(replaceFailure, di.ErrBuildInFlight) {
		t.Fatalf("Replace during build failure = %v, want ErrBuildInFlight", replaceFailure)
	}
	close(releaseBuild)
	<-waiterDone
	if waiterFailure != nil {
		t.Fatalf("waiter Invoke failure = %v, want nil", waiterFailure)
	}
	if waiterFetched.Balance != 21 {
		t.Fatalf("waiter Balance = %d, want 21 (original build wins)", waiterFetched.Balance)
	}
	cached, invokeFailure := di.Invoke[ledgerService](blockedInjector, "blocked")
	if invokeFailure != nil {
		t.Fatalf("Invoke after blocked build failure = %v, want nil", invokeFailure)
	}
	if cached.Balance != 21 {
		t.Fatalf("Invoke after blocked build Balance = %d, want 21 (Replace changed nothing)", cached.Balance)
	}
	if executions.Load() != 1 {
		t.Fatalf("constructor calls = %d, want 1 (no rebuild from refused Replace)", executions.Load())
	}
}

func TestReplaceGuardsReportErrors(t *testing.T) {
	var nilInjector *di.Injector
	nilFailure := di.Replace(nilInjector, "ledger", func(active *di.Injector) (ledgerService, error) {
		return ledgerService{}, nil
	})
	if !errors.Is(nilFailure, di.ErrServiceNotFound) {
		t.Fatalf("Replace on nil failure = %v, want ErrServiceNotFound", nilFailure)
	}
	guardInjector := di.New()
	provideFailure := di.Provide(guardInjector, "ledger", func(active *di.Injector) (ledgerService, error) {
		return ledgerService{Balance: 1}, nil
	})
	if provideFailure != nil {
		t.Fatalf("Provide failure = %v, want nil", provideFailure)
	}
	emptyFailure := di.Replace(guardInjector, "", func(active *di.Injector) (ledgerService, error) {
		return ledgerService{}, nil
	})
	if emptyFailure == nil {
		t.Fatal("Replace empty name failure = nil, want descriptive error")
	}
	nilConstructorFailure := di.Replace[ledgerService](guardInjector, "ledger", nil)
	if nilConstructorFailure == nil {
		t.Fatal("Replace nil constructor failure = nil, want descriptive error")
	}
	unknownFailure := di.Replace(guardInjector, "missing", func(active *di.Injector) (ledgerService, error) {
		return ledgerService{}, nil
	})
	if !errors.Is(unknownFailure, di.ErrServiceNotFound) {
		t.Fatalf("Replace unknown failure = %v, want ErrServiceNotFound", unknownFailure)
	}
}
