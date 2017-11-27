package helpers

import (
	"database/sql"
	"database/sql/driver"
	"time"

	"code.cloudfoundry.org/lager"
)

// BEGIN TRANSACTION; f ... ; COMMIT; or
// BEGIN TRANSACTION; f ... ; ROLLBACK; if f returns an error.
func (h *sqlHelper) Transact(logger lager.Logger, db *sql.DB, f func(logger lager.Logger, tx *sql.Tx) error) error {
	var err error

	for attempts := 0; attempts < 3; attempts++ {
		err = func() error {
			tx, err := db.Begin()
			if err != nil {
				logger.Error("failed-starting-transaction", err)
				return err
			}
			defer tx.Rollback()

			err = f(logger, tx)
			if err != nil {
				return err
			}

			err = tx.Commit()
			if err != nil {
				logger.Error("failed-committing-transaction", err)

			}
			return err
		}()

		convertedErr := h.ConvertSQLError(err)
		// golang sql package does not always retry query on ErrBadConn, e.g. if it
		// is in the middle of a transaction. This make sense since the package
		// cannot retry the entire transaction and has to return control to the
		// caller to initiate a retry
		if attempts >= 2 || (convertedErr != ErrDeadlock && convertedErr != driver.ErrBadConn) {
			break
		} else {
			logger.Error("deadlock-transaction", err, lager.Data{"attempts": attempts})
			time.Sleep(500 * time.Millisecond)
		}
	}

	return err
}
