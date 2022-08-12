package tinyhomecommunity

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"unicode"

	"cloud.google.com/go/pubsub"
)

// TinyHomeMessageAttributes sets are attributes set on the given message published
// based on the combination of subscription destination
type TinyHomeMessageAttributes struct {
	GroupsCreated    string `json:"groupsCreated"`
	WorkspaceCreated string `json:"workspaceCreated"`
	TenantCreated    string `json:"tenantCreated"`
	FluxCreated      string `json:"fluxCreated"`
	DeliveredFrom    string `json:"deliveredFrom"`
	TenantName       string `json:"tenantName"`
}

type TinyHomeInstructions struct {
	TenantName           string   `json:"tenantName"`
	Environment          string   `json:"environment"`
	BusinessUnit         string   `json:"businessUnit"`
	TenantOwner          string   `json:"tenantOwner"`
	TenantOwnerSecondary string   `json:"tenantOwnerSecondary"`
	TenantCostCenter     string   `json:"tenantCostCenter"`
	Domain               string   `json:"domain"`
	Organization         string   `json:"organization"`
	Breakglass           bool     `json:"breakglass"`
	BreakglassWindow     string   `json:"breakglassWindow"`
	AddlGkeTenantSaRoles []string `json:"addlGkeTenantSaRoles"`
	AddlGroupIamBindings struct {
		RolesRolesTest []string `json:"roles/roles.test"`
	} `json:"addlGroupIamBindings"`
	NsQuota struct {
		Requests struct {
			Cpu    string `json:"cpu"`
			Memory string `json:"memory"`
		} `json:"requests"`
		Limits struct {
			Cpu    string `json:"cpu"`
			Memory string `json:"memory"`
		} `json:"limits"`
	} `json:"nsQuota"`
}

func (message *TinyHomeInstructions) PublishTinyHomeInstructions(messageAttributes *TinyHomeMessageAttributes) (string, error) {
	// Validate the TinyHomeMessageAttributes
	attrMessage, err := messageAttributes.validateAttributes()
	if err != nil {
		return "", fmt.Errorf("PublishTinyHomeInstructions: %v", err)
	}

	// Validate all TinyHomeInstructions
	err = message.validateInstructions()
	if err != nil {
		return "", fmt.Errorf("PublishTinyHomeInstructions: %v", err)
	}

	byteMessage, _ := json.Marshal(&message)
	projectId := "tdigangi-demos"
	topicId := "tiny-home-api-0.0.1"
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectId)
	if err != nil {
		return "", fmt.Errorf("pubsub.NewClient: %v", err)
	}
	defer client.Close()

	t := client.Topic(topicId)
	result := t.Publish(ctx, &pubsub.Message{
		Data: byteMessage,
		Attributes: map[string]string{
			"groupsCreated":    messageAttributes.GroupsCreated,    // true or false
			"workspaceCreated": messageAttributes.WorkspaceCreated, // true or false
			"tenantCreated":    messageAttributes.TenantCreated,    // true or false
			"fluxCreated":      messageAttributes.FluxCreated,      // true or false
			"deliveredFrom":    messageAttributes.DeliveredFrom,    // manual or galaxy
			"tenantName":       message.TenantName,
		},
	})

	// Block until the result is returned and a server-generated
	// ID is returned for the published message.
	id, err := result.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("PublishTinyHomeInstructions: %v", err)
	}

	log.Printf("tenant name: %v, published message id: %v with attributes: %v \n\n", message.TenantName, id, messageAttributes)
	log.Printf(attrMessage)
	return id, nil
}

// validateInstructions is meant to only cover cases not directly embedded in the pubsub avro messages including
// validate of TenantName, AddlGkeTenantSaRoles
func (message TinyHomeInstructions) validateInstructions() error {
	if len(message.TenantName) > 20 {
		return fmt.Errorf("tenantName greater than 20 characters")
	}

	supportedSpecialChars := []string{"-"}
	// tenantName string can only contain lower case letters & supportedSpecialChars
	for _, r := range message.TenantName {
		if !unicode.IsLower(r) && unicode.IsLetter(r) {
			return fmt.Errorf("tenantName supports only lower case characters")
		}

		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			if !contains(supportedSpecialChars, string(r)) {
				return fmt.Errorf("tenantName is using unsuported special characters, only supported characters are: %s", supportedSpecialChars)
			}
		}
	}

	return nil
}

func (messageAttributes *TinyHomeMessageAttributes) validateAttributes() (string, error) {
	boolVals := []string{"true", "false"}
	deliveryVals := []string{"galaxy", "manual"}
	// Check to make sure all the values supplied are correct
	if !contains(boolVals, messageAttributes.GroupsCreated) {
		return "", fmt.Errorf("message attribute GroupsCreated does not equal true or false")
	}

	if !contains(boolVals, messageAttributes.WorkspaceCreated) {
		return "", fmt.Errorf("message attribute WorkspaceCreated does not equal true or false")
	}

	if !contains(boolVals, messageAttributes.TenantCreated) {
		return "", fmt.Errorf("message attribute TenantCreated does not equal true or false")
	}

	if !contains(boolVals, messageAttributes.FluxCreated) {
		return "", fmt.Errorf("message attribute FluxCreated does not equal true or false")
	}

	if !contains(deliveryVals, messageAttributes.DeliveredFrom) {
		return "", fmt.Errorf("message attribute DeliveredFrom does not equal galaxy or manual")
	}

	var subscriptionText string
	// Messages filtered to the createGroups Subscription
	if messageAttributes.GroupsCreated == "false" && messageAttributes.WorkspaceCreated == "false" && messageAttributes.TenantCreated == "false" && messageAttributes.FluxCreated == "false" {
		subscriptionText = "createGroups"
	} else if messageAttributes.GroupsCreated == "true" && messageAttributes.WorkspaceCreated == "false" && messageAttributes.TenantCreated == "false" && messageAttributes.FluxCreated == "false" {
		subscriptionText = "createWorkspace"
	} else if messageAttributes.GroupsCreated == "true" && messageAttributes.WorkspaceCreated == "true" && messageAttributes.TenantCreated == "false" && messageAttributes.FluxCreated == "false" {
		subscriptionText = "createTenant"
	} else if messageAttributes.GroupsCreated == "true" && messageAttributes.WorkspaceCreated == "true" && messageAttributes.TenantCreated == "true" && messageAttributes.FluxCreated == "false" {
		subscriptionText = "createFlux"
	} else if messageAttributes.GroupsCreated == "true" && messageAttributes.WorkspaceCreated == "true" && messageAttributes.TenantCreated == "true" && messageAttributes.FluxCreated == "true" {
		//Coming soon
		subscriptionText = "deliverEmail"
	} else {
		return "", fmt.Errorf("message attributes not set for known subscription")
	}
	deliveryText := fmt.Sprintf("%s: %s", "message will be delivered to subscription", subscriptionText)
	return deliveryText, nil
}

// If slice of array contains the string searched for
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
