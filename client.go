package golinode

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-resty/resty"
)

const (
	// APIHost Linode API hostname
	APIHost = "api.linode.com"
	// APIVersion Linode API version
	APIVersion = "v4"
	// APIProto connect to API with http(s)
	APIProto = "https"
	// Version of golinode
	Version = "1.0.0"
	// APIEnvVar environment var to check for API key
	APIEnvVar = "LINODE_API_KEY"
)

// Client is a wrapper around the Resty client
type Client struct {
	apiKey    string
	resty     *resty.Client
	resources map[string]*Resource

	Images       *Resource
	Disks        *Resource
	Configs      *Resource
	Backups      *Resource
	Instances    *Resource
	Regions      *Resource
	StackScripts *Resource
	Volumes      *Resource
	Kernels      *Resource
	Types        *Resource
}

// R wraps resty's R method
func (c *Client) R() *resty.Request {
	return c.resty.R()
}

// SetDebug sets the debug on resty's client
func (c *Client) SetDebug(debug bool) *Client {
	c.resty.SetDebug(debug)
	return c
}

// Resource looks up a resource by name
func (c Client) Resource(resourceName string) *Resource {
	selectedResource, ok := c.resources[resourceName]
	if !ok {
		log.Fatalf("Could not find resource named '%s', exiting.", resourceName)
	}
	return selectedResource
}

// ListOptions are the pagination parameters for List endpoints
type ListOptions struct {
	Page    int `url:"page,omitempty"`
	PerPage int `url:"per_page,omitempty"`
	Results int `url:"results,omitempty"`
}

// NewClient factory to create new Client struct
func NewClient(codeAPIKey *string, transport http.RoundTripper) (*Client, error) {
	linodeAPIKey := ""

	if codeAPIKey != nil {
		linodeAPIKey = *codeAPIKey
	} else if envAPIKey, ok := os.LookupEnv(APIEnvVar); ok {
		linodeAPIKey = envAPIKey
	}
	if len(linodeAPIKey) == 0 || linodeAPIKey == "" {
		return nil, errors.New("No API key was provided or LINODE_API_KEY was not set")
	}

	restyClient := resty.New().
		SetHostURL(fmt.Sprintf("%s://%s/%s", APIProto, APIHost, APIVersion)).
		SetAuthToken(linodeAPIKey).
		SetTransport(transport).
		SetHeader("User-Agent", fmt.Sprintf("go-linode %s https://github.com/chiefy/go-linode", Version))

	resources := map[string]*Resource{
		stackscriptsName: NewResource(stackscriptsName, stackscriptsEndpoint, false),
		imagesName:       NewResource(imagesName, imagesEndpoint, false),
		instancesName:    NewResource(instancesName, instancesEndpoint, false),
		regionsName:      NewResource(regionsName, regionsEndpoint, false),
		disksName:        NewResource(disksName, disksEndpoint, true),
		configsName:      NewResource(configsName, configsEndpoint, true),
		backupsName:      NewResource(backupsName, backupsEndpoint, true),
		volumesName:      NewResource(volumesName, volumesEndpoint, false),
		kernelsName:      NewResource(kernelsName, kernelsEndpoint, false),
		typesName:        NewResource(typesName, typesEndpoint, false),
	}

	return &Client{
		apiKey:    linodeAPIKey,
		resty:     restyClient,
		resources: resources,

		Images:       resources[imagesName],
		StackScripts: resources[stackscriptsName],
		Instances:    resources[instancesName],
		Regions:      resources[regionsName],
		Disks:        resources[disksName],
		Configs:      resources[configsName],
		Backups:      resources[backupsName],
		Volumes:      resources[volumesName],
		Kernels:      resources[kernelsName],
		Types:        resources[typesName],
	}, nil
}
