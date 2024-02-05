package migrations

import (
	"context"
	"fmt"
	"io"

	"github.com/onflow/cadence/migrations/statictypes"
	"github.com/rs/zerolog"

	"github.com/onflow/cadence/migrations"
	"github.com/onflow/cadence/migrations/capcons"
	"github.com/onflow/cadence/migrations/entitlements"
	"github.com/onflow/cadence/migrations/string_normalization"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/interpreter"

	"github.com/onflow/flow-go/cmd/util/ledger/reporters"
	"github.com/onflow/flow-go/cmd/util/ledger/util"
	"github.com/onflow/flow-go/fvm/environment"
	"github.com/onflow/flow-go/fvm/tracing"
	"github.com/onflow/flow-go/ledger"
)

type CadenceBaseMigrator struct {
	name            string
	log             zerolog.Logger
	reporter        reporters.ReportWriter
	valueMigrations func(
		inter *interpreter.Interpreter,
		accounts environment.Accounts,
		reporter *cadenceValueMigrationReporter,
	) []migrations.ValueMigration
}

var _ AccountBasedMigration = (*CadenceBaseMigrator)(nil)
var _ io.Closer = (*CadenceBaseMigrator)(nil)

func (m *CadenceBaseMigrator) Close() error {
	// Close the report writer so it flushes to file.
	m.reporter.Close()
	return nil
}

func (m *CadenceBaseMigrator) InitMigration(
	log zerolog.Logger,
	_ []*ledger.Payload,
	_ int,
) error {
	m.log = log.With().Str("migration", m.name).Logger()
	return nil
}

func (m *CadenceBaseMigrator) MigrateAccount(
	_ context.Context,
	address common.Address,
	oldPayloads []*ledger.Payload,
) ([]*ledger.Payload, error) {

	// Create all the runtime components we need for the migration
	migrationRuntime, err := newMigratorRuntime(
		address,
		oldPayloads,
		util.RuntimeInterfaceConfig{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrator runtime: %w", err)
	}

	migration := migrations.NewStorageMigration(
		migrationRuntime.Interpreter,
		migrationRuntime.Storage,
	)

	reporter := newValueMigrationReporter(m.reporter, m.log)

	m.log.Info().Msg("Migrating cadence values")

	migration.Migrate(
		&migrations.AddressSliceIterator{
			Addresses: []common.Address{
				address,
			},
		},
		migration.NewValueMigrationsPathMigrator(
			reporter,
			m.valueMigrations(migrationRuntime.Interpreter, migrationRuntime.Accounts, reporter)...,
		),
	)

	m.log.Info().Msg("Committing changes")
	err = migration.Commit()
	if err != nil {
		return nil, fmt.Errorf("failed to commit changes: %w", err)
	}

	// finalize the transaction
	result, err := migrationRuntime.TransactionState.FinalizeMainTransaction()
	if err != nil {
		return nil, fmt.Errorf("failed to finalize main transaction: %w", err)
	}

	// Merge the changes to the original payloads.
	return MergeRegisterChanges(
		migrationRuntime.Snapshot.Payloads,
		result.WriteSet,
		m.log,
	)
}

func NewCadenceValueMigrator(
	rwf reporters.ReportWriterFactory,
	capabilityIDs map[interpreter.AddressPath]interpreter.UInt64Value,
	compositeTypeConverter statictypes.CompositeTypeConverterFunc,
	interfaceTypeConverter statictypes.InterfaceTypeConverterFunc,
) *CadenceBaseMigrator {
	return &CadenceBaseMigrator{
		name:     "cadence-value-migration",
		reporter: rwf.ReportWriter("cadence-value-migrator"),
		valueMigrations: func(
			inter *interpreter.Interpreter,
			_ environment.Accounts,
			reporter *cadenceValueMigrationReporter,
		) []migrations.ValueMigration {
			// All cadence migrations except the `capcons.LinkValueMigration`.
			return []migrations.ValueMigration{
				&capcons.CapabilityValueMigration{
					CapabilityIDs: capabilityIDs,
					Reporter:      reporter,
				},
				entitlements.NewEntitlementsMigration(inter),
				string_normalization.NewStringNormalizingMigration(),
				statictypes.NewStaticTypeMigration().
					WithCompositeTypeConverter(compositeTypeConverter).
					WithInterfaceTypeConverter(interfaceTypeConverter),
			}
		},
	}
}

func NewCadenceLinkValueMigrator(
	rwf reporters.ReportWriterFactory,
	capabilityIDs map[interpreter.AddressPath]interpreter.UInt64Value,
) *CadenceBaseMigrator {
	return &CadenceBaseMigrator{
		name:     "cadence-link-value-migration",
		reporter: rwf.ReportWriter("cadence-link-value-migrator"),
		valueMigrations: func(
			_ *interpreter.Interpreter,
			accounts environment.Accounts,
			reporter *cadenceValueMigrationReporter,
		) []migrations.ValueMigration {
			idGenerator := environment.NewAccountLocalIDGenerator(
				tracing.NewMockTracerSpan(),
				util.NopMeter{},
				accounts,
			)
			return []migrations.ValueMigration{
				&capcons.LinkValueMigration{
					CapabilityIDs:      capabilityIDs,
					AccountIDGenerator: idGenerator,
					Reporter:           reporter,
				},
			}
		},
	}
}

// cadenceValueMigrationReporter is the reporter for cadence value migrations
type cadenceValueMigrationReporter struct {
	rw  reporters.ReportWriter
	log zerolog.Logger
}

var _ capcons.LinkMigrationReporter = &cadenceValueMigrationReporter{}
var _ capcons.CapabilityMigrationReporter = &cadenceValueMigrationReporter{}
var _ migrations.Reporter = &cadenceValueMigrationReporter{}

func newValueMigrationReporter(rw reporters.ReportWriter, log zerolog.Logger) *cadenceValueMigrationReporter {
	return &cadenceValueMigrationReporter{
		rw:  rw,
		log: log,
	}
}

func (t *cadenceValueMigrationReporter) Migrated(
	storageKey interpreter.StorageKey,
	storageMapKey interpreter.StorageMapKey,
	migration string,
) {
	t.rw.Write(cadenceValueMigrationReportEntry{
		StorageKey:    storageKey,
		StorageMapKey: storageMapKey,
		Migration:     migration,
	})
}

func (t *cadenceValueMigrationReporter) Error(
	storageKey interpreter.StorageKey,
	storageMapKey interpreter.StorageMapKey,
	migration string,
	err error,
) {
	t.log.Error().Msgf(
		"failed to run %s in account %s, domain %s, key %s: %s",
		migration,
		storageKey.Address,
		storageKey.Key,
		storageMapKey,
		err,
	)
}

func (t *cadenceValueMigrationReporter) MigratedPathCapability(
	accountAddress common.Address,
	addressPath interpreter.AddressPath,
	borrowType *interpreter.ReferenceStaticType,
) {
	t.rw.Write(capConsPathCapabilityMigration{
		AccountAddress: accountAddress,
		AddressPath:    addressPath,
		BorrowType:     borrowType,
	})
}

func (t *cadenceValueMigrationReporter) MissingCapabilityID(
	accountAddress common.Address,
	addressPath interpreter.AddressPath,
) {
	t.rw.Write(capConsMissingCapabilityID{
		AccountAddress: accountAddress,
		AddressPath:    addressPath,
	})
}

func (t *cadenceValueMigrationReporter) MigratedLink(
	accountAddressPath interpreter.AddressPath,
	capabilityID interpreter.UInt64Value,
) {
	t.rw.Write(capConsLinkMigration{
		AccountAddressPath: accountAddressPath,
		CapabilityID:       capabilityID,
	})
}

func (t *cadenceValueMigrationReporter) CyclicLink(err capcons.CyclicLinkError) {
	t.rw.Write(err)
}

func (t *cadenceValueMigrationReporter) MissingTarget(accountAddressPath interpreter.AddressPath) {
	t.rw.Write(capConsMissingTarget{
		AddressPath: accountAddressPath,
	})
}

type cadenceValueMigrationReportEntry struct {
	StorageKey    interpreter.StorageKey    `json:"storageKey"`
	StorageMapKey interpreter.StorageMapKey `json:"storageMapKey"`
	Migration     string                    `json:"migration"`
}

type capConsLinkMigration struct {
	AccountAddressPath interpreter.AddressPath `json:"address"`
	CapabilityID       interpreter.UInt64Value `json:"capabilityID"`
}

type capConsPathCapabilityMigration struct {
	AccountAddress common.Address                   `json:"address"`
	AddressPath    interpreter.AddressPath          `json:"addressPath"`
	BorrowType     *interpreter.ReferenceStaticType `json:"borrowType"`
}

type capConsMissingCapabilityID struct {
	AccountAddress common.Address          `json:"address"`
	AddressPath    interpreter.AddressPath `json:"addressPath"`
}

type capConsMissingTarget struct {
	AddressPath interpreter.AddressPath `json:"addressPath"`
}
