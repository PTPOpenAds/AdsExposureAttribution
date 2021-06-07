/*
 * @Author: xiaoyangma@tencent.com
 * @Date: 2021-06-02 11:14:49
 * @Last Modified by: xiaoyangma
 * @Last Modified time: 2021-06-02 11:14:49
 */

package utility

import (
	"database/sql"

	"github.com/TencentAd/attribution/attribution/pkg/modules/contract_ad/config"
	"github.com/golang/glog"
)

type JobStatus struct {
	AttributionID string
	Stage         string
	ProcessRate   int
}

func parseResult(result sql.Result, err error) error {
	if err != nil {
		glog.Info("Insert data error: ", err)
		return err
	}
	ID, _ := result.LastInsertId()
	i, _ := result.RowsAffected()
	glog.Infof("last insert id：%d , affected lines：%d \n", ID, i)
	return nil
}

func UpdateJobStatus(attributionID string, stage string, rate int) error {
	connectionStr := config.Configuration.DB["location_user_properties"]
	db, err := sql.Open("mysql", connectionStr)
	if err != nil {
		glog.Errorf("[UpdateJobStatus] connect to db failed %v", err)
	}
	return parseResult(db.Exec(config.Configuration.SQL["insert_job_status"], attributionID, stage, rate))
}

func GetJobStatus(attributionID string) (*JobStatus, error) {
	connectionStr := config.Configuration.DB["location_user_properties"]
	db, err := sql.Open("mysql", connectionStr)
	if err != nil {
		glog.Errorf("[GetJobStatus] connect to db failed %v", err)
	}

	rows := db.QueryRow(config.Configuration.SQL["select_job_status"], attributionID)
	status := &JobStatus{
		AttributionID: attributionID,
	}
	err = rows.Scan(&status.Stage, &status.ProcessRate)
	if err == sql.ErrNoRows {
		glog.Errorf("[GetJobStatus] No result for attributionID: %s\n", attributionID)
		return nil, err
	}
	glog.Info("[GetJobStatus] Status", status)
	return status, nil
}
