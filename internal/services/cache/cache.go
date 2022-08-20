package cache

import (
	"log"
	"sync"
	"time"

	"github.com/ArtemVoronov/artforintrovert-test/internal/services/records"
	"github.com/ArtemVoronov/artforintrovert-test/internal/utils"
	"github.com/gin-gonic/gin"
)

type CacheService interface {
	ShutDown()
	RecordsCacheToJSON(c *gin.Context)
}

type Service struct {
	quit         chan struct{}
	recordsCache *[]records.Record
	rwm          sync.RWMutex

	minDelay    time.Duration
	maxDelay    time.Duration
	factorDelay time.Duration
}

var once sync.Once
var instance *Service

func Instance() *Service {
	once.Do(func() {
		if instance == nil {
			instance = createService()
			instance.setup()
			instance.startSync()
		}
	})
	return instance
}

func (s *Service) ShutDown() {
	defer close(s.quit)
	s.quit <- struct{}{}

}

func (s *Service) RecordsCacheToJSON(c *gin.Context, status int) {
	s.rwm.RLock()
	defer s.rwm.RUnlock()
	c.JSON(status, s.recordsCache)
}

func (s *Service) startSync() {
	go func() {
		delay := s.minDelay
		for {
			select {
			case <-s.quit:
				log.Printf("sync cache stopped")
				return
			default:
				time.Sleep(delay)
				records, err := records.Instance().GetAll()
				if err != nil {
					log.Printf("sync cache error: %v", err)
					if delay < s.maxDelay {
						delay = delay * s.factorDelay
						log.Printf("sync delay increased by '%d' FACTOR to the value: %v", s.factorDelay, delay)
					}

					if delay >= s.maxDelay {
						delay = s.maxDelay
						log.Printf("sync delay has maximum value: %v", delay)
					}

					continue
				} else {
					delay = s.minDelay
				}

				s.reloadCache(&records)
			}
		}
	}()
}

func createService() *Service {
	return &Service{
		quit:         make(chan struct{}),
		recordsCache: &[]records.Record{},
		minDelay:     updateMinCacheInterval(),
		maxDelay:     updateMaxCacheInterval(),
		factorDelay:  updateCacheIntervalFactor(),
	}
}

func (s *Service) setup() {
	records, err := records.Instance().GetAll()
	if err != nil {
		log.Printf("unable to init records cache: %v", err)
		return
	}
	s.reloadCache(&records)
	log.Printf("records cache initiation succeed")
}

func (s *Service) reloadCache(newRecordsCache *[]records.Record) {
	s.rwm.Lock()
	defer s.rwm.Unlock()
	s.recordsCache = newRecordsCache
}

func updateMinCacheInterval() time.Duration {
	value := utils.EnvVarIntDefault("UPDATE_CACHE_MIN_INTERVAL_IN_SECONDS", "30")
	return time.Duration(value) * time.Second
}

func updateMaxCacheInterval() time.Duration {
	value := utils.EnvVarIntDefault("UPDATE_CACHE_MAX_INTERVAL_IN_SECONDS", "86400")
	return time.Duration(value) * time.Second
}

func updateCacheIntervalFactor() time.Duration {
	value := utils.EnvVarIntDefault("UPDATE_CACHE_INTERVAL_FACTOR", "2")
	return time.Duration(value)
}
