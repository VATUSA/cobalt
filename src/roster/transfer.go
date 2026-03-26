package roster

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
	"vatusa-cobalt/action"
	"vatusa-cobalt/db"
	"vatusa-cobalt/dbconn"
)

func CreateTransferRequest(cid int64, fromFacility string, toFacility string, reason string, actorCid int64) (*db.TransferRequest, error) {
	if fromFacility == toFacility {
		return nil, errors.New("fromFacility and toFacility must not be equal")
	}
	params := db.CreateTransferRequestParams{
		Cid:          cid,
		FromFacility: fromFacility,
		ToFacility:   toFacility,
		Reason:       reason,
	}
	queries := dbconn.Queries()
	ctx := context.Background()

	res, err := queries.CreateTransferRequest(ctx, params)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	transferRequest, err := queries.GetTransferRequestById(ctx, id)
	if err != nil {
		return nil, err
	}
	err = action.Log(cid, action.TransferRequest,
		fmt.Sprintf("%s -> %s (%s)", fromFacility, toFacility, reason), actorCid)
	if err != nil {
		return nil, err
	}
	return &transferRequest, nil
}

func AcceptTransferRequest(request db.TransferRequest, actorCid int64) error {
	err := doTransfer(request.Cid, request.FromFacility, request.ToFacility, request.Reason, request.CreatedAt, actorCid)
	if err != nil {
		return err
	}
	queries := dbconn.Queries()
	ctx := context.Background()
	err = queries.DeleteTransferRequest(ctx, request.ID)
	if err != nil {
		return err
	}
	err = action.Log(request.Cid, action.TransferApproved,
		fmt.Sprintf("%s -> %s", request.FromFacility, request.ToFacility), actorCid)
	if err != nil {
		return err
	}
	return nil
}

func ForceTransfer(cid int64, fromFacility string, toFacility string, reason string, actorCid int64) error {
	if fromFacility == toFacility {
		return errors.New("fromFacility and toFacility must not be equal")
	}
	err := doTransfer(cid, fromFacility, toFacility, reason, time.Now(), actorCid)
	if err != nil {
		return err
	}
	err = action.Log(cid, action.ForceTransfer,
		fmt.Sprintf("%s -> %s (%s)", fromFacility, toFacility, reason), actorCid)
	if err != nil {
		return err
	}
	return nil
}

func doTransfer(cid int64, fromFacility string, toFacility string, reason string, requestedAt time.Time, actorCid int64) error {
	if fromFacility == toFacility {
		return errors.New("fromFacility and toFacility must not be equal")
	}
	transferTime := time.Now()
	queries := dbconn.Queries()
	ctx := context.Background()

	err := queries.UpdateUserForTransfer(ctx, db.UpdateUserForTransferParams{
		Facility: toFacility,
		LastTransferTime: sql.NullTime{
			Time:  transferTime,
			Valid: true,
		},
		Cid: cid,
	})
	if err != nil {
		return err
	}

	_, err = queries.CreateTransferHistoryRecord(ctx, db.CreateTransferHistoryRecordParams{
		Cid:          cid,
		FromFacility: fromFacility,
		ToFacility:   toFacility,
		Reason:       reason,
		RequestedAt:  requestedAt,
		ApproverCid: sql.NullInt64{
			Int64: actorCid,
			Valid: true,
		},
	})
	if err != nil {
		return err
	}
	return nil
}
