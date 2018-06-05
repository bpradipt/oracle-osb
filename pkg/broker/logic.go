package broker

import (
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/golang/glog"
	"github.com/pmorie/osb-broker-lib/pkg/broker"

	"reflect"

	osb "github.com/pmorie/go-open-service-broker-client/v2"
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
		instances: make(map[string]*exampleInstance, 10),
		//TBD - secure way of getting the DB access creds
		//sysconn: "user/password@db_host:db_port/service",
		//host: "XXX.YYY.ZZZ.DDD",
		//port: "1521",
		sysconn: o.dbConnStr,
		host:    o.dbHost,
		port:    o.dbPort,
	}, nil
}

// BusinessLogic provides an implementation of the broker.BusinessLogic
// interface.
type BusinessLogic struct {
	// Indicates if the broker should handle the requests asynchronously.
	async bool
	// Synchronize go routines.
	sync.RWMutex
	// Add fields here! These fields are provided purely as an example
	instances map[string]*exampleInstance
	sysconn   string
	host      string
	port      int
}

var _ broker.Interface = &BusinessLogic{}

func truePtr() *bool {
	b := true
	return &b
}

func (b *BusinessLogic) GetCatalog(c *broker.RequestContext) (*broker.CatalogResponse, error) {
	// Your catalog business logic goes here
	response := &broker.CatalogResponse{}
	osbResponse := &osb.CatalogResponse{
		Services: []osb.Service{
			{
				Name:          "create-oracle-table",
				ID:            "4f6e6cf6-ffdd-425f-a2c7-3c9258ad246a",
				Description:   "Create an Oracle DB User and Table with specified schema",
				Bindable:      true,
				PlanUpdatable: truePtr(),
				Metadata: map[string]interface{}{
					"displayName": "Create Oracle Table",
					"imageUrl":    "https://avatars2.githubusercontent.com/u/19862012?s=200&v=4",
				},
				Plans: []osb.Plan{
					{
						Name:        "default",
						ID:          "86064792-7ea2-467b-af93-ac9694d96d5b",
						Description: "The default plan",
						Free:        truePtr(),
						Schemas: &osb.Schemas{
							ServiceInstance: &osb.ServiceInstanceSchema{
								Create: &osb.InputParametersSchema{
									Parameters: map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"dbusername": map[string]interface{}{
												"type":        "string",
												"description": "DB User to create",
											},
											"dbpassword": map[string]interface{}{
												"type":        "string",
												"description": "DB Password",
											},
											"tablename": map[string]interface{}{
												"type":        "string",
												"description": "Table Name",
											},
											"tableschema": map[string]interface{}{
												"type":        "string",
												"description": "Table Schema as SQL Statement",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	glog.Infof("catalog response: %#+v", osbResponse)

	response.CatalogResponse = *osbResponse

	return response, nil
}

//Provision the service broker instance
func (b *BusinessLogic) Provision(request *osb.ProvisionRequest, c *broker.RequestContext) (*broker.ProvisionResponse, error) {
	// Your provision business logic goes here
	//Params:
	//The following details are provided by the user when provisioning the instance
	// dbusername
	// dbpassword
	// dbcontainer
	// The following details should not come from user as these are part of managed service
	// sysconn

	// example implementation:
	b.Lock()
	defer b.Unlock()

	response := broker.ProvisionResponse{}

	exampleInstance := &exampleInstance{
		ID:        request.InstanceID,
		ServiceID: request.ServiceID,
		PlanID:    request.PlanID,
		Params:    request.Parameters,
	}

	// Check to see if this is the same instance
	if i := b.instances[request.InstanceID]; i != nil {
		if i.Match(exampleInstance) {
			response.Exists = true
			return &response, nil
		} else {
			// Instance ID in use, this is a conflict.
			description := "InstanceID in use"
			return nil, osb.HTTPStatusCodeError{
				StatusCode:  http.StatusConflict,
				Description: &description,
			}
		}
	}
	b.instances[request.InstanceID] = exampleInstance

	//glog.V(4).Infof("Request Params: %s", request.Parameters)
	//glog.V(4).Infof("Context Params: %s %s", request.Context, request.Context["namespace"])
	//glog.V(4).Infof("Request : %#+v", request)

	dbusername, _ := exampleInstance.Params["dbusername"].(string)
	dbpassword, _ := exampleInstance.Params["dbpassword"].(string)
	tablename, _ := exampleInstance.Params["tablename"].(string)
	tableschema, _ := exampleInstance.Params["tableschema"].(string)
	dbservice := strings.Split(b.sysconn, "/")[2]

	if err := createUser(dbusername, dbpassword, b.sysconn); err != nil {
		return &response, err
	}

	connURI := dbusername + "/" + dbpassword + "@" + b.host + ":" + strconv.Itoa(b.port) + "/" + dbservice

	if err := createTable(connURI, tablename, tableschema); err != nil {
		return &response, err
	}

	if request.AcceptsIncomplete {
		response.Async = b.async
	}

	return &response, nil
}

//Deprovision the instance. Do any cleanups
func (b *BusinessLogic) Deprovision(request *osb.DeprovisionRequest, c *broker.RequestContext) (*broker.DeprovisionResponse, error) {

	b.Lock()
	defer b.Unlock()

	response := broker.DeprovisionResponse{}

	glog.V(4).Infof("Deprovision : %s", request.InstanceID)

	if val, ok := b.instances[request.InstanceID]; ok {
		//TBD - handle tear down logic for the data (user/table etc) created in the DB.
		dbusername, _ := val.Params["dbusername"].(string)
		glog.V(4).Infof("Delete user : %s", dbusername)
		if err := deleteUser(dbusername, b.sysconn); err != nil {
			glog.Errorf("Error in deleting the user: %v", err)
		}
	}

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
	b.Lock()
	defer b.Unlock()

	instance, ok := b.instances[request.InstanceID]
	if !ok {
		return nil, osb.HTTPStatusCodeError{
			StatusCode: http.StatusNotFound,
		}
	}

	dbservice := strings.Split(b.sysconn, "/")[2]
	uri := instance.Params["dbusername"].(string) + "/" + instance.Params["dbpassword"].(string) +
		"@" + b.host + ":" + strconv.Itoa(b.port) + "/" + dbservice

	instance.Params["uri"] = uri

	response := broker.BindResponse{
		BindResponse: osb.BindResponse{
			Credentials: instance.Params,
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

// example types

// exampleInstance is intended as an example of a type that holds information about a service instance
type exampleInstance struct {
	ID        string
	ServiceID string
	PlanID    string
	Params    map[string]interface{}
}

func (i *exampleInstance) Match(other *exampleInstance) bool {
	return reflect.DeepEqual(i, other)
}
