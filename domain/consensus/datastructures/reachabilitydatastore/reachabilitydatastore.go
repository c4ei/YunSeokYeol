package reachabilitydatastore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/dbkeys"
)

var reachabilityDataBucket = dbkeys.MakeBucket([]byte("reachability-data"))
var reachabilityReindexRootKey = dbkeys.MakeBucket().Key([]byte("reachability-reindex-root"))

// reachabilityDataStore represents a store of ReachabilityData
type reachabilityDataStore struct {
	reachabilityDataStaging        map[externalapi.DomainHash]*model.ReachabilityData
	reachabilityReindexRootStaging *externalapi.DomainHash
}

// New instantiates a new ReachabilityDataStore
func New() model.ReachabilityDataStore {
	return &reachabilityDataStore{
		reachabilityDataStaging:        make(map[externalapi.DomainHash]*model.ReachabilityData),
		reachabilityReindexRootStaging: nil,
	}
}

// StageReachabilityData stages the given reachabilityData for the given blockHash
func (rds *reachabilityDataStore) StageReachabilityData(blockHash *externalapi.DomainHash, reachabilityData *model.ReachabilityData) {
	rds.reachabilityDataStaging[*blockHash] = reachabilityData
}

// StageReachabilityReindexRoot stages the given reachabilityReindexRoot
func (rds *reachabilityDataStore) StageReachabilityReindexRoot(reachabilityReindexRoot *externalapi.DomainHash) {
	rds.reachabilityReindexRootStaging = reachabilityReindexRoot
}

func (rds *reachabilityDataStore) IsAnythingStaged() bool {
	return len(rds.reachabilityDataStaging) != 0 || rds.reachabilityReindexRootStaging != nil
}

func (rds *reachabilityDataStore) Discard() {
	rds.reachabilityDataStaging = make(map[externalapi.DomainHash]*model.ReachabilityData)
	rds.reachabilityReindexRootStaging = nil
}

func (rds *reachabilityDataStore) Commit(dbTx model.DBTransaction) error {
	if rds.reachabilityReindexRootStaging != nil {
		err := dbTx.Put(reachabilityReindexRootKey, rds.serializeReachabilityReindexRoot(rds.reachabilityReindexRootStaging))
		if err != nil {
			return err
		}
	}
	for hash, reachabilityData := range rds.reachabilityDataStaging {
		err := dbTx.Put(rds.reachabilityDataBlockHashAsKey(&hash), rds.serializeReachabilityData(reachabilityData))
		if err != nil {
			return err
		}
	}

	rds.Discard()
	return nil
}

// ReachabilityData returns the reachabilityData associated with the given blockHash
func (rds *reachabilityDataStore) ReachabilityData(dbContext model.DBReader,
	blockHash *externalapi.DomainHash) (*model.ReachabilityData, error) {

	if reachabilityData, ok := rds.reachabilityDataStaging[*blockHash]; ok {
		return reachabilityData, nil
	}

	reachabilityDataBytes, err := dbContext.Get(rds.reachabilityDataBlockHashAsKey(blockHash))
	if err != nil {
		return nil, err
	}

	return rds.deserializeReachabilityData(reachabilityDataBytes)
}

// ReachabilityReindexRoot returns the current reachability reindex root
func (rds *reachabilityDataStore) ReachabilityReindexRoot(dbContext model.DBReader) (*externalapi.DomainHash, error) {
	if rds.reachabilityReindexRootStaging != nil {
		return rds.reachabilityReindexRootStaging, nil
	}
	reachabilityReindexRootBytes, err := dbContext.Get(reachabilityReindexRootKey)
	if err != nil {
		return nil, err
	}

	reachabilityReindexRoot, err := rds.deserializeReachabilityReindexRoot(reachabilityReindexRootBytes)
	if err != nil {
		return nil, err
	}
	return reachabilityReindexRoot, nil
}

func (rds *reachabilityDataStore) reachabilityDataBlockHashAsKey(hash *externalapi.DomainHash) model.DBKey {
	return reachabilityDataBucket.Key(hash[:])
}

func (rds *reachabilityDataStore) serializeReachabilityData(reachabilityData *model.ReachabilityData) []byte {
	panic("implement me")
}

func (rds *reachabilityDataStore) deserializeReachabilityData(reachabilityDataBytes []byte) (*model.ReachabilityData, error) {
	panic("implement me")
}

func (rds *reachabilityDataStore) serializeReachabilityReindexRoot(reachabilityReindexRoot *externalapi.DomainHash) []byte {
	panic("implement me")
}

func (rds *reachabilityDataStore) deserializeReachabilityReindexRoot(reachabilityReindexRootBytes []byte) (*externalapi.DomainHash, error) {
	panic("implement me")
}