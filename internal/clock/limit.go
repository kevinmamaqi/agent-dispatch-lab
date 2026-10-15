package clock

import "errors"

func fmtError() error {
	return errors.New("simulation exceeded 5,000,000 events; refusing a partial report")
}
