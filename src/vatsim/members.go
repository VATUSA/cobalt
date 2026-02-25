package vatsim

import (
	"encoding/json"
	"fmt"
)

func GetMemberDataByCid(cid int) (*MemberData, error) {
	uri := fmt.Sprintf("/members/%d", cid)
	data, err := API2Request(uri, "GET", nil)
	if err != nil {
		return nil, err
	}

	var memberData MemberData
	err = json.Unmarshal(data, &memberData)
	if err != nil {
		return nil, err
	}
	return &memberData, nil
}

func GetDivisionMembersPage(page int) (*DivisionMemberData, error) {
	recordsPerPage := 1000
	limit := recordsPerPage
	offset := (page - 1) * recordsPerPage
	uri := fmt.Sprintf("/orgs/division/USA?limit=%d&offset=%d", limit, offset)
	data, err := API2Request(uri, "GET", nil)
	if err != nil {
		return nil, err
	}

	var divisionMemberData DivisionMemberData
	err = json.Unmarshal(data, &divisionMemberData)
	if err != nil {
		return nil, err
	}

	return &divisionMemberData, nil
}
