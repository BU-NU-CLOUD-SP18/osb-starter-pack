package broker

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"fmt"
	"regexp"
	"strconv"
	"strings"

	"os"
	"path/filepath"

	"reflect"

	"github.com/golang/glog"

	osb "github.com/pmorie/go-open-service-broker-client/v2"
)

// Displays the individual metadata of the metadatablocks from a dataverse through injection of their ids
// if they are of type dataverse
func DataverseMeta(base string, id float64) {

	// Injecting string version of id into the search uri
	search_uri := "/api/dataverses/" + fmt.Sprint(id)

	// Variable that will store the json from the GET request
	var status map[string]interface{}

	// Executing GET request
	resp, err := http.Get(base + search_uri)

	if err != nil {
		// Exit on error
		fmt.Println("Error on http GET at address", base+search_uri)
		fmt.Println(err)
		panic("")
	}

	// Must close response when finished
	defer resp.Body.Close()

	// Convert resp into a DataverseResponse object
	body, err := ioutil.ReadAll(resp.Body)

	err = json.Unmarshal(body, &status)
	if err != nil {
		fmt.Println(err)
	}

	// Checking GET response json
	if status["status"] == "ERROR" {

		// Skipping..
		fmt.Println("passing, not a dataverse.")

	} else {

		// Printing metadata
		fmt.Println(string(body))

	}

}

// Gathers the ids of the metadatablocks from a dataverse, calls on other function which displays the metadata
// of the individual metadatablocks (DataverseMeta), and returns an array of those ids if needed
func DataverseMetadataIds(base string) []float64 {

	// This search uri finds the metadata of the dataverse and displays info objects with according id's
	search_uri := "/api/metadatablocks"

	// Variable that will store the json from the GET request
	var metadata map[string]interface{}

	// Make a GET request
	resp, err := http.Get(base + search_uri)

	if err != nil {
		// Exit on error
		fmt.Println("Error on http GET at address", base+search_uri)
		fmt.Println(err)
		panic("")
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	err = json.Unmarshal(body, &metadata)
	if err != nil {
		fmt.Println(err)
	}

	// Creating return array object
	ids := make([]float64, 0)

	i := 0

	// Iterating through metadatablock
	for {

		// This recover function is to catch the panic when the index is out of range
		// indictating that the function has reached the end of the list of metadatablocks
		// and thus returns the array
		defer func() []float64 {
			if r := recover(); r != nil {
				return ids
			}
			return make([]float64, 0)
		}()

		// Retrieving ids from metadatablock
		test := metadata["data"].([]interface{})[i].(map[string]interface{})["id"]

		// Asserting test's type before injecting to separate function and return array
		string_test := test.(float64)

		// Calling function which displays the individual metadata of the metadatablocks
		DataverseMeta(base, string_test)

		// Appending to the return array
		ids = append(ids, string_test)

		// Incrementing iterator
		i++

	}

}

func DataverseToService(dataverses map[string]*dataverseInstance) ([]osb.Service, error) {
	// Use DataverseDescription to populate osb.Service objects
	services := make([]osb.Service, len(dataverses))

	i := 0
	reg, err := regexp.Compile("[^a-zA-Z0-9-.]+")
	if err != nil {
		return nil, err
	}

	for _, dataverse := range dataverses {
		// Check that each field has a value
		// This name MUST be alphanumeric, dashes, and periods ONLY (no spaces)
		service_dashname := strings.ToLower(reg.ReplaceAllString(strings.Replace(dataverse.Description.Name, " ", "-", -1), ""))

		service_id := dataverse.ServiceID
		plan_id := dataverse.PlanID
		service_description := dataverse.Description.Description
		service_name := dataverse.Description.Name
		service_image_url := dataverse.Description.Image_url

		if service_description == "" {
			service_description = "A Dataverse service"
		}

		if service_image_url == "" {
			// Default image for osb service
			service_image_url = "https://avatars2.githubusercontent.com/u/19862012?s=200&v=4"
		}

		services[i] = osb.Service{
			Name:          service_dashname,
			ID:            service_id,
			Description:   service_description,
			Bindable:      true,
			PlanUpdatable: truePtr(),
			Metadata: map[string]interface{}{
				"displayName": service_name,
				"imageUrl":    service_image_url,
			},
			Plans: []osb.Plan{
				{
					Name:        "default",
					ID:          plan_id,
					Description: "The default plan for " + service_name,
					Free:        truePtr(),
					Schemas: &osb.Schemas{
						ServiceInstance: &osb.ServiceInstanceSchema{
							Create: &osb.InputParametersSchema{
								Parameters: map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"credentials": map[string]interface{}{
											"type":        "string",
											"description": "API key to access restricted files and datasets on Dataverse",
											"default":     "",
										},
									},
								},
							},
						},
					},
				},
			},
		}

		i += 1
	}

	return services, nil
}

func GetDataverseInstances(target_dataverse string, server_alias string) map[string]*dataverseInstance {

	dataverses, err := SearchForDataverses(&target_dataverse, 10)

	if err != nil {
		panic(err)
	}

	services := make(map[string]*dataverseInstance, len(dataverses))

	for _, dataverse := range dataverses {
		services[server_alias+"-"+dataverse.Identifier] = &dataverseInstance{
			ID:          server_alias + "-" + dataverse.Identifier,
			ServiceID:   server_alias + "-" + dataverse.Identifier,
			PlanID:      server_alias + "-" + dataverse.Identifier + "-default",
			ServerName:  server_alias,
			ServerUrl:   target_dataverse,
			Description: dataverse,
		}
	}

	return services
}

func FileToService(path string) ([]*dataverseInstance, error) {
	// Read from dataverses.json and make each of them available on OpenShift

	var jsonPath string
	jsonPath = path + "/dataverses.json"
	jsonFile, err := os.Open(jsonPath)
	if err != nil {
		glog.Error(err)
		return nil, err
	}
	byteValue, _ := ioutil.ReadAll(jsonFile)

	var instances []*dataverseInstance
	json.Unmarshal(byteValue, &instances)
	defer jsonFile.Close()
	return instances, nil

}

func ServiceToFile(instance *dataverseInstance, path string) (bool, error) {
	// Take a service and store as JSON object in file
	// Save as a file in path

	err := os.MkdirAll(path, os.ModePerm)

	if err != nil {
		return false, err
	}

	// Get JSON from instance
	jsonInstance, err := json.Marshal(instance)

	if err != nil {
		return false, err
	}

	// Write to file
	err = ioutil.WriteFile(filepath.Join(path, instance.ServiceID+".json"), jsonInstance, 0777)

	if err != nil {
		return false, err
	}

	return true, nil
}

// Get all dataverses within a Dataverse server
// Takes a base Dataverse URL
// Returns a slice of string JSON objects, representing each dataverse
func SearchForDataverses(base *string, max_results_opt ...int) ([]*DataverseDescription, error) {
	// Send a GET request to Dataverse url
	max_results := 0
	if len(max_results_opt) > 0 {
		max_results = max_results_opt[0]
	}

	// Search API for dataverses
	search_uri := "/api/search"

	options := "?q=*&type=dataverse&start="

	// Start with first search results, and only read back per_page number of dataverses per GET
	start := 0
	per_page := 10

	total_count := 0

	query_completed := false

	// Slice to hold list of
	dataverses := make([]*DataverseDescription, 0)

	for query_completed == false {

		// Make a GET request
		if max_results > 0 && max_results < start+per_page {
			// Don't go over max_results
			per_page = max_results - start
		}
		resp, err := http.Get(*base + search_uri + options + strconv.Itoa(start) + "&per_page=" + strconv.Itoa(per_page))

		if err != nil {
			return nil, err
		}

		// Must close response when finished
		defer resp.Body.Close()

		// Convert resp into a DataverseResponse object
		body, err := ioutil.ReadAll(resp.Body)

		response := DataverseResponseWrapper{}
		err = json.Unmarshal(body, &response)

		// Check that Get was successful
		if err != nil {
			return nil, err
		}

		if response.Status != "OK" {
			err = fmt.Errorf("Error in DataverseResponse status: %s\n", response.Status)
			return nil, err
		}

		// Obtain "total_count" for number of dataverses available at the server
		total_count = response.Data.Total_count

		// In case there are no results...
		if total_count == 0 {
			err = fmt.Errorf("No results from GET query")
			return nil, err
		}
		// Otherwise, set condition to false if we've reached total_count
		if total_count == start+response.Data.Count_in_response {
			query_completed = true
		}
		// Reached max results
		if max_results > 0 && max_results <= start+response.Data.Count_in_response {
			query_completed = true
		}

		// Iterate across each DataverseDescription
		for i := 0; i < response.Data.Count_in_response; i++ {
			// Cast elements of list to DataverseDescription objects
			desc := DataverseDescription{}

			desc = response.Data.Items[i]

			// Append DataverseDescription to dataverses slice
			dataverses = append(dataverses, &desc)
		}

		// Update start value
		start += response.Data.Count_in_response
	}

	return dataverses, nil

}

func PingDataverseToken(serverUrl string, token string) (bool, error) {
	// Ping the url, return bool for success or failure, and error code on fail
	resp, err := http.Get(serverUrl + "/api/dataverses/:root?key=" + token)

	if err != nil {
		return false, osb.HTTPStatusCodeError{
			StatusCode: http.StatusNotFound,
		}
	}

	// Must close response when finished
	defer resp.Body.Close()

	// Convert resp into a DataverseResponse object
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return false, osb.HTTPStatusCodeError{
			StatusCode: http.StatusNotFound,
		}
	}

	dataverseResp := DataverseResponseWrapper{}
	err = json.Unmarshal(body, &dataverseResp)

	if err != nil || dataverseResp.Status != "OK" {
		return false, osb.HTTPStatusCodeError{
			StatusCode:  http.StatusBadRequest,
			Description: &dataverseResp.Message,
		}
	}

	// Reaching here means successful ping
	return true, nil
}

func PingDataverse(url string) (bool, error) {
	resp, err := http.Get(url)

	if err != nil || resp.StatusCode == 404 {
		return false, osb.HTTPStatusCodeError{
			StatusCode: http.StatusNotFound,
		}
	}

	// Must close response when finished
	defer resp.Body.Close()

	// Reaching here means successful ping
	return true, nil
}

func truePtr() *bool {
	b := true
	return &b
}

func (i *dataverseInstance) Match(other *dataverseInstance) bool {
	return reflect.DeepEqual(i, other)
}
