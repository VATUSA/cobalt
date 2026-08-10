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

const (
	TransferAccept string = "accept"
	TransferReject string = "reject"
)

func CreateTransferRequest(user db.GetCombinedUserRow, toFacility string, reason string, actor db.GetCombinedUserRow) (*db.TransferRequest, error) {
	if user.Facility == toFacility {
		return nil, errors.New("fromFacility and toFacility must not be equal")
	}
	blockers := GetUserBlockers(user)
	if blockers.IsTransferBlocked {
		return nil, errors.New("user is transfer blocked")
	}
	params := db.CreateTransferRequestParams{
		Cid:          user.Cid,
		FromFacility: user.Facility,
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
	err = action.Log(user, action.TransferRequest,
		fmt.Sprintf("Transfer request from %s to %s: %s", user.Facility, toFacility, reason), actor.Cid)
	if err != nil {
		return nil, err
	}
	return &transferRequest, nil
}

func AcceptTransferRequest(request db.TransferRequest, actorCid int64) error {
	user, err := dbconn.GetCombinedUserByCID(int(request.Cid))
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}
	err = doTransfer(*user, request.FromFacility, request.ToFacility, request.Reason, request.CreatedAt, actorCid)
	if err != nil {
		return err
	}
	queries := dbconn.Queries()
	ctx := context.Background()
	err = queries.DeleteTransferRequest(ctx, request.ID)
	if err != nil {
		return err
	}
	actor, err := dbconn.GetCombinedUserByCID(int(actorCid))
	if err != nil {
		return err
	}
	if actor == nil {
		return errors.New("actor not found")
	}
	err = action.Log(*user, action.TransferApproved,
		fmt.Sprintf("Transfer request to %s accepted by %s (%d)", request.ToFacility, actor.DisplayName, actorCid), actorCid)
	if err != nil {
		return err
	}
	return nil
}

func RejectTransferRequest(request db.TransferRequest, actorCid int64) error {
	user, err := dbconn.GetCombinedUserByCID(int(request.Cid))
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}
	ctx := context.Background()
	params := db.UpdateTransferRequestStatusParams{
		Status: "rejected",
		ID:     request.ID,
	}
	err = dbconn.Queries().UpdateTransferRequestStatus(ctx, params)
	if err != nil {
		return err
	}
	actor, err := dbconn.GetCombinedUserByCID(int(actorCid))
	if err != nil {
		return err
	}
	if actor == nil {
		return errors.New("actor not found")
	}
	err = action.Log(*user, action.TransferDenied,
		fmt.Sprintf("Transfer request to %s denied by %s (%d): %s", request.ToFacility, actor.DisplayName, actorCid, request.Reason), actorCid)
	if err != nil {
		return err
	}
	return nil
}

func ForceTransfer(user db.GetCombinedUserRow, fromFacility string, toFacility string, reason string, actorCid int64) error {
	if fromFacility == toFacility {
		return errors.New("fromFacility and toFacility must not be equal")
	}
	err := doTransfer(user, fromFacility, toFacility, reason, time.Now(), actorCid)
	if err != nil {
		return err
	}
	actor, err := dbconn.GetCombinedUserByCID(int(actorCid))
	if err != nil {
		return err
	}
	if actor == nil {
		return errors.New("actor not found")
	}
	err = action.Log(user, action.ForceTransfer,
		fmt.Sprintf("Forced transfer from %s to %s by %s (%d): %s", fromFacility, toFacility, actor.DisplayName, actorCid, reason), actorCid)
	if err != nil {
		return err
	}
	return nil
}

func doTransfer(user db.GetCombinedUserRow, fromFacility string, toFacility string, reason string, requestedAt time.Time, actorCid int64) error {
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
		Cid: user.Cid,
	})
	if err != nil {
		return err
	}

	_, err = queries.CreateTransferHistoryRecord(ctx, db.CreateTransferHistoryRecordParams{
		Cid:          user.Cid,
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
