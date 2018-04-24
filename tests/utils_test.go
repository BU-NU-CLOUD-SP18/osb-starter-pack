package broker

import (
	"testing"

	logic "github.com/dataverse-broker/dataverse-broker/pkg/broker"
)

func TestServiceToFile(t *testing.T) {

	server_alias := "demo"
	target_dataverse := "https://demo.dataverse.org"

	whitelistPath := "./test/"

	// Gets some dataverse info from the demo dataverse
	dataverses := logic.GetDataverseInstances(target_dataverse, server_alias)

	for _, dataverse := range dataverses {
		// Write the dataverses collected into json files
		succ, err := logic.ServiceToFile(dataverse, whitelistPath)

		if err != nil || succ != true {
			t.Errorf("Error writing json to files: %#+v\n", err)
		}

	}

	// Read in the json files written above for validity
	_, err := logic.FileToService(whitelistPath)

	if err != nil {
		t.Errorf("Error creating files: %#+v\n", err)
	}

}
