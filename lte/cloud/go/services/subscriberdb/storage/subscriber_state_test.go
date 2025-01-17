/*
 Copyright 2022 The Magma Authors.

 This source code is licensed under the BSD-style license found in the
 LICENSE file in the root directory of this source tree.

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package storage

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"

	"magma/orc8r/cloud/go/sqorc"
)

const nid1 = "network_1"
const nid2 = "network_2"

const gwid1 = "snowflake_ID_1"
const gwid2 = "snowflake_ID_2"

func TestSubscriberStorage(t *testing.T) {
	db, err := sqorc.Open("sqlite3", ":memory:")
	assert.NoError(t, err)
	s := NewSubscriberStorage(db, sqorc.GetSqlBuilder())
	assert.NoError(t, s.Initialize())

	t.Run("Test inserting, querying and deleting subscriber states for a gateway", func(t *testing.T) {
		subscriberStates := []SubscriberState{subscriberState("IMSI001010000000123"), subscriberState("IMSI001010000000456")}
		err = s.SetAllSubscribersForGateway(nid1, gwid1, subscriberStates)
		assert.NoError(t, err)

		foundStates, err := s.GetSubscribersForGateway(nid2, gwid1)
		assert.NoError(t, err)
		assert.Empty(t, foundStates)
		foundStates, err = s.GetSubscribersForGateway(nid1, gwid2)
		assert.NoError(t, err)
		assert.Empty(t, foundStates)

		foundStates, err = s.GetSubscribersForGateway(nid1, gwid1)
		assert.NoError(t, err)
		assert.True(t, cmp.Equal(foundStates, subscriberStates))

		err = s.DeleteSubscribersForGateway(nid1, gwid1)
		assert.NoError(t, err)
		foundStates, err = s.GetSubscribersForGateway(nid1, gwid1)
		assert.NoError(t, err)
		assert.Empty(t, foundStates)
	})

	t.Run("Test updating subscriber states for a gateway", func(t *testing.T) {
		oldSubscriberStatesGw1 := []SubscriberState{subscriberState("IMSI001010000000123"), subscriberState("IMSI001010000000456")}
		subscriberStatesGw2 := []SubscriberState{subscriberState("IMSI001010000000789")}

		err = s.SetAllSubscribersForGateway(nid1, gwid1, oldSubscriberStatesGw1)
		assert.NoError(t, err)
		err = s.SetAllSubscribersForGateway(nid1, gwid2, subscriberStatesGw2)
		assert.NoError(t, err)

		foundStates, err := s.GetSubscribersForGateway(nid1, gwid1)
		assert.NoError(t, err)
		assert.True(t, cmp.Equal(foundStates, oldSubscriberStatesGw1))

		// Test updating states for gateway 1

		newSubscriberStatesGw1 := []SubscriberState{subscriberState("IMSI001010000000012"), subscriberState("IMSI001010000000123")}
		err = s.SetAllSubscribersForGateway(nid1, gwid1, newSubscriberStatesGw1)
		assert.NoError(t, err)

		foundStates, err = s.GetSubscribersForGateway(nid1, gwid1)
		assert.NoError(t, err)
		assert.True(t, cmp.Equal(foundStates, newSubscriberStatesGw1))

		// Test emptying states for gateway 1

		err = s.SetAllSubscribersForGateway(nid1, gwid1, []SubscriberState{})
		assert.NoError(t, err)

		foundStates, err = s.GetSubscribersForGateway(nid1, gwid1)
		assert.NoError(t, err)
		assert.Empty(t, foundStates)

		// Gateway 2 should be unaffected

		foundStates, err = s.GetSubscribersForGateway(nid1, gwid2)
		assert.NoError(t, err)
		assert.True(t, cmp.Equal(foundStates, subscriberStatesGw2))

		err = s.DeleteSubscribersForGateway(nid1, gwid1)
		assert.NoError(t, err)
		err = s.DeleteSubscribersForGateway(nid1, gwid2)
		assert.NoError(t, err)
	})
}

func subscriberState(imsi string) SubscriberState {
	return SubscriberState{
		Imsi: imsi,
		Value: map[string]interface{}{
			"magma.ipv4": []interface{}{
				map[string]interface{}{
					"active_policy_rules": []interface{}{},
					"active_duration_sec": 6.,
					"lifecycle_state":     "SESSION_RELEASED",
					"session_start_time":  1653484144.,
					"apn":                 "magma.ipv4",
					"ipv4":                "192.168.128.12",
					"msisdn":              "",
					"session_id":          fmt.Sprintf("%s-1234", imsi),
				},
			},
		},
	}
}
