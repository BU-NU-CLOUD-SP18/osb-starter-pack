// Test Provision and Bind
package tests

import(
	"testing"

	"github.com/pmorie/osb-broker-lib/pkg/broker"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	logic "github.com/SamiSousa/dataverse-broker/pkg/broker"
)

func TestBrokerLogic(t *testing.T){
	// create a BusinessLogic struct instance (tests dataverse functions)
	businessLogic, err := logic.NewBusinessLogic(logic.Options{CatalogPath: "", Async: false})

	if err != nil{
		t.Errorf("Error on BusinessLogic creation: %#+v", err)
	}

	// Run Provision on a couple of test cases:
	// credentials blank
	provisionResultBlank, err := businessLogic.Provision(
		&osb.ProvisionRequest{
			InstanceID:	"dataverse-1",
			AcceptsIncomplete:	false,
			ServiceID:	"dataverse-1",
			PlanID:	"dataverse-1-default",
			OrganizationGUID:	"bdc",
			SpaceGUID:	"bdc",
			Parameters:	map[string]interface{}{},
			// The following two properties are omitempty
			/*Context:	map[string]interface{}{

				},
			OriginatingIdentity:	&OriginatingIdentity{

				},*/
		}, 
		// empty because we don't use it
		&broker.RequestContext{})

	if err != nil {
		t.Errorf("Error on Provision with blank token: %#+v", err)
	}

	// credentials notblank improper credentials
	provisionResultImproper, err := businessLogic.Provision(&osb.ProvisionRequest{
			InstanceID:	"ephelps",
			AcceptsIncomplete:	false,
			ServiceID:	"ephelps",
			PlanID:	"ephelps-default",
			OrganizationGUID:	"bdc",
			SpaceGUID:	"bdc",
			Parameters:	map[string]interface{}{
					"credentials":"not-real-token",
				},
			// The following two properties are omitempty
			/*Context:	map[string]interface{}{

				},
			OriginatingIdentity:	&OriginatingIdentity{

				},*/
		}, 
		// empty because we don't use it
		&broker.RequestContext{})

	// we want an error here
	if err == nil {
		t.Errorf("Error on Provision with invalid token: no error returned")
	}

	// credentials notblank proper credentials
	provisionResultProper, err := businessLogic.Provision(&osb.ProvisionRequest{
			InstanceID:	"ephelps",
			AcceptsIncomplete:	false,
			ServiceID:	"ephelps",
			PlanID:	"ephelps-default",
			OrganizationGUID:	"bdc",
			SpaceGUID:	"bdc",
			Parameters:	map[string]interface{}{
					"credentials":"totally-real-token", // replace this with real token in secure way
				},
			// The following two properties are omitempty
			/*Context:	map[string]interface{}{

				},
			OriginatingIdentity:	&OriginatingIdentity{

				},*/
		}, 
		// empty because we don't use it
		&broker.RequestContext{})

	// this should succeed, using token from config
	if err != nil {
		t.Errorf("Error on Provision with valid token: %#+v", err)
	}

	// Run Bind on a couple of test cases
	// credentials blank
	bindResultBlank, err := businessLogic.Bind(&osb.BindRequest{
			BindingID:	"ephelps",
			InstanceID: "ephelps",
			AcceptsIncomplete:	false,
			ServiceID:	"ephelps",
			PlanID:	"ephelps-default",
			Parameters:	map[string]interface{}{},
		}, 
		// empty because we don't use it
		&broker.RequestContext{})

	if err != nil{
		t.Errorf("Error on Bind with no token: %#+v", err)
	}

	// credentials nonblank
	bindResultProper, err := businessLogic.Bind(&osb.BindRequest{
			BindingID:	"ephelps",
			InstanceID: "ephelps",
			AcceptsIncomplete:	false,
			ServiceID:	"ephelps",
			PlanID:	"ephelps-default",
			Parameters:	map[string]interface{}{
				"credentials": "totally-real-token",
			},
		}, 
		// empty because we don't use it
		&broker.RequestContext{})

	if err != nil{
		t.Errorf("Error on Bind with valid token: %#+v", err)
	}

	if bindResultProper.BindResponse.Credentials["credentials"] != "totally-real-token" || bindResultProper.BindResponse.Credentials["coordinates"] == nil{
		t.Errorf("Error on Bind: credentials and coordinates not passed properly")
	}

}