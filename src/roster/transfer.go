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

	systemActorDisplayName = "Automated"
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
	ctx := context.Background()

	return dbconn.WithTransactionResult(ctx, func(q *db.Queries) (*db.TransferRequest, error) {
		res, err := q.CreateTransferRequest(ctx, params)
		if err != nil {
			return nil, err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return nil, err
		}

		transferRequest, err := q.GetTransferRequestById(ctx, id)
		if err != nil {
			return nil, err
		}
		err = action.Log(q, user, action.TransferRequest,
			fmt.Sprintf("Transfer request from %s to %s: %s", user.Facility, toFacility, reason), actor.Cid)
		if err != nil {
			return nil, err
		}
		return &transferRequest, nil
	})
}

func AcceptTransferRequest(request db.TransferRequest, actorCid int64) error {
	user, err := dbconn.GetCombinedUserByCID(int(request.Cid))
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}
	actorName, err := resolveActorDisplayName(actorCid)
	if err != nil {
		return err
	}
	ctx := context.Background()
	return dbconn.WithTransaction(ctx, func(q *db.Queries) error {
		err := doTransfer(q, *user, request.FromFacility, request.ToFacility, request.Reason, request.CreatedAt, actorCid)
		if err != nil {
			return err
		}
		err = q.DeleteTransferRequest(ctx, request.ID)
		if err != nil {
			return err
		}
		return action.Log(q, *user, action.TransferApproved,
			fmt.Sprintf("Transfer request to %s accepted by %s (%d)", request.ToFacility, actorName, actorCid), actorCid)
	})
}

func RejectTransferRequest(request db.TransferRequest, actorCid int64) error {
	user, err := dbconn.GetCombinedUserByCID(int(request.Cid))
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}
	actorName, err := resolveActorDisplayName(actorCid)
	if err != nil {
		return err
	}
	ctx := context.Background()
	params := db.UpdateTransferRequestStatusParams{
		Status: "rejected",
		ID:     request.ID,
	}
	return dbconn.WithTransaction(ctx, func(q *db.Queries) error {
		err := q.UpdateTransferRequestStatus(ctx, params)
		if err != nil {
			return err
		}
		return action.Log(q, *user, action.TransferDenied,
			fmt.Sprintf("Transfer request to %s denied by %s (%d): %s", request.ToFacility, actorName, actorCid, request.Reason), actorCid)
	})
}

func ForceTransfer(user db.GetCombinedUserRow, fromFacility string, toFacility string, reason string, actorCid int64) error {
	if fromFacility == toFacility {
		return errors.New("fromFacility and toFacility must not be equal")
	}
	actorName, err := resolveActorDisplayName(actorCid)
	if err != nil {
		return err
	}
	ctx := context.Background()
	return dbconn.WithTransaction(ctx, func(q *db.Queries) error {
		err := doTransfer(q, user, fromFacility, toFacility, reason, time.Now(), actorCid)
		if err != nil {
			return err
		}
		return action.Log(q, user, action.ForceTransfer,
			fmt.Sprintf("Forced transfer from %s to %s by %s (%d): %s", fromFacility, toFacility, actorName, actorCid, reason), actorCid)
	})
}

func resolveActorDisplayName(actorCid int64) (string, error) {
	if actorCid == 0 {
		return systemActorDisplayName, nil
	}
	actor, err := dbconn.GetCombinedUserByCID(int(actorCid))
	if err != nil {
		return "", err
	}
	if actor == nil {
		return "", errors.New("actor not found")
	}
	return actor.DisplayName, nil
}

func doTransfer(q *db.Queries, user db.GetCombinedUserRow, fromFacility string, toFacility string, reason string, requestedAt time.Time, actorCid int64) error {
	if fromFacility == toFacility {
		return errors.New("fromFacility and toFacility must not be equal")
	}
	transferTime := time.Now()
	ctx := context.Background()

	err := q.UpdateUserForTransfer(ctx, db.UpdateUserForTransferParams{
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

	_, err = q.CreateTransferHistoryRecord(ctx, db.CreateTransferHistoryRecordParams{
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
