package wal

import (
	"bufio"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/onflow/flow-go/ledger/complete/mtrie/flattener"
	"github.com/onflow/flow-go/ledger/complete/mtrie/node"
	"github.com/onflow/flow-go/ledger/complete/mtrie/trie"
	utilsio "github.com/onflow/flow-go/utils/io"
	"github.com/rs/zerolog"
)

const subtrieLevel = 4
const subtrieCount = 1 << subtrieLevel

type NodeEncoder func(node *trie.MTrie, index uint64, scratch []byte) []byte

type resultStoringSubTrie struct {
	Index     int
	Roots     map[*node.Node]uint64 // node index for root nodes
	NodeCount uint64
	Err       error
}

func StoreCheckpointConcurrently(tries []*trie.MTrie, outputDir string, logger *zerolog.Logger) error {

	subtrieRoots := createSubTrieRoots(tries)

	estimatedSubtrieNodeCount := estimateSubtrieNodeCount(tries)
	subTrieRootIndices, subTriesNodeCount, err := storeSubTrieConcurrently(
		subtrieRoots,
		estimatedSubtrieNodeCount,
		outputDir,
	)
	if err != nil {
		return fmt.Errorf("could not store sub trie: %w", err)
	}

	// the remaining nodes and data will be stored intot he same file
	writer, err := createWriterForTopTries(outputDir, logger)
	if err != nil {
		return fmt.Errorf("could not create writer for top tries: %w", err)
	}
	defer func() {
		closeErr := writer.Close()
		// Return close error if there isn't any prior error to return.
		if err == nil {
			err = closeErr
		}
	}()

	topLevelNodeIndices, totalNodeCount, err := storeTopLevelNodes(
		tries,
		subTrieRootIndices,
		subTriesNodeCount,
		writer)

	if err != nil {
		return fmt.Errorf("could not store top level nodes: %w", err)
	}

	err = storeRootNodes(
		tries,
		topLevelNodeIndices,
		flattener.EncodeTrie,
		writer)
	if err != nil {
		return fmt.Errorf("could not store top level nodes: %w", err)
	}

	err = storeFooter(totalNodeCount, uint16(len(tries)), writer)
	if err != nil {
		return fmt.Errorf("could not store footer: %w", err)
	}

	return nil
}

func createSubTrieRoots(tries []*trie.MTrie) [subtrieCount][]*node.Node {
	var subtrieRoots [subtrieCount][]*node.Node
	for i := 0; i < len(subtrieRoots); i++ {
		subtrieRoots[i] = make([]*node.Node, len(tries))
	}

	for trieIndex, t := range tries {
		// subtries is an array with subtrieCount trie nodes
		// in breadth-first order at subtrieLevel of the trie `t`
		subtries := getNodesAtLevel(t.RootNode(), subtrieLevel)
		for subtrieIndex, subtrieRoot := range subtries {
			subtrieRoots[subtrieIndex][trieIndex] = subtrieRoot
		}
	}
	return subtrieRoots
}

func estimateSubtrieNodeCount(tries []*trie.MTrie) int {
	if len(tries) == 0 {
		return 0
	}
	estimatedTrieNodeCount := 2*int(tries[0].AllocatedRegCount()) - 1
	return estimatedTrieNodeCount / subtrieCount
}

func storeSubTrieConcurrently(
	subtrieRoots [subtrieCount][]*node.Node,
	estimatedSubtrieNodeCount int,
	outputDir string,
) (map[*node.Node]uint64, uint64, error) {
	resultChs := make([]chan *resultStoringSubTrie, 0, len(subtrieRoots))
	for i, subTrieRoot := range subtrieRoots {
		resultCh := make(chan *resultStoringSubTrie)
		go func(i int, subTrieRoot []*node.Node) {
			roots, nodeCount, err := storeCheckpointSubTrie(i, subTrieRoot, estimatedSubtrieNodeCount, outputDir)
			resultCh <- &resultStoringSubTrie{
				Index:     i,
				Roots:     roots,
				NodeCount: nodeCount,
				Err:       err,
			}
		}(i, subTrieRoot)
		resultChs = append(resultChs, resultCh)
	}

	results := make(map[*node.Node]uint64, 1<<(subtrieLevel+1))
	results[nil] = 0
	nodeCounter := uint64(0)
	for _, resultCh := range resultChs {
		result := <-resultCh
		if result.Err != nil {
			return nil, 0, fmt.Errorf("fail to store %v-th subtrie, trie: %w", result.Index, result.Err)
		}

		for root, index := range result.Roots {
			results[root] = index
		}
		nodeCounter += result.NodeCount
	}

	return results, nodeCounter, nil
}

func createWriterForTopTries(dir string, logger *zerolog.Logger) (io.WriteCloser, error) {
	fileName := "17" // TODO: move 17 to const, define file name so that no matter checkpoint file part is stored under the folder of the checkpoint, or the folder of all checkpoints, there will be no overlap
	fullPath := path.Join(dir, fileName)
	if utilsio.FileExists(fullPath) {
		return nil, fmt.Errorf("checkpoint file for top tries %s already exists", fullPath)
	}

	return createClosableWriter(dir, logger, fileName, fullPath)
}

func createWriterForSubtrie(dir string, logger *zerolog.Logger, index int) (io.WriteCloser, error) {
	fileName := fmt.Sprintf("%v", index)
	fullPath := path.Join(dir, fileName)
	if utilsio.FileExists(fullPath) {
		return nil, fmt.Errorf("checkpoint file for %v-th sub trie %s already exists", index, fullPath)
	}

	return createClosableWriter(dir, logger, fileName, fullPath)
}

func createClosableWriter(dir string, logger *zerolog.Logger, fileName string, fullPath string) (io.WriteCloser, error) {
	tmpFile, err := os.CreateTemp(dir, fmt.Sprintf("writing-%v-*", fileName))
	if err != nil {
		return nil, fmt.Errorf("could not create temporary file for checkpoint toptries: %w", err)
	}

	writer := bufio.NewWriterSize(tmpFile, defaultBufioWriteSize)
	return &SyncOnCloseRenameFile{
		logger:     logger,
		file:       tmpFile,
		targetName: fullPath,
		Writer:     writer,
	}, nil
}

func storeCheckpointSubTrie(
	i int,
	roots []*node.Node,
	estimatedSubtrieNodeCount int,
	outputDir string) (
	map[*node.Node]uint64, uint64, error) {

	// traversedSubtrieNodes contains all unique nodes of subtries of the same path and their index.
	traversedSubtrieNodes := make(map[*node.Node]uint64, estimatedSubtrieNodeCount)
	// Index 0 is a special case with nil node.
	traversedSubtrieNodes[nil] = 0

	writer, err := createWriterForSubtrie(outputDir, nil, i)
	if err != nil {
		return nil, 0, fmt.Errorf("could not create writer for sub trie: %w", err)
	}
	defer func() {
		closeErr := writer.Close()
		// Return close error if there isn't any prior error to return.
		if err == nil {
			err = closeErr
		}
	}()

	// topLevelNodes contains all unique nodes of given tries
	// from root to subtrie root and their index
	// (ordered by node traversal sequence).
	// Index 0 is a special case with nil node.
	subtrieRootNodes := make(map[*node.Node]uint64, 1<<(subtrieLevel+1))
	subtrieRootNodes[nil] = 0

	// nodeCounter is counter for all unique nodes.
	// It starts from 1, as 0 marks nil node.
	nodeCounter := uint64(1)

	scratch := make([]byte, 1024*4)
	for _, root := range roots {
		// Note: nodeCounter is to assign an global index to each node in the order of it being seralized
		// into the checkpoint file. Therefore, it has to be reused when iterating each subtrie.
		// storeUniqueNodes will add the unique visited node into traversedSubtrieNodes with key as the node
		// itself, and value as n-th node being seralized in the checkpoint file.
		nodeCounter, err = storeUniqueNodes(root, traversedSubtrieNodes, 0, scratch, writer)
		if err != nil {
			return nil, 0, fmt.Errorf("fail to store nodes in step 1 for subtrie root %v: %w", root.Hash(), err)
		}
		// Save subtrie root node index in topLevelNodes,
		// so when traversing top level tries
		// (from level 0 to subtrieLevel) using topLevelNodes,
		// node iterator skips subtrie as visited nodes.
		subtrieRootNodes[root] = traversedSubtrieNodes[root]
	}

	return subtrieRootNodes, nodeCounter, nil
}

func storeTopLevelNodes(
	tries []*trie.MTrie,
	subTrieRootIndices map[*node.Node]uint64,
	nodeCounter uint64,
	writer io.Writer) (
	map[*node.Node]uint64,
	uint64,
	error) {
	scratch := make([]byte, 1024*4)
	var err error
	for _, t := range tries {
		root := t.RootNode()
		if root == nil {
			continue
		}
		// if we iterate through the root trie with an empty visited nodes map, then it will iterate through
		// all nodes at all levels. In order to skip the nodes above subtrieLevel, since they have been seralized in step 1,
		// we will need to pass in a visited nodes map that contains all the subtrie root nodes, which is the topLevelNodes.
		// The topLevelNodes was built in step 1, when seralizing each subtrie root nodes.
		nodeCounter, err = storeUniqueNodes(root, subTrieRootIndices, nodeCounter, scratch, writer)
		if err != nil {
			return nil, 0, fmt.Errorf("fail to store nodes in step 2 for root trie %v: %w", root.Hash(), err)
		}
	}

	return subTrieRootIndices, nodeCounter, nil
}

func storeRootNodes(
	tries []*trie.MTrie,
	topLevelNodes map[*node.Node]uint64,
	encodeNode NodeEncoder,
	writer io.Writer) error {
	scratch := make([]byte, 1024*4)
	for _, t := range tries {
		rootNode := t.RootNode()

		// Get root node index
		rootIndex, found := topLevelNodes[rootNode]
		if !found {
			rootHash := t.RootHash()
			return fmt.Errorf("internal error: missing node with hash %s", hex.EncodeToString(rootHash[:]))
		}

		encTrie := encodeNode(t, rootIndex, scratch)
		_, err := writer.Write(encTrie)
		if err != nil {
			return fmt.Errorf("cannot serialize trie: %w", err)
		}
	}

	return nil
}

func storeFooter(totalNodeCount uint64, rootTrieCount uint16, writer io.Writer) error {
	footer := encodeFooter(totalNodeCount, rootTrieCount)
	_, err := writer.Write(footer)
	if err != nil {
		return fmt.Errorf("cannot write checkpoint footer: %w", err)
	}
	return nil
}

func encodeFooter(totalNodeCount uint64, rootTrieCount uint16) []byte {
	footer := make([]byte, encNodeCountSize+encTrieCountSize)
	binary.BigEndian.PutUint64(footer, totalNodeCount-1) // -1 to account for 0 node meaning nil
	binary.BigEndian.PutUint16(footer[encNodeCountSize:], rootTrieCount)
	return footer
}
