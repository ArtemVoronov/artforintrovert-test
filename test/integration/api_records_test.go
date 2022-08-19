//go:build integration
// +build integration

package integration

import (
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/ArtemVoronov/artforintrovert-test/internal/api"
	"github.com/stretchr/testify/assert"
)

const (
	GOROUTINES_LIMIT = 10000
	DELAY_BETWEEN_OP = 5
)

var (
	ERROR_RECORD_ID_IS_REQUIRED string = "{\"errors\":[" +
		"{\"Field\":\"Id\",\"Msg\":\"This field is required\"}" +
		"]}"
	ERROR_RECORD_DATA_IS_REQUIRED string = "{\"errors\":[" +
		"{\"Field\":\"Data\",\"Msg\":\"This field is required\"}" +
		"]}"
)

func TestApiRecordGetAll(t *testing.T) {
	t.Run("BasicCase", RunWithRecreateDB(func(t *testing.T) {
		httpStatusCode, body, err := testHttpClient.GetAllRecords()
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, httpStatusCode)
		assert.Equal(t, "[]", body)
	}))
}
func TestApiRecordInsert(t *testing.T) {
	t.Run("BasicCase", RunWithRecreateDB(func(t *testing.T) {
		httpStatusCode, body, err := testHttpClient.UpsertRecord(nil, "exponent")

		assert.Nil(t, err)
		assert.Equal(t, http.StatusCreated, httpStatusCode)
		assert.Equal(t, 26, len(body))
	}))
	t.Run("Bulk", RunWithRecreateDB(func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(GOROUTINES_LIMIT)
		for i := 0; i < GOROUTINES_LIMIT; i++ {
			go func() {
				defer wg.Done()
				httpStatusCode, body, err := testHttpClient.UpsertRecord(nil, "exponent")

				assert.Nil(t, err)
				assert.Equal(t, http.StatusCreated, httpStatusCode)
				assert.Equal(t, 26, len(body))
			}()
		}

		wg.Wait()
		time.Sleep(DELAY_BETWEEN_OP * time.Second)
		httpStatusCode, body, err := testHttpClient.GetAllRecords()
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, httpStatusCode)
		records, err := ToRecords(body)
		assert.Nil(t, err)
		assert.Equal(t, GOROUTINES_LIMIT, len(records))
		for _, record := range records {
			assert.Equal(t, record.Data, "exponent")
		}
	}))
}
func TestApiRecordUpdate(t *testing.T) {
	t.Run("BasicCaseUpdate", RunWithRecreateDB(func(t *testing.T) {
		httpStatusCode, body, err := testHttpClient.UpsertRecord(nil, "exponent")
		id := ToId(body)

		httpStatusCode, body, err = testHttpClient.UpsertRecord(id, "pi")

		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, httpStatusCode)
		assert.Equal(t, "\"Done\"", body)
	}))
	t.Run("MissedData", RunWithRecreateDB(func(t *testing.T) {
		httpStatusCode, body, err := testHttpClient.UpsertRecord("62ffcac20074ec24bbb5810d", nil)

		assert.Nil(t, err)
		assert.Equal(t, http.StatusBadRequest, httpStatusCode)
		assert.Equal(t, ERROR_RECORD_DATA_IS_REQUIRED, body)
	}))
	t.Run("WrongTypeForData", RunWithRecreateDB(func(t *testing.T) {
		httpStatusCode, body, err := testHttpClient.UpsertRecord(nil, 123)

		assert.Nil(t, err)
		assert.Equal(t, http.StatusBadRequest, httpStatusCode)
		assert.Equal(t, "\""+api.ERROR_MESSAGE_PARSING_BODY_JSON+"\"", body)
	}))
	t.Run("Bulk", RunWithRecreateDB(func(t *testing.T) {
		var m sync.Mutex
		var wg sync.WaitGroup
		var ids []string = make([]string, 0)
		wg.Add(GOROUTINES_LIMIT)
		for i := 0; i < GOROUTINES_LIMIT; i++ {
			go func() {
				defer wg.Done()
				httpStatusCode, body, err := testHttpClient.UpsertRecord(nil, "exponent")

				assert.Nil(t, err)
				assert.Equal(t, http.StatusCreated, httpStatusCode)
				assert.Equal(t, 26, len(body))
				id := ToId(body)
				m.Lock()
				ids = append(ids, id)
				m.Unlock()
			}()
		}

		wg.Wait()
		time.Sleep(DELAY_BETWEEN_OP * time.Second)
		assert.Equal(t, GOROUTINES_LIMIT, len(ids))

		wg.Add(len(ids))
		for _, id := range ids {
			idLocal := id
			go func() {
				defer wg.Done()
				httpStatusCode, body, err := testHttpClient.UpsertRecord(idLocal, "pi")

				assert.Nil(t, err)
				assert.Equal(t, http.StatusOK, httpStatusCode)
				assert.Equal(t, "\"Done\"", body)
			}()
		}

		wg.Wait()
		time.Sleep(DELAY_BETWEEN_OP * time.Second)

		httpStatusCode, body, err := testHttpClient.GetAllRecords()
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, httpStatusCode)
		records, err := ToRecords(body)
		assert.Nil(t, err)
		assert.Equal(t, GOROUTINES_LIMIT, len(records))
		for _, record := range records {
			assert.Equal(t, record.Data, "pi")
		}
	}))
}

func TestApiRecordDelete(t *testing.T) {
	t.Run("BasicCase", RunWithRecreateDB(func(t *testing.T) {
		httpStatusCode, body, err := testHttpClient.UpsertRecord(nil, "exponent")
		id := ToId(body)

		httpStatusCode, body, err = testHttpClient.DeleteRecord(id)

		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, httpStatusCode)
		assert.Equal(t, "\"Done\"", body)

		time.Sleep(DELAY_BETWEEN_OP * time.Second)

		httpStatusCode, body, err = testHttpClient.GetAllRecords()
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, httpStatusCode)
		assert.Equal(t, "[]", body)
	}))
	t.Run("MissedId", RunWithRecreateDB(func(t *testing.T) {
		httpStatusCode, body, err := testHttpClient.DeleteRecord(nil)

		assert.Nil(t, err)
		assert.Equal(t, http.StatusBadRequest, httpStatusCode)
		assert.Equal(t, ERROR_RECORD_ID_IS_REQUIRED, body)
	}))
	t.Run("WrongTypeForId", RunWithRecreateDB(func(t *testing.T) {
		httpStatusCode, body, err := testHttpClient.DeleteRecord(123)

		assert.Nil(t, err)
		assert.Equal(t, http.StatusBadRequest, httpStatusCode)
		assert.Equal(t, "\""+api.ERROR_MESSAGE_PARSING_BODY_JSON+"\"", body)
	}))
	t.Run("Bulk", RunWithRecreateDB(func(t *testing.T) {
		var m sync.Mutex
		var wg sync.WaitGroup
		var ids []string = make([]string, 0)
		wg.Add(GOROUTINES_LIMIT)
		for i := 0; i < GOROUTINES_LIMIT; i++ {
			go func() {
				defer wg.Done()
				httpStatusCode, body, err := testHttpClient.UpsertRecord(nil, "exponent")

				assert.Nil(t, err)
				assert.Equal(t, http.StatusCreated, httpStatusCode)
				assert.Equal(t, 26, len(body))
				id := ToId(body)
				m.Lock()
				ids = append(ids, id)
				m.Unlock()
			}()
		}

		wg.Wait()
		time.Sleep(DELAY_BETWEEN_OP * time.Second)
		assert.Equal(t, GOROUTINES_LIMIT, len(ids))

		wg.Add(len(ids))
		for _, id := range ids {
			idLocal := id
			go func() {
				defer wg.Done()
				httpStatusCode, body, err := testHttpClient.DeleteRecord(idLocal)

				assert.Nil(t, err)
				assert.Equal(t, http.StatusOK, httpStatusCode)
				assert.Equal(t, "\"Done\"", body)
			}()
		}

		wg.Wait()
		time.Sleep(DELAY_BETWEEN_OP * time.Second)

		httpStatusCode, body, err := testHttpClient.GetAllRecords()
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, httpStatusCode)
		assert.Equal(t, "[]", body)
	}))
}
