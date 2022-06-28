package cmd

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-go/cmd/util/cmd/common"
	"github.com/onflow/flow-go/consensus/hotstuff/persister"
	"github.com/onflow/flow-go/model/flow"
)

var (
	flagView uint64
)

var Cmd = &cobra.Command{
	Use:   "set",
	Short: "set hotstuff view",
	Run:   run,
}

func init() {
	rootCmd.AddCommand(Cmd)

	Cmd.Flags().Uint64Var(&flagView, "view", 0,
		"hotstuff view")

}

type Reader struct {
	persister *persister.Persister
}

func NewReader(persister *persister.Persister) *Reader {
	return &Reader{
		persister: persister,
	}
}

func (r *Reader) SetHotstuffView(view uint64) error {
	log.Info().Msgf("setting hotstuff view to %v", view)

	err := r.persister.PutStarted(view)
	if err != nil {
		return fmt.Errorf("could not put hotstuff view %v: %w", view, err)
	}

	log.Info().Msgf("successfully set hotstuff view to %v", view)
	return nil
}

func (r *Reader) GetFinal() (*flow.Block, error) {
	header, err := r.state.Final().Head()
	if err != nil {
		return nil, fmt.Errorf("could not get finalized, %w", err)
	}

	block, err := r.getBlockByHeader(header)
	if err != nil {
		return nil, fmt.Errorf("could not get block by header: %w", err)
	}
	return block, nil
}

func (r *Reader) GetSealed() (*flow.Block, error) {
	header, err := r.state.Sealed().Head()
	if err != nil {
		return nil, fmt.Errorf("could not get sealed block, %w", err)
	}

	block, err := r.getBlockByHeader(header)
	if err != nil {
		return nil, fmt.Errorf("could not get block by header: %w", err)
	}
	return block, nil
}

func (r *Reader) GetBlockByID(blockID flow.Identifier) (*flow.Block, error) {
	header, err := r.state.AtBlockID(blockID).Head()
	if err != nil {
		return nil, fmt.Errorf("could not get header by blockID: %v, %w", blockID, err)
	}

	block, err := r.getBlockByHeader(header)
	if err != nil {
		return nil, fmt.Errorf("could not get block by header: %w", err)
	}
	return block, nil
}

func run(*cobra.Command, []string) {
	db := common.InitStorage(flagDatadir)
	defer db.Close()

	storages := common.InitStorages(db)
	state, err := common.InitProtocolState(db, storages)
	if err != nil {
		log.Fatal().Err(err).Msg("could not init protocol state")
	}

	reader := NewReader(state, storages)

	if flagHeight > 0 {
		log.Info().Msgf("get block by height: %v", flagHeight)
		block, err := reader.GetBlockByHeight(flagHeight)
		if err != nil {
			log.Fatal().Err(err).Msg("could not get block by height")
		}

		common.PrettyPrintEntity(block)
		return
	}

	if flagBlockID != "" {
		blockID, err := flow.HexStringToIdentifier(flagBlockID)
		if err != nil {
			log.Fatal().Err(err).Msgf("malformed block ID: %v", flagBlockID)
		}
		log.Info().Msgf("get block by ID: %v", blockID)
		block, err := reader.GetBlockByID(blockID)
		if err != nil {
			log.Fatal().Err(err).Msg("could not get block by ID")
		}
		common.PrettyPrintEntity(block)
		return
	}

	if flagFinal {
		log.Info().Msgf("get last finalized block")
		block, err := reader.GetFinal()
		if err != nil {
			log.Fatal().Err(err).Msg("could not get finalized block")
		}
		common.PrettyPrintEntity(block)
		return
	}

	if flagSealed {
		log.Info().Msgf("get last sealed block")
		block, err := reader.GetSealed()
		if err != nil {
			log.Fatal().Err(err).Msg("could not get sealed block")
		}
		common.PrettyPrintEntity(block)
		return
	}

	log.Fatal().Msgf("missing flag, try --final or --sealed or --height or --block-id")
}
