package sdk

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
)

type NodeStatus struct {
	ID        string  `json:"id"`
	Type      int     `json:"type"`
	Status    int     `json:"status"`
	ClusterID string  `json:"clusterId"`
	NodeID    string  `json:"nodeId"`
	Message   string  `json:"message,omitempty"`
	Result    string  `json:"result,omitempty"`
	Percent   float32 `json:"percent,omitempty"`
}

func NewStatusReport(statusHost string, id string, typ int) func(statue int, message string, result string, percent float32) error {
	gns := NodeStatus{
		ID:        id,
		Type:      typ,
		ClusterID: os.Getenv("CLUSTER_ID"),
		NodeID:    os.Getenv("NODE_ID"),
	}
	return func(statue int, message, result string, percent float32) error {
		ns := gns

		ns.Status = statue
		ns.Message = message
		ns.Result = result
		ns.Percent = percent

		jsonData, err := json.Marshal(ns)
		if err != nil {
			return err
		}

		resp, err := http.Post(
			fmt.Sprintf("http://%s/coordinator/outer/set-worker-info", statusHost),
			"application/json",
			bytes.NewBuffer(jsonData),
		)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return errors.New(fmt.Sprintf("http status code: %d", resp.StatusCode))
		}

		// var r map[string]interface{}
		// err = json.NewDecoder(resp.Body).Decode(&r)
		// if err != nil {
		// 	return err
		// }

		return nil
	}
}
