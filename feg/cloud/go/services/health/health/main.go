/*
Copyright 2020 The Magma Authors.

This source code is licensed under the BSD-style license found in the
LICENSE file in the root directory of this source tree.

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"time"

	"github.com/golang/glog"

	"magma/feg/cloud/go/feg"
	"magma/feg/cloud/go/protos"
	"magma/feg/cloud/go/services/health"
	health_protos "magma/feg/cloud/go/services/health/protos"
	"magma/feg/cloud/go/services/health/reporter"
	protected "magma/feg/cloud/go/services/health/servicers/protected"
	southbound "magma/feg/cloud/go/services/health/servicers/southbound"
	"magma/orc8r/cloud/go/blobstore"
	"magma/orc8r/cloud/go/service"
	"magma/orc8r/cloud/go/sqorc"
	"magma/orc8r/cloud/go/storage"
)

const (
	NETWORK_HEALTH_STATUS_REPORT_INTERVAL = time.Second * 60
)

func main() {
	// Create the service
	srv, err := service.NewOrchestratorService(feg.ModuleName, health.ServiceName)
	if err != nil {
		glog.Fatalf("Error creating service: %+v", err)
	}
	db, err := sqorc.Open(storage.GetSQLDriver(), storage.GetDatabaseSource())
	if err != nil {
		glog.Fatalf("Failed to connect to database: %+v", err)
	}
	store := blobstore.NewSQLStoreFactory(health.DBTableName, db, sqorc.GetSqlBuilder())
	err = store.InitializeFactory()
	if err != nil {
		glog.Fatalf("Error initializing health database: %+v", err)
	}
	// Add servicers to the service
	healthServer, err := southbound.NewHealthServer(store)
	if err != nil {
		glog.Fatalf("Error creating health servicer: %+v", err)
	}
	protos.RegisterHealthServer(srv.GrpcServer, healthServer)

	healthUpdatesServer, err := protected.NewHealthInternalServer(store)
	if err != nil {
		glog.Fatalf("Error creating health servicer: %+v", err)
	}
	health_protos.RegisterHealthInternalServer(srv.GrpcServer, healthUpdatesServer)

	// create a networkHealthStatusReporter to monitor and periodically log metrics
	// on if all gateways in a network are unhealthy
	healthStatusReporter := &reporter.NetworkHealthStatusReporter{}
	go healthStatusReporter.ReportHealthStatus(NETWORK_HEALTH_STATUS_REPORT_INTERVAL)

	// Run the service
	err = srv.Run()
	if err != nil {
		glog.Fatalf("Error running health service: %+v", err)
	}
}
