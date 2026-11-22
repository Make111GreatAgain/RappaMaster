package Database

import (
	"BHLayer2Node/paradigm"
	"context"
	"time"
)

const emptyEpochCleanupPredicate = `
	dev_epoches.created_at < ?
	AND JSON_TYPE(dev_epoches.process) = 'OBJECT' AND JSON_LENGTH(dev_epoches.process) = 0
	AND JSON_TYPE(dev_epoches.commits) = 'OBJECT' AND JSON_LENGTH(dev_epoches.commits) = 0
	AND JSON_TYPE(dev_epoches.justifieds) = 'OBJECT' AND JSON_LENGTH(dev_epoches.justifieds) = 0
	AND JSON_TYPE(dev_epoches.finalizes) = 'OBJECT' AND JSON_LENGTH(dev_epoches.finalizes) = 0
	AND JSON_TYPE(dev_epoches.invalids) = 'ARRAY' AND JSON_LENGTH(dev_epoches.invalids) = 0
	AND JSON_TYPE(dev_epoches.init_tasks) = 'ARRAY' AND JSON_LENGTH(dev_epoches.init_tasks) = 0
	AND EXISTS (
		SELECT 1 FROM dev_references AS r
		WHERE r.t_id = dev_epoches.t_id AND r.rf = ? AND r.epoch_id = dev_epoches.epoch_id
	)`

type EmptyEpochCleanupPreview struct {
	Eligible   int64
	DataBytes  int64
	IndexBytes int64
	FreeBytes  int64
}

func (o DatabaseService) PreviewEmptyEpochCleanup(ctx context.Context, cutoff time.Time) (EmptyEpochCleanupPreview, error) {
	db := o.db.WithContext(ctx)
	var result EmptyEpochCleanupPreview
	if err := db.Table("dev_epoches").Where(emptyEpochCleanupPredicate, cutoff, paradigm.EpochTx).
		Count(&result.Eligible).Error; err != nil {
		return result, err
	}
	err := db.Raw(`SELECT COALESCE(data_length, 0) AS data_bytes,
		COALESCE(index_length, 0) AS index_bytes, COALESCE(data_free, 0) AS free_bytes
		FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'dev_epoches'`).
		Scan(&result).Error
	return result, err
}

// DeleteEmptyEpochBatch deletes at most 1000 old empty details in one statement.
func (o DatabaseService) DeleteEmptyEpochBatch(ctx context.Context, cutoff time.Time) (int64, error) {
	result := o.db.WithContext(ctx).Exec(
		"DELETE FROM dev_epoches WHERE "+emptyEpochCleanupPredicate+" ORDER BY epoch_id LIMIT 1000",
		cutoff, paradigm.EpochTx,
	)
	return result.RowsAffected, result.Error
}
