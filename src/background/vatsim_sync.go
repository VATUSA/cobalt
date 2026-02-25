package background

import (
	"strconv"
	"vatusa-cobalt/vatsim"
)

func VatsimSync(args []string) error {
	if len(args) >= 1 {
		cid, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}
		err = vatsim.SyncByCID(cid)
		if err != nil {
			return err
		}
		return nil
	}
	err := vatsim.SyncDivisionMembers()
	if err != nil {
		return err
	}
	return nil
}
