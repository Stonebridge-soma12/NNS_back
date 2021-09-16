package trainMonitor

import (
	"net/http"
)

func PostEpochHandler(r *http.Request, er EpochRepository) error {
	var epoch Epoch
	err := epoch.Bind(r)
	if err != nil {
		return err
	}

	err = er.Insert(epoch)
	if err != nil {
		return err
	}

	return nil
}

func PostLogHandler(r *http.Request, er TrainLogRepository) error {
	var trainLog TrainLog
	err := trainLog.Bind(r)
	if err != nil {
		return err
	}

	err = er.Insert(trainLog)
	if err != nil {
		return err
	}

	return nil
}