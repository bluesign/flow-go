// (c) 2022 Dapper Labs - ALL RIGHTS RESERVED

package uploader

import (
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/onflow/flow-go/model/flow"
	"github.com/onflow/flow-go/module/metrics"
	"github.com/onflow/flow-go/utils/unittest"

	"github.com/stretchr/testify/mock"

	storageMock "github.com/onflow/flow-go/storage/mock"

	"github.com/stretchr/testify/assert"

	"github.com/onflow/flow-go/engine"

	"github.com/onflow/flow-go/engine/execution"
	uploaderMock "github.com/onflow/flow-go/engine/execution/computation/computer/uploader/mock"
)

func Test_non_ReadyDoneAware_Uploader(t *testing.T) {
	mockNonReadyDoneAwareUploader := new(uploaderMock.Uploader)
	testRetryableUploader := NewBadgerRetryableUploader(
		mockNonReadyDoneAwareUploader,
		nil,
		&metrics.NoopCollector{},
	)

	<-testRetryableUploader.Ready()
	<-testRetryableUploader.Done()

	mockNonReadyDoneAwareUploader.AssertNotCalled(t, "Ready")
	mockNonReadyDoneAwareUploader.AssertNotCalled(t, "Done")
}

func Test_ReadyDoneAware_Uploader(t *testing.T) {
	mockReadyDoneAwareUploader := NewTestUploader()

	testRetryableUploader := NewBadgerRetryableUploader(
		mockReadyDoneAwareUploader,
		nil,
		&metrics.NoopCollector{},
	)
	defer testRetryableUploader.Done()

	<-testRetryableUploader.Ready()
	<-testRetryableUploader.Done()

	assert.True(t, mockReadyDoneAwareUploader.ReadyCalled)
	assert.True(t, mockReadyDoneAwareUploader.DoneCalled)
}

func Test_Upload_invoke(t *testing.T) {
	mockTestUploader := NewTestUploader()
	mockComputationResultStorage := new(storageMock.ComputationResults)
	mockComputationResultStorage.On("Store", mock.Anything, mock.Anything).Return(nil)

	testRetryableUploader := NewBadgerRetryableUploader(
		mockTestUploader,
		mockComputationResultStorage,
		&metrics.NoopCollector{},
	)
	defer testRetryableUploader.Done()

	// nil input - no call to Upload()
	err := testRetryableUploader.Upload(nil)
	assert.NotNil(t, err)
	assert.False(t, mockTestUploader.UploadCalled)

	// non-nil input - Upload() should be called
	testComputationResult := createTestComputationResult()
	err = testRetryableUploader.Upload(testComputationResult)
	assert.Nil(t, err)
	assert.True(t, mockTestUploader.UploadCalled)
}

func Test_RetryUpload(t *testing.T) {
	mockTestUploader := NewTestUploader()
	mockComputationResultStorage := new(storageMock.ComputationResults)

	testID := flow.HashToID([]byte{1, 2, 3})
	mockComputationResultStorage.On("GetAllIDs").Return([]flow.Identifier{testID}, nil)
	testComputationResult := createTestComputationResult()
	mockComputationResultStorage.On("ByID", testID).Return(testComputationResult, nil)

	testRetryableUploader := NewBadgerRetryableUploader(
		mockTestUploader,
		mockComputationResultStorage,
		&metrics.NoopCollector{},
	)
	defer testRetryableUploader.Done()

	err := testRetryableUploader.RetryUpload()
	assert.Nil(t, err)
	assert.True(t, mockTestUploader.UploadCalled)
}

func Test_AsyncUploaderCallback(t *testing.T) {
	wgUploadCalleded := sync.WaitGroup{}
	wgUploadCalleded.Add(1)

	uploader := &DummyUploader{
		f: func() error {
			wgUploadCalleded.Done()
			return nil
		},
	}

	asyncUploader := NewAsyncUploader(uploader,
		1*time.Nanosecond, 1, zerolog.Nop(), &metrics.NoopCollector{})

	mockComputationResultStorage := new(storageMock.ComputationResults)
	mockComputationResultStorage.On("Store", mock.Anything, mock.Anything).Return(nil)
	// Remove() should be called when callback func is called.
	mockComputationResultStorage.On("Remove", mock.Anything).Return(nil).Once()

	testRetryableUploader := NewBadgerRetryableUploader(
		asyncUploader,
		mockComputationResultStorage,
		&metrics.NoopCollector{},
	)
	defer testRetryableUploader.Done()

	testComputationResult := createTestComputationResult()
	err := testRetryableUploader.Upload(testComputationResult)
	assert.Nil(t, err)

	wgUploadCalleded.Wait()
}

// createTestComputationResult() creates ComputationResult with valid ExecutableBlock ID
func createTestComputationResult() *execution.ComputationResult {
	testComputationResult := &execution.ComputationResult{}
	blockA := unittest.BlockHeaderFixture()
	blockB := unittest.ExecutableBlockFixtureWithParent(nil, &blockA)
	testComputationResult.ExecutableBlock = blockB
	return testComputationResult
}

// TestUploader is an implementation to ReadyDoneAware and Uploader interface.
type TestUploader struct {
	unit         *engine.Unit
	ReadyCalled  bool
	DoneCalled   bool
	UploadCalled bool
}

func NewTestUploader() *TestUploader {
	return &TestUploader{
		unit: engine.NewUnit(),
	}
}

func (t *TestUploader) Ready() <-chan struct{} {
	t.ReadyCalled = true
	return t.unit.Ready()
}

func (t *TestUploader) Done() <-chan struct{} {
	t.DoneCalled = true
	return t.unit.Done()
}

func (t *TestUploader) Upload(_ *execution.ComputationResult) error {
	t.UploadCalled = true
	return nil
}
