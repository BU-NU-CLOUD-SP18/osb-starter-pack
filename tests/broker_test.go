package broker

import (
	"os"
	"path/filepath"
	"testing"

	logic "github.com/dataverse-broker/dataverse-broker/pkg/broker"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/pmorie/osb-broker-lib/pkg/broker"
)

func TestBrokerLogic(t *testing.T) {
	// Create a BusinessLogic struct instance (tests dataverse functions)
	businessLogic, errCreate := logic.NewBusinessLogic(logic.Options{CatalogPath: filepath.Join(os.Getenv("GOPATH"), "/src/github.com/dataverse-broker/dataverse-broker/image/whitelist/"), Async: false})

	if errCreate != nil {
		t.Errorf("Error on BusinessLogic creation: %#+v\n", errCreate)
	}

	// Run GetCatalog
	_, errCatalog := businessLogic.GetCatalog(&broker.RequestContext{})

	if errCatalog != nil {
		t.Errorf("Error on GetCatalog: %#+v\n", errCatalog)
	}

	// Run Provision on a couple of test cases:
	// Blank credentials

	_, errProvisionBlank := businessLogic.Provision(
		&osb.ProvisionRequest{
			InstanceID:        "test1",
			AcceptsIncomplete: false,
			ServiceID:         "c241d773-97a1-4d5a-9d7c-c3bea965d601",
			PlanID:            "060c93ba-3bab-4ae0-94ab-81128e946d6c",
			OrganizationGUID:  "bdc",
			SpaceGUID:         "bdc",
			Parameters:        map[string]interface{}{},
		},
		&broker.RequestContext{})

	if errProvisionBlank != nil {
		t.Errorf("Error on Provision with blank token: %#+v\n", errProvisionBlank)
	}

	// Improper credentials
	_, errProvisionImproper := businessLogic.Provision(
		&osb.ProvisionRequest{
			InstanceID:        "test2",
			AcceptsIncomplete: false,
			ServiceID:         "c241d773-97a1-4d5a-9d7c-c3bea965d601",
			PlanID:            "060c93ba-3bab-4ae0-94ab-81128e946d6c",
			OrganizationGUID:  "bdc",
			SpaceGUID:         "bdc",
			Parameters: map[string]interface{}{
				"credentials": "not-real-token",
			},
		},
		&broker.RequestContext{})

	// This should give an error
	if errProvisionImproper == nil {
		t.Errorf("Error on Provision with invalid token: no error returned\n")
	}

	// Provisioning an instance doesn't exist
	_, errProvisionFake := businessLogic.Provision(
		&osb.ProvisionRequest{
			InstanceID:        "not-an-instance",
			AcceptsIncomplete: false,
			ServiceID:         "probably-unique-service",
			PlanID:            "probably-unique-plan",
			OrganizationGUID:  "bdc",
			SpaceGUID:         "bdc",
			Parameters:        map[string]interface{}{},
		},
		&broker.RequestContext{})

	if errProvisionFake == nil {
		t.Errorf("Error on Provision with fake service: No error returned\n")
	}

	// Instance exists, already provisioned
	provisionResponseExists, errProvisionExists := businessLogic.Provision(
		&osb.ProvisionRequest{
			InstanceID:        "test1",
			AcceptsIncomplete: false,
			ServiceID:         "c241d773-97a1-4d5a-9d7c-c3bea965d601",
			PlanID:            "060c93ba-3bab-4ae0-94ab-81128e946d6c",
			OrganizationGUID:  "bdc",
			SpaceGUID:         "bdc",
			Parameters:        map[string]interface{}{},
		},
		&broker.RequestContext{})

	if errProvisionExists != nil {
		t.Errorf("Error on Provision with instance that already exists: %#+v\n", errProvisionExists)
	}
	if provisionResponseExists.Exists == false {
		t.Errorf("Error on Provision with instance that already exists: Response's 'Exists' field should be true: %#+v\n", provisionResponseExists)
	}

	// Instance ID in use, but requesting different service
	_, errProvisionInUse := businessLogic.Provision(
		&osb.ProvisionRequest{
			InstanceID:        "test1",
			AcceptsIncomplete: false,
			ServiceID:         "c241d773-97a1-4d5a-9d7c-c3bea965d601",
			PlanID:            "different-plan",
			OrganizationGUID:  "bdc",
			SpaceGUID:         "bdc",
			Parameters:        map[string]interface{}{},
		},
		&broker.RequestContext{})

	if errProvisionInUse == nil {
		t.Errorf("Error on Provision with service in use: No error returned\n")
	}

	/*
		// Proper credentials
		_, err = businessLogic.Provision(
			&osb.ProvisionRequest{
				InstanceID:	"harvard-ephelps",
				AcceptsIncomplete:	false,
				ServiceID:	"harvard-ephelps",
				PlanID:	"harvard-ephelps-default",
				OrganizationGUID:	"bdc",
				SpaceGUID:	"bdc",
				Parameters:	map[string]interface{}{
						"credentials":"totally-real-token", // replace this with real token in secure way
					},
			},
			&broker.RequestContext{})

		// this should succeed, using token from config
		if err != nil {
			t.Errorf("Error on Provision with valid token: %#+v", err)
		}
	*/

	// Run Bind on a couple of test cases
	// Blank credentials

	_, errBindBlank := businessLogic.Bind(
		&osb.BindRequest{
			BindingID:         "test-binding1",
			InstanceID:        "test1",
			AcceptsIncomplete: false,
			ServiceID:         "c241d773-97a1-4d5a-9d7c-c3bea965d601",
			PlanID:            "060c93ba-3bab-4ae0-94ab-81128e946d6c",
			Parameters:        map[string]interface{}{},
		},
		&broker.RequestContext{})

	if errBindBlank != nil {
		t.Errorf("Error on Bind with no token: %#+v\n", errBindBlank)
	}

	/*
		// Proper credentials
		bindResultProper, err := businessLogic.Bind(
			&osb.BindRequest{
				BindingID:	"harvard-ephelps",
				InstanceID:	"harvard-ephelps",
				AcceptsIncomplete:	false,
				ServiceID:	"harvard-ephelps",
				PlanID:	"harvard-ephelps-default",
				Parameters:	map[string]interface{}{
					"credentials": "totally-real-token",
				},
			},
			&broker.RequestContext{})

		if err != nil{
			t.Errorf("Error on Bind with valid token: %#+v", err)
		}

		if bindResultProper.BindResponse.Credentials["credentials"] != "totally-real-token" || bindResultProper.BindResponse.Credentials["coordinates"] == nil{
			t.Errorf("Error on Bind: Credentials and coordinates not passed properly")
		}
	*/

	// Instance doesn't exist
	_, errBindFake := businessLogic.Bind(
		&osb.BindRequest{
			BindingID:         "fudged-binding",
			InstanceID:        "not-an-instance",
			AcceptsIncomplete: false,
			ServiceID:         "probably-unique-service",
			PlanID:            "probably-unique-plan",
			Parameters:        map[string]interface{}{},
		},
		&broker.RequestContext{})

	if errBindFake == nil {
		t.Errorf("Error on Bind to fake service: No error returned\n")
	}

	// Run Unbind and Deprovision on a few cases
	// Service that we Provisioned and Binded to
	_, errUnbindReal := businessLogic.Unbind(&osb.UnbindRequest{
		BindingID:         "test-binding1",
		InstanceID:        "test1",
		AcceptsIncomplete: false,
		ServiceID:         "c241d773-97a1-4d5a-9d7c-c3bea965d601",
		PlanID:            "060c93ba-3bab-4ae0-94ab-81128e946d6c",
	},
		&broker.RequestContext{})

	if errUnbindReal != nil {
		t.Errorf("Error on Unbind: %#+v\n", errUnbindReal)
	}

	_, errDeprovisionReal := businessLogic.Deprovision(&osb.DeprovisionRequest{
		InstanceID:        "test1",
		AcceptsIncomplete: false,
		ServiceID:         "c241d773-97a1-4d5a-9d7c-c3bea965d601",
		PlanID:            "060c93ba-3bab-4ae0-94ab-81128e946d6c",
	},
		&broker.RequestContext{})

	if errDeprovisionReal != nil {
		t.Errorf("Error on Deprovision: %#+v\n", errDeprovisionReal)
	}

}
