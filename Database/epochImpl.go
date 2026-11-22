package Database

import (
	"BHLayer2Node/paradigm"
	"database/sql"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// SaveEpochReceipt always records the receipt. A nil epoch omits only the detail.
func (o DatabaseService) SaveEpochReceipt(reference *paradigm.DevReference, epoch *paradigm.DevEpoch) error {
	return o.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(reference).Error; err != nil {
			return err
		}
		if epoch == nil {
			return nil
		}
		epoch.TID = reference.TID
		return tx.Create(epoch).Error
	})
}

func emptyEpoch(id int32) *paradigm.DevEpoch {
	return &paradigm.DevEpoch{
		EpochID:    id,
		Process:    make(map[paradigm.SupportModelType]int32),
		Commits:    make(map[paradigm.SupportModelType][]paradigm.SlotHash),
		Justifieds: make(map[paradigm.SupportModelType][]paradigm.SlotHash),
		Finalizes:  make(map[paradigm.SupportModelType][]paradigm.SlotHash),
		Invalids:   []*paradigm.Slot{},
		InitTasks:  []*paradigm.Task{},
		SlotMap:    make(map[paradigm.SlotHash]*paradigm.Slot),
	}
}

func attachEpochReceipt(epoch *paradigm.DevEpoch, ref *paradigm.DevReference) {
	epoch.TID = ref.TID
	epoch.TxHash = ref.TxHash
	epoch.TxBlockHash = ref.TxBlockHash
	epoch.TxReceipt = &ref.TxReceipt
	if epoch.CreatedAt.IsZero() {
		epoch.CreatedAt = ref.UpchainTime
	}
}

func (o DatabaseService) GetEpochByID(id int32) (*paradigm.DevEpoch, error) {
	var epoch *paradigm.DevEpoch
	err := o.db.Transaction(func(tx *gorm.DB) error {
		var err error
		epoch, err = (DatabaseService{db: tx}).getEpoch(id, nil)
		return err
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	return epoch, err
}

func (o DatabaseService) getEpoch(id int32, ref *paradigm.DevReference) (*paradigm.DevEpoch, error) {
	if id < 0 {
		return nil, gorm.ErrRecordNotFound
	}
	var epoch paradigm.DevEpoch
	err := o.db.Where("epoch_id = ?", id).First(&epoch).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if ref == nil {
			maxID, err := o.GetMaxEpochID()
			if err != nil {
				return nil, err
			}
			if id > maxID {
				return nil, gorm.ErrRecordNotFound
			}
			var found paradigm.DevReference
			err = o.db.Where("epoch_id = ? AND rf = ?", id, paradigm.EpochTx).Order("t_id DESC").Take(&found).Error
			if err == nil {
				ref = &found
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, err
			}
		}
		empty := emptyEpoch(id)
		if ref != nil {
			attachEpochReceipt(empty, ref)
		}
		return empty, nil
	}
	if err != nil {
		return nil, err
	}
	var storedRef paradigm.DevReference
	if err := o.db.Take(&storedRef, epoch.TID).Error; err != nil {
		return nil, fmt.Errorf("failed to get associated transaction: %w", err)
	}
	attachEpochReceipt(&epoch, &storedRef)
	epoch.SlotMap = make(map[paradigm.SlotHash]*paradigm.Slot)
	var hashes []paradigm.SlotHash
	for _, group := range []map[paradigm.SupportModelType][]paradigm.SlotHash{epoch.Commits, epoch.Justifieds, epoch.Finalizes} {
		for _, slotHashes := range group {
			hashes = append(hashes, slotHashes...)
		}
	}
	if len(hashes) > 0 {
		var slots []*paradigm.Slot
		if err := o.db.Where("slot_id IN ?", hashes).Find(&slots).Error; err != nil {
			return nil, fmt.Errorf("failed to query slots: %w", err)
		}
		for _, slot := range slots {
			epoch.SlotMap[slot.SlotID] = slot
		}
	}
	return &epoch, nil
}

func (o DatabaseService) GetEpochByTxHash(hash string) (*paradigm.DevEpoch, error) {
	ref, err := o.GetTransactionByHash(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}
	if ref.Rf != paradigm.EpochTx {
		return nil, fmt.Errorf("transaction is not an epoch transaction")
	}
	if ref.EpochID < 0 {
		return nil, fmt.Errorf("invalid epoch ID in transaction")
	}
	return o.getEpoch(ref.EpochID, ref)
}

func (o DatabaseService) GetLatestEpochs(limit int) ([]*paradigm.DevEpoch, error) {
	epochs := make([]*paradigm.DevEpoch, 0)
	if limit <= 0 {
		return epochs, nil
	}
	err := o.db.Transaction(func(tx *gorm.DB) error {
		reader := DatabaseService{db: tx}
		maxID, err := reader.GetMaxEpochID()
		if err != nil || maxID < 0 {
			return err
		}
		minID := int32(max(0, int(maxID)-(limit-1)))
		var details []*paradigm.DevEpoch
		if err := tx.Where("epoch_id BETWEEN ? AND ?", minID, maxID).Find(&details).Error; err != nil {
			return err
		}
		byID := make(map[int32]*paradigm.DevEpoch, len(details))
		tids := make([]int64, 0, len(details))
		for _, epoch := range details {
			byID[epoch.EpochID] = epoch
			tids = append(tids, epoch.TID)
		}
		var refs []*paradigm.DevReference
		refQuery := tx.Where("rf = ? AND epoch_id BETWEEN ? AND ?", paradigm.EpochTx, minID, maxID)
		if len(tids) > 0 {
			refQuery = refQuery.Or("t_id IN ?", tids)
		}
		if err := refQuery.Order("t_id ASC").Find(&refs).Error; err != nil {
			return err
		}
		refByID := make(map[int32]*paradigm.DevReference)
		refByTID := make(map[int64]*paradigm.DevReference)
		for _, ref := range refs {
			refByTID[ref.TID] = ref
			if ref.Rf == paradigm.EpochTx {
				refByID[ref.EpochID] = ref
			}
		}
		for id := maxID; id >= minID; id-- {
			epoch := byID[id]
			if epoch == nil {
				epoch = emptyEpoch(id)
				if ref := refByID[id]; ref != nil {
					attachEpochReceipt(epoch, ref)
				}
			} else if ref := refByTID[epoch.TID]; ref != nil {
				attachEpochReceipt(epoch, ref)
			}
			epochs = append(epochs, epoch)
		}
		return nil
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	return epochs, err
}

// GetMaxEpochID includes epochs with a receipt but no detail, and legacy details.
func (o DatabaseService) GetMaxEpochID() (int32, error) {
	var id int32
	err := o.db.Raw(`SELECT GREATEST(
		COALESCE((SELECT MAX(epoch_id) FROM dev_references WHERE rf = ?), -1),
		COALESCE((SELECT MAX(epoch_id) FROM dev_epoches), -1)
	)`, paradigm.EpochTx).Scan(&id).Error
	return id, err
}
