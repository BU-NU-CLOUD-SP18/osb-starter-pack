package broker

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	logic "github.com/dataverse-broker/dataverse-broker/pkg/broker"
)

/* Check all whitelisted services for validity:
- Completeness: Has all required info for a dataverse service
	(server name, identifier, Name, etc)
- Uniqueness: Unique service ids and plan ids, etc
- Existence: Pings the server to see if the dataverse exists/is live
*/
func TestWhitelist(t *testing.T) {

	whitelistPath := filepath.Join(os.Getenv("GOPATH"), "/src/github.com/dataverse-broker/dataverse-broker/image/whitelist/")
	services, err := logic.FileToService(whitelistPath)

	if err != nil {
		t.Errorf("Error creating files: %#+v\n", err)
	}

	uuids := make(map[string]int)

	for _, dataverse := range services {

		// Completeness
		complete := func() bool {
			// check if a DataverseInstance object has all required fields
			// MUST have a non-empty ServiceID and PlanID
			if dataverse.ServiceID == "" || dataverse.PlanID == "" ||
				// Must have a non-empty ServerName and ServerUrl
				dataverse.ServerName == "" || dataverse.ServerUrl == "" ||
				// Must have a non-empty Name, Identifier, and Url
				dataverse.Description.Name == "" || dataverse.Description.Identifier == "" || dataverse.Description.Url == "" {

				return false
			}

			// Name MUST be non-empty after removing illegal characters (non-alphanumeric with exception for dash and period)
			reg, err := regexp.Compile("[^a-zA-Z0-9-.]+")
			if err != nil {
				return false
			}

			if reg.ReplaceAllString(strings.Replace(dataverse.Description.Name, " ", "-", -1), "") == "" {
				return false
			}

			return true
		}

		if complete() == false {
			t.Errorf("Error in whitelist: Dataverse Service not compliant:  %#+v\n", dataverse.ID)
		}

		// Uniqueness
		if _, service_present := uuids[dataverse.ServiceID]; service_present {
			t.Errorf("Error in whitelist: Dataverse Service %s does not have unique ServiceID\n", dataverse.ID)
		} else {
			uuids[dataverse.ServiceID] = 1
		}

		if _, plan_present := uuids[dataverse.PlanID]; plan_present {
			t.Errorf("Error in whitelist: Dataverse Service %s does not have unique PlanID\n", dataverse.ID)
		} else {
			uuids[dataverse.PlanID] = 1
		}

		// Existence
		succ, err := logic.PingDataverse(dataverse.Description.Url)
		if succ == false || err != nil {
			t.Errorf("Error accessing dataverse %s: %#+v\n", dataverse.Description.Url, err)
		}
	}

}
