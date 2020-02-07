package config

import (
	"fmt"
	"strings"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
)

// ConsulService ...
type ConsulService api.CatalogRegistration

// ToConsulService ...
func (c *ConsulService) ToConsulService() *api.CatalogRegistration {
	return &api.CatalogRegistration{
		ID:      c.ID,
		Node:    c.Node,
		Address: c.Address,
		Service: c.Service,
		Check:   c.Check,
	}
}

// ConsulServices struct
//
type ConsulServices []*ConsulService

// add ...
func (cs *ConsulServices) add(service *ConsulService) {
	*cs = append(*cs, service)
}

func (cs *ConsulServices) List() ConsulServices {
	return *cs
}

func (c *Config) parseConsulServiceStanza(list *ast.ObjectList, environment *Environment) error {
	if len(list.Items) == 0 {
		return nil
	}

	c.logger = c.logger.WithField("stanza", "service")
	c.logger.Debugf("Found %d service{}", len(list.Items))
	for _, serviceAST := range list.Items {
		x := serviceAST.Val.(*ast.ObjectType).List

		valid := []string{"id", "address", "node", "port", "tags", "meta"}
		if err := c.checkHCLKeys(x, valid); err != nil {
			return err
		}

		if len(serviceAST.Keys) != 1 {
			return fmt.Errorf("Missing service name in line %+v", serviceAST.Keys[0].Pos())
		}

		serviceName := serviceAST.Keys[0].Token.Value().(string)

		address, err := getKeyString("address", x)
		if err != nil {
			return err
		}

		node, err := getKeyString("node", x)
		if err != nil {
			return err
		}

		port, err := getKeyInt("port", x)
		if err != nil {
			return err
		}

		tags, err := getKeyStringList("tags", x)
		if err != nil {
			if strings.Contains(err.Error(), "missing tags") {
				tags = make([]string, 0)
			} else {
				return err
			}
		}

		serviceID, err := getKeyString("id", x)
		if err != nil {
			if strings.Contains(err.Error(), "missing id") {
				serviceID = serviceName
			} else {
				return err
			}
		}

		var m map[string]string
		if serviceMetaObj := x.Filter("meta").Items; len(serviceMetaObj) > 0 {
			if err := hcl.DecodeObject(&m, serviceMetaObj[0].Val); err != nil {
				return err
			}
		}

		service := &ConsulService{
			Node:    node,
			Address: address,
			Service: &api.AgentService{
				Address: address,
				ID:      serviceID,
				Port:    port,
				Service: serviceName,
				Tags:    tags,
				Meta:    m,
			},
			Check: &api.AgentCheck{
				CheckID:     fmt.Sprintf("service:%s", serviceID),
				Name:        serviceName,
				Node:        node,
				Notes:       "created by hashi-helper",
				ServiceName: serviceName,
				ServiceID:   serviceID,
				Status:      "passing",
			},
		}

		c.ConsulServices.add(service)
	}

	return nil
}

func getKeyString(key string, x *ast.ObjectList) (string, error) {
	list := x.Filter(key)
	if len(list.Items) == 0 {
		return "", fmt.Errorf("missing %s", key)
	}

	if len(list.Items) > 1 {
		return "", fmt.Errorf("More than one match for %s", key)
	}

	value := list.Items[0].Val.(*ast.LiteralType).Token.Value().(string)

	return value, nil
}

func getKeyStringList(key string, x *ast.ObjectList) ([]string, error) {
	list := x.Filter(key)
	if len(list.Items) != 1 {
		return nil, fmt.Errorf("missing %s", key)
	}

	z := list.Items[0].Val.(*ast.ListType)

	res := make([]string, 0)
	for _, i := range z.List {
		val := i.(*ast.LiteralType).Token.Value().(string)
		res = append(res, val)
	}

	return res, nil
}

func getKeyInt(key string, x *ast.ObjectList) (int, error) {
	list := x.Filter(key)
	if len(list.Items) == 0 {
		return 0, fmt.Errorf("missing %s", key)
	}

	if len(list.Items) > 1 {
		return 0, fmt.Errorf("More than one match for %s", key)
	}

	value := int(list.Items[0].Val.(*ast.LiteralType).Token.Value().(int64))

	return value, nil
}
