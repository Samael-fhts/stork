package configmigrator

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	dbmodel "isc.org/stork/server/database/model"
	storkutil "isc.org/stork/util"
)

type Service struct {
	filter     dbmodel.HostsByPageFilters
	totalCount int64
	startDate  time.Time

	processedCount atomic.Int64
	errors         map[int64]error
	generalErrors  []error
	// The sync package has an atomic map type, but it has no length method.
	errorsMutex sync.RWMutex
	endDate     storkutil.AtomicTime

	cancelSignal func()
	cancelWg     sync.WaitGroup

	limit int64
}

func NewService() *Service {
	return &Service{
		errors: make(map[int64]error),
		limit:  100,
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

func (s *Service) GetFilter() dbmodel.HostsByPageFilters {
	return s.filter
}

func (s *Service) Start(filter dbmodel.HostsByPageFilters) error {
	if s.HasMigration() {
		return errors.New("Migration already started")
	}

	s.filter = filter
	s.startDate = time.Now()

	total, err := s.countTotal()
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

	s.filter = dbmodel.HostsByPageFilters{}
	s.totalCount = 0
	s.processedCount.Store(0)
	s.errors = make(map[int64]error)
	s.startDate = time.Time{}
	s.endDate.Store(time.Time{})
	s.generalErrors = []error{}

	s.cancelSignal = nil
	s.cancelWg.Done()

	return nil
}

func (s *Service) countTotal() (int64, error) {

}

func (s *Service) loadChunk(offset int64) ([]any, error) {

}

func (s *Service) getID(item any) int64 {

}

func (s *Service) migrate(item any) error {

}

func (s *Service) cancel() error {
	s.cancelSignal()
	s.cancelWg.Wait()
	return nil
}

func (s *Service) execute(ctx context.Context) {
	s.cancelWg.Add(1)

	go func(ctx context.Context) {
		offset := int64(0)

		// The label is needed to break the loop inside the select statement.
	LOOP:
		for {
			select {
			case <-ctx.Done():
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

		s.cancelWg.Done()
		s.endDate.Store(time.Now())
	}(ctx)

}

func (s *Service) executeIteration(offset int64) (int64, bool) {
	items, err := s.loadChunk(offset)
	if err != nil {
		s.errorsMutex.Lock()
		defer s.errorsMutex.Unlock()
		s.generalErrors = append(s.generalErrors, err)
		return s.limit, false
	}

	itemsCount := len(items)
	if itemsCount == 0 {
		return 0, true
	}

	for _, item := range items {
		err := s.migrate(item)
		if err != nil {
			s.errorsMutex.Lock()
			defer s.errorsMutex.Unlock()
			s.errors[s.getID(item)] = err
		}
	}
	return int64(itemsCount), false
}
