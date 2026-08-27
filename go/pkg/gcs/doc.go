// Package gcs provides the Go client for Ray Global Control Store (GCS).
//
// The GCS client enables Go applications to access Ray cluster metadata,
// including nodes, jobs, actors, workers, and placement groups.
//
// Basic usage:
//
//	client, err := gcs.ConnectClient(gcs.ClientOptions{
//	    Address:   "localhost:6379",
//	    TimeoutMs: 10000,
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
//	// Access cluster state
//	clusterID := client.ClusterID()
//	nodes, err := client.Nodes().GetAll(context.Background(), nil)
package gcs
