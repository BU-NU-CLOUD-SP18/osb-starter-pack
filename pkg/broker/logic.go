package broker

import (
	"sync"
	"encoding/json"
	"fmt"

	"github.com/golang/glog"
	"github.com/pmorie/osb-broker-lib/pkg/broker"

	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"reflect"
)

// NewBusinessLogic is a hook that is called with the Options the program is run
// with. NewBusinessLogic is the place where you will initialize your
// BusinessLogic the parameters passed in.
func NewBusinessLogic(o Options) (*BusinessLogic, error) {
	// For example, if your BusinessLogic requires a parameter from the command
	// line, you would unpack it from the Options and set it on the
	// BusinessLogic here.
	return &BusinessLogic{
		async:     o.Async,
		instances: make(map[string]*dataverseService, 10),
		dataverse_server: "harvard",
		dataverse_url: "https://dataverse.harvard.edu",
		// call dataverse server as little as possible
		dataverses: GetDataverseServices("https://dataverse.harvard.edu", "harvard"),
	}, nil
}

var _ broker.Interface = &BusinessLogic{}


func truePtr() *bool {
	b := true
	return &b
}

func (b *BusinessLogic) GetCatalog(c *broker.RequestContext) (*broker.CatalogResponse, error) {
	// Your catalog business logic goes here
	response := &broker.CatalogResponse{}

	// Create Service objects from dataverses
	services, err :=  DataverseToService(b.dataverses, b.dataverse_server)

	if err != nil {
		panic(err)
	}

	osbResponse := &osb.CatalogResponse{
		Services : services,
	}

	glog.Infof("catalog response: %#+v", osbResponse)

	response.CatalogResponse = *osbResponse

	return response, nil
}


func (b *BusinessLogic) Provision(request *osb.ProvisionRequest, c *broker.RequestContext) (*broker.ProvisionResponse, error) {
	// Your provision business logic goes here

	// example implementation:
	b.Lock()
	defer b.Unlock()

	response := broker.ProvisionResponse{}

	dataverseService := &dataverseService{
		ID:        request.InstanceID,
		ServiceID: request.ServiceID,
		PlanID:    request.PlanID,
		Params:    request.Parameters,
	}

	// Check to see if this is the same instance
	if i := b.instances[request.InstanceID]; i != nil {
		if i.Match(dataverseService) {
			response.Exists = true
			return &response, nil
		} else {
			// Instance ID in use, this is a conflict.
			description := "InstanceID in use"
			return nil, osb.HTTPStatusCodeError{
				StatusCode: http.StatusConflict,
				Description: &description,
			}
		}
	}

	// this should probably run asynchronously if possible
	if dataverseService.Params["credentials"] != nil && dataverseService.Params["credentials"].(string) != "" {
		// check that the token is valid, make a call to the Dataverse server
		// make a GET request
		
		resp, err := http.Get(b.dataverse_url + "/api/dataverses/:root?key=" + dataverseService.Params["credentials"].(string))

		if err != nil{
			return nil, osb.HTTPStatusCodeError{
				StatusCode: http.StatusNotFound,
			}
		}

		// Must close response when finished
		defer resp.Body.Close()

		//convert resp into a DataverseResponse object
		body, err := ioutil.ReadAll(resp.Body)

		if err != nil{
			return nil, osb.HTTPStatusCodeError{
				StatusCode: http.StatusNotFound,
			}
		}

		dataverseResp := DataverseResponseWrapper{}
		err = json.Unmarshal(body, &dataverseResp)

		// failed GET means token is invalid (what to do?)
		if err != nil || dataverseResp.Status != "OK"{
			description := "Bad api key '" + dataverseService.Params["credentials"].(string) + "'"
			return nil, osb.HTTPStatusCodeError{
				StatusCode: http.StatusBadRequest,
				Description: &description,
			}
		}

	}
	
  
	b.instances[request.InstanceID] = dataverseService

	if request.AcceptsIncomplete {
		response.Async = b.async
	}

	return &response, nil
}

func (b *BusinessLogic) Deprovision(request *osb.DeprovisionRequest, c *broker.RequestContext) (*broker.DeprovisionResponse, error) {
	// Your deprovision business logic goes here

	// example implementation:
	b.Lock()
	defer b.Unlock()

	response := broker.DeprovisionResponse{}

	delete(b.instances, request.InstanceID)

	if request.AcceptsIncomplete {
		response.Async = b.async
	}

	return &response, nil
}

func (b *BusinessLogic) LastOperation(request *osb.LastOperationRequest, c *broker.RequestContext) (*broker.LastOperationResponse, error) {
	// Your last-operation business logic goes here

	return nil, nil
}

func (b *BusinessLogic) Bind(request *osb.BindRequest, c *broker.RequestContext) (*broker.BindResponse, error) {
	// Your bind business logic goes here

	// example implementation:
	b.Lock()
	defer b.Unlock()

	instance, ok := b.instances[request.InstanceID]
	if !ok {
		return nil, osb.HTTPStatusCodeError{
			StatusCode: http.StatusNotFound,
		}
	}

	credentials := ""
	if instance.Params["credentials"] != nil {
			credentials = instance.Params["credentials"].(string)
	}

	response := broker.BindResponse{
		BindResponse: osb.BindResponse{
			// Get the service URL based on the serviceID (which is funny because they're the same thing right now...)
			Credentials: map[string]interface{}{
				"coordinates": b.dataverses[instance.ServiceID].Url,
				"credentials": credentials,
				},
		},

	}
	if request.AcceptsIncomplete {
		response.Async = b.async
	}

	return &response, nil
}

func (b *BusinessLogic) Unbind(request *osb.UnbindRequest, c *broker.RequestContext) (*broker.UnbindResponse, error) {
	// Your unbind business logic goes here
	return &broker.UnbindResponse{}, nil
}

func (b *BusinessLogic) Update(request *osb.UpdateInstanceRequest, c *broker.RequestContext) (*broker.UpdateInstanceResponse, error) {
	// Your logic for updating a service goes here.
	response := broker.UpdateInstanceResponse{}
	if request.AcceptsIncomplete {
		response.Async = b.async
	}

	return &response, nil
}

func (b *BusinessLogic) ValidateBrokerAPIVersion(version string) error {
	return nil
}

func (i *dataverseService) Match(other *dataverseService) bool {
	return reflect.DeepEqual(i, other)
}
