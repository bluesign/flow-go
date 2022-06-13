package extract

import (
	"fmt"

	"github.com/rs/zerolog"

	mgr "github.com/onflow/flow-go/cmd/util/ledger/migrations"
	"github.com/onflow/flow-go/cmd/util/ledger/reporters"
	"github.com/onflow/flow-go/ledger"
	"github.com/onflow/flow-go/ledger/common/pathfinder"
	"github.com/onflow/flow-go/ledger/complete"
	"github.com/onflow/flow-go/ledger/complete/wal"
	"github.com/onflow/flow-go/model/bootstrap"
	"github.com/onflow/flow-go/model/flow"
	"github.com/onflow/flow-go/module/metrics"
	"github.com/onflow/flow-go/storage"
)

func getStateCommitment(commits storage.Commits, blockHash flow.Identifier) (flow.StateCommitment, error) {
	return commits.ByBlockID(blockHash)
}

func extractExecutionState(
	dir string,
	targetHash flow.StateCommitment,
	outputDir string,
	log zerolog.Logger,
	chain flow.Chain,
	migrate bool,
	report bool,
) error {

	diskWal, err := wal.NewDiskWAL(
		zerolog.Nop(),
		nil,
		metrics.NewNoopCollector(),
		dir,
		complete.DefaultCacheSize,
		pathfinder.PathByteSize,
		wal.SegmentSize,
	)
	if err != nil {
		return fmt.Errorf("cannot create disk WAL: %w", err)
	}
	defer func() {
		<-diskWal.Done()
	}()

	led, err := complete.NewLedger(
		diskWal,
		complete.DefaultCacheSize,
		&metrics.NoopCollector{},
		log,
		complete.DefaultPathFinderVersion)
	if err != nil {
		return fmt.Errorf("cannot create ledger from write-a-head logs and checkpoints: %w", err)
	}

	var migrations []ledger.Migration
	var rs map[string]ledger.Reporter
	extractionReportName := "extractionReport"
	newState := ledger.State(targetHash)

	if migrate {
		storageUsedUpdateMigration := mgr.StorageUsedUpdateMigration{
			Log:       log,
			OutputDir: outputDir,
		}

		coreContractMigration := mgr.CoreContractsMigration{
			Log:       log,
			Chain:     chain,
			OutputDir: outputDir,
		}

		migrations = []ledger.Migration{
			coreContractMigration.Migrate,
			storageUsedUpdateMigration.Migrate,
			mgr.PruneMigration,
		}

	}
	// generating reports at the end, so that the checkpoint file can be used
	// for sporking as soon as it's generated.
	if report {
		log.Info().Msgf("preparing reporter files")
		reportFileWriterFactory := reporters.NewReportFileWriterFactory(outputDir, log)

		rs = map[string]ledger.Reporter{
			// The ExportReporter needs to be run first so that it can be used
			// immediately after execution
			extractionReportName: reporters.NewExportReporter(log,
				chain,
				func() flow.StateCommitment { return targetHash },
			),
			"account": &reporters.AccountReporter{
				Log:   log,
				Chain: chain,
				RWF:   reportFileWriterFactory,
			},
			"newFungibleTokenTracker": reporters.NewFungibleTokenTracker(log, reportFileWriterFactory, chain, []string{reporters.FlowTokenTypeID(chain)}),
			"atree": &reporters.AtreeReporter{
				Log: log,
				RWF: reportFileWriterFactory,
			},
		}
	}

	migratedState, err := led.ExportCheckpointAt(
		newState,
		migrations,
		rs,
		extractionReportName,
		complete.DefaultPathFinderVersion,
		outputDir,
		bootstrap.FilenameWALRootCheckpoint,
	)
	if err != nil {
		return fmt.Errorf("cannot generate the output checkpoint: %w", err)
	}

	log.Info().Msgf(
		"New state commitment for the exported state is: %s (base64: %s)",
		migratedState.String(),
		migratedState.Base64(),
	)

	return nil
}
