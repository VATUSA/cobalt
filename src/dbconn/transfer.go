package dbconn

import (
	"context"
	"vatusa-cobalt/db"
)

type TransferRequestCombined struct {
	TransferRequest db.TransferRequest
	CombinedUserRow db.GetCombinedUserRow
}

func GetFacilityPendingTransferRequests(facility string) ([]TransferRequestCombined, error) {
	ctx := context.Background()

	transferRequests, err := Queries().GetPendingTransferRequestsForFacility(ctx, facility)
	if err != nil {
		return nil, err
	}
	var out []TransferRequestCombined
	for _, transferRequest := range transferRequests {
		// Consider refactoring this to get a list of cids and extract them all in a single query
		// Ideally, the number of transfer requests should be small enough that the performance doesn't matter
		combinedUser, err := GetCombinedUserByCID(int(transferRequest.Cid))
		if err != nil {
			return nil, err
		}
		if combinedUser == nil {
			continue
		}
		out = append(out, TransferRequestCombined{
			TransferRequest: transferRequest,
			CombinedUserRow: *combinedUser,
		})
	}

	return out, nil
}

func GetUserPendingTransferRequestsCount(cid int) (int, error) {
	ctx := context.Background()
	result, err := Queries().GetCountPendingTransferRequestsForCID(ctx, int64(cid))
	if err != nil {
		return 0, err
	}
	return int(result), nil
}
