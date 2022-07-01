// (c) 2022 Dapper Labs - ALL RIGHTS RESERVED

package uploader

import (
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/onflow/flow-go/engine"
	"github.com/onflow/flow-go/engine/execution"
	"github.com/onflow/flow-go/module"
	"github.com/onflow/flow-go/storage"
)

// RetryableUploader defines the interface for uploader that is retryable
type RetryableUploader interface {
	Uploader
	RetryUpload() error
}

// BadgerRetryableUploader is the BadgerDB based implementation to RetryableUploader
type BadgerRetryableUploader struct {
	uploader Uploader
	store    storage.ComputationResults
	unit     *engine.Unit
	metrics  module.ExecutionMetrics
}

func NewBadgerRetryableUploader(
	uploader Uploader,
	store storage.ComputationResults,
	metrics module.ExecutionMetrics) *BadgerRetryableUploader {

	// NOTE: only AsyncUploader is supported for now.
	switch uploader.(type) {
	case *AsyncUploader:
		// When Uploade() is successful, the stored ComputationResult in BadgerDB will be removed.
		onCompleteCB := func(computationResult *execution.ComputationResult, err error) {
			if err != nil {
				log.Warn().Msg(fmt.Sprintf("ComputationResults upload failed with ID %s",
					computationResult.ExecutableBlock.ID()))
				return
			}

			if computationResult == nil || computationResult.ExecutableBlock == nil {
				log.Warn().Msg(fmt.Sprintf("Invalid ComputationResults parameter"))
				return
			}

			if err = store.Remove(computationResult.ExecutableBlock.ID()); err != nil {
				log.Warn().Msg(fmt.Sprintf(
					"ComputationResults with ID %s failed to be removed on local disk. ERR: %s ",
					computationResult.ExecutableBlock.ID(), err.Error()))
			}

			metrics.ExecutionComputationResultUploaded()
		}
		uploader.(*AsyncUploader).SetOnCompleteCallback(onCompleteCB)
	}

	return &BadgerRetryableUploader{
		uploader,
		store,
		engine.NewUnit(),
		metrics,
	}
}

func (b *BadgerRetryableUploader) Ready() <-chan struct{} {
	switch b.uploader.(type) {
	case module.ReadyDoneAware:
		readyDoneAwareUploader := b.uploader.(module.ReadyDoneAware)
		return readyDoneAwareUploader.Ready()
	}
	return b.unit.Ready()
}

func (b *BadgerRetryableUploader) Done() <-chan struct{} {
	switch b.uploader.(type) {
	case module.ReadyDoneAware:
		readyDoneAwareUploader := b.uploader.(module.ReadyDoneAware)
		return readyDoneAwareUploader.Done()
	}
	return b.unit.Done()
}

func (b *BadgerRetryableUploader) Upload(computationResult *execution.ComputationResult) error {
	if computationResult == nil || computationResult.ExecutableBlock == nil {
		return errors.New("ComputationResult or its ExecutableBlock is nil when Upload() is called.")
	}

	// Before upload we store ComputationResult to BadgerDB. It will be removed when upload succeeds.
	//
	// NOTE: the size of ComputationResult storage will keep increase when uploader keeps failing,
	//		 but we don't expect that to happen very frequently so not handling the potential unbounded
	//		 storage overhead scenario for now.
	if err := b.store.Store(computationResult.ExecutableBlock.ID(), computationResult); err != nil {
		log.Warn().Msg(
			fmt.Sprintf("Failed to store ComputationResult into local DB with ID %s",
				computationResult.ExecutableBlock.ID()))
	}

	return b.uploader.Upload(computationResult)
}

func (b *BadgerRetryableUploader) RetryUpload() error {
	computationResultsIDs, err := b.store.GetAllIDs()
	if err != nil {
		log.Error().Err(err).Msg("Failed to load list of un-uploaded ComputationResult from local DB")
		return err
	}

	for _, computationResultsID := range computationResultsIDs {
		// Load stored ComputationResult from BadgerDB
		computationResult, cr_err := b.store.ByID(computationResultsID)
		if cr_err != nil {
			log.Error().Err(cr_err).Msg(
				fmt.Sprintf("Failed to load ComputationResult from local DB with ID %s", computationResultsID))
			err = cr_err
			continue
		}
		if computationResult == nil || computationResult.ExecutableBlock == nil {
			errMsg := fmt.Sprintf("Invalid ComputationResult returned from local DB with ID %s", computationResultsID)
			log.Error().Err(cr_err).Msg(errMsg)
			err = errors.New(errMsg)
			continue
		}

		// Do Upload
		if cr_err = b.uploader.Upload(computationResult); cr_err != nil {
			log.Error().Err(cr_err).Msg(
				fmt.Sprintf("Failed to update ComputationResult from local DB with ID %s", computationResultsID))
			err = cr_err
		}

		b.metrics.ExecutionComputationResultUploadRetried()
	}

	// return latest occurred error
	return err
}
