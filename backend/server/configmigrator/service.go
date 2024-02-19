package configmigrator

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	storkutil "isc.org/stork/util"
)

type Service struct {
	migrator Migrator

	totalCount int64
	startDate  time.Time

	processedCount atomic.Int64
	errors         map[int64]error
	generalErrors  []error
	// The sync package has an atomic map type, but it has no length method.
	errorsMutex sync.RWMutex
	endDate     storkutil.AtomicTime

	cancelSignal func()
	finishWg     sync.WaitGroup
}

func NewService() *Service {
	return &Service{
		errors: make(map[int64]error),
	}
}

func (s *Service) GetProgress() float64 {
	return float64(s.processedCount.Load()) / float64(s.totalCount)
}

func (s *Service) GetErrorCount() int {
	s.errorsMutex.RLock()
	defer s.errorsMutex.RUnlock()
	return len(s.errors)
}

func (s *Service) IsInProgress() bool {
	date, _ := s.endDate.Load()
	return date == time.Time{}
}

func (s *Service) HasMigration() bool {
	return s.startDate != time.Time{}
}

func (s *Service) Start() error {
	if s.HasMigration() {
		return errors.New("Migration already started")
	}

	s.startDate = time.Now()

	total, err := s.migrator.CountTotal()
	if err != nil {
		s.endDate.Store(time.Now())
		return err
	}
	s.totalCount = total

	ctx, cancel := context.WithCancel(context.Background())
	s.cancelSignal = cancel

	go s.execute(ctx)
	return nil
}

func (s *Service) Stop() error {
	if !s.HasMigration() {
		return errors.New("Migration not started")
	}

	if !s.IsInProgress() {
		return nil
	}

	return s.cancel()
}

func (s *Service) Clear() error {
	if !s.HasMigration() {
		return nil
	}

	if s.IsInProgress() {
		return errors.New("Migration in progress")
	}

	s.totalCount = 0
	s.processedCount.Store(0)
	s.errors = make(map[int64]error)
	s.startDate = time.Time{}
	s.endDate.Store(time.Time{})
	s.generalErrors = []error{}

	s.cancelSignal = nil
	s.finishWg.Done()

	return nil
}

func (s *Service) cancel() error {
	s.cancelSignal()
	s.finishWg.Wait()
	return nil
}

func (s *Service) execute(ctx context.Context) {
	s.finishWg.Add(1)

	go func(ctx context.Context) {
		offset := int64(0)
		canceled := false

		// The label is needed to break the loop inside the select statement.
	LOOP:
		for {
			select {
			case <-ctx.Done():
				canceled = true
				break LOOP
			default:
				processed, done := s.executeIteration(int64(offset))
				s.processedCount.Add(processed)
				offset += processed
				if done {
					break LOOP
				}
			}
		}

		if !canceled {
			// The migration was not interrupted.
			s.migrator.Finish()
		}

		s.finishWg.Done()
		s.endDate.Store(time.Now())
	}(ctx)

}

func (s *Service) executeIteration(offset int64) (int64, bool) {
	loadedCount, err := s.migrator.LoadItems(offset)
	if err != nil {
		s.appendGeneralError(err)
		return loadedCount, false
	}

	if loadedCount == 0 {
		return 0, true
	}

	errs := s.migrator.Migrate()

	s.appendMigrationErrors(errs)
	return loadedCount, false
}

func (s *Service) appendGeneralError(err error) {
	s.errorsMutex.Lock()
	defer s.errorsMutex.Unlock()
	s.generalErrors = append(s.generalErrors, err)
}

func (s *Service) appendMigrationErrors(errs map[int64]error) {
	s.errorsMutex.Lock()
	defer s.errorsMutex.Unlock()
	for id, err := range errs {
		s.errors[id] = err
	}
}
