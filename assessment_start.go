package main

import (
	"context"
	"errors"
	"fmt"
)

var errAssessmentBusy = errors.New("an assessment is already queued or running; wait for it to finish")
var errAssessmentCredit = errors.New("you need 1 credit to start this check")

// Verify before billing, then reserve the job, balance and ledger atomically.
func queueVerifiedAssessment(ctx context.Context, userID, domain, mode string, cost float64) (string, error) {
	if err := recheckDomainControl(ctx, userID, domain); err != nil {
		return "", err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	var domainID int64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM domains WHERE user_id=? AND domain=? AND status='verified'`, userID, domain).Scan(&domainID); err != nil {
		return "", errors.New("verify this website first")
	}
	var inflight int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM pentests WHERE user_id=? AND status IN ('reserved','queued','running')`, userID).Scan(&inflight); err != nil {
		return "", err
	}
	if inflight > 0 {
		return "", errAssessmentBusy
	}
	res, err := tx.ExecContext(ctx, `UPDATE credit_balance SET balance=balance-?, updated_at=? WHERE user_id=? AND balance>=?`, cost, now(), userID, cost)
	if err != nil {
		return "", err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return "", errAssessmentCredit
	}
	id := "pt_" + randomHex(8)
	if _, err := tx.ExecContext(ctx, `INSERT INTO pentests (id,user_id,domain_id,mode,status,authorized_at) VALUES (?,?,?,?,'queued',?)`, id, userID, domainID, mode, now()); err != nil {
		return "", err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO credit_transactions (user_id,type,amount,description,reference_id) VALUES (?,'spend',?,?,?)`, userID, cost, "Assessment scheduled", fmt.Sprintf("assessment-%s", id)); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	return id, nil
}

// A proof may disappear while a paid job waits in the queue. Close and refund
// that self-service job once, without ever handing the target to the scanner.
func failUnverifiedQueuedAssessment(ctx context.Context, id, userID, mode string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	res, err := tx.ExecContext(ctx, `UPDATE pentests SET status='failed', completed_at=?, failure_reason='website verification could not be confirmed before execution' WHERE id=? AND user_id=? AND status='queued' AND billing_source='credits'`, now(), id, userID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil
	}
	cost := 1.0
	if mode == "destructive" {
		cost = 2
	}
	if _, err := tx.ExecContext(ctx, `UPDATE credit_balance SET balance=balance+?,updated_at=? WHERE user_id=?`, cost, now(), userID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO credit_transactions(user_id,type,amount,description,reference_id) VALUES (?,'admin_credit',?,'Refund — verification proof could not be confirmed',?)`, userID, cost, id); err != nil {
		return err
	}
	return tx.Commit()
}
