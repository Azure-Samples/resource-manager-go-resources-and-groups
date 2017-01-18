// This package demonstrates how to manage resources and resource groups in Azure using Go.
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
)

var (
	groupName = "your-azure-sample-group"
	location  = "westus"

	tenantID     string
	namespace    = "Microsoft.KeyVault"
	resourceType = "vaults"
	resourceName = "golangrocksonazure"

	groupsClient    resources.GroupsClient
	resourcesClient resources.Client
)

func init() {
	subscriptionID := getEnvVarOrExit("AZURE_SUBSCRIPTION_ID")
	tenantID = getEnvVarOrExit("AZURE_TENANT_ID")

	oauthConfig, err := azure.PublicCloud.OAuthConfigForTenant(tenantID)
	onErrorFail(err, "OAuthConfigForTenant failed")

	clientID := getEnvVarOrExit("AZURE_CLIENT_ID")
	clientSecret := getEnvVarOrExit("AZURE_CLIENT_SECRET")
	spToken, err := azure.NewServicePrincipalToken(*oauthConfig, clientID, clientSecret, azure.PublicCloud.ResourceManagerEndpoint)
	onErrorFail(err, "NewServicePrincipalToken failed")

	createClients(subscriptionID, spToken)
}

func main() {
	rg := createResourceGroup()
	updateResourceGroup(rg)

	listResourceGroups()

	gr := createResource()
	updateResource(gr)

	listResources()
	exportTemplate()

	fmt.Print("Press enter to delete the resources created in this sample...")

	var input string
	fmt.Scanln(&input)

	deleteResource()
	deleteResourceGroup()
}

// createResourceGroup creates a resource group
func createResourceGroup() resources.ResourceGroup {
	fmt.Println("Create resource group")
	rg := resources.ResourceGroup{
		Location: to.StringPtr(location),
	}
	_, err := groupsClient.CreateOrUpdate(groupName, rg)
	onErrorFail(err, "CreateOrUpdate failed")

	return rg
}

// updateResourceGroups updates a resource roup
func updateResourceGroup(rg resources.ResourceGroup) {
	fmt.Println("Update resource group")
	rg.Tags = &map[string]*string{
		"who rocks": to.StringPtr("golang"),
		"where":     to.StringPtr("on azure"),
	}
	_, err := groupsClient.CreateOrUpdate(groupName, rg)
	onErrorFail(err, "CreateOrUpdateFailed")
}

// listResourceGroups lists all resource groups and prints them.
func listResourceGroups() {
	fmt.Println("List resource groups")
	groupsList, err := groupsClient.List("", nil)
	onErrorFail(err, "List failed")
	if groupsList.Value != nil && len(*groupsList.Value) > 0 {
		allGroupsList := []resources.ResourceGroup{}
		appendResourceGroups(&allGroupsList, groupsList, to.IntPtr(0))
		fmt.Println("Resource groups in subscription")
		for _, group := range allGroupsList {
			tags := "\n"
			if group.Tags == nil || len(*group.Tags) <= 0 {
				tags += "\t\tNo tags yet\n"
			} else {
				for k, v := range *group.Tags {
					tags += fmt.Sprintf("\t\t%s = %s\n", k, *v)
				}
			}
			fmt.Printf("Resource group '%s'\n", *group.Name)
			elements := map[string]interface{}{
				"ID":                 *group.ID,
				"Location":           *group.Location,
				"Provisioning state": *group.Properties.ProvisioningState,
				"Tags":               tags}
			for k, v := range elements {
				fmt.Printf("\t%s: %s\n", k, v)
			}
		}
	} else {
		fmt.Println("There aren't any resource groups")
	}
}

func appendResourceGroups(allGroupsList *[]resources.ResourceGroup, lastResults resources.ResourceGroupListResult, page *int) {
	for _, g := range *lastResults.Value {
		*allGroupsList = append(*allGroupsList, g)
	}
	if lastResults.NextLink != nil {
		lastResults, err := groupsClient.ListNextResults(lastResults)
		onErrorFail(err, fmt.Sprintf("ListNext failed on page %v", *page))
		*page++
		appendResourceGroups(allGroupsList, lastResults, page)
	}
}

// createResource creates a generic resource
func createResource() resources.GenericResource {
	fmt.Println("Create a Key Vault via Generic Resource Put")
	genericResource := resources.GenericResource{
		Location: to.StringPtr(location),
		Properties: &map[string]interface{}{
			"sku": map[string]string{
				"Family": "A",
				"Name":   "standard",
			},
			"tenantID":             tenantID,
			"accessPolicies":       []string{},
			"enabledForDeployment": true,
		},
	}
	_, err := resourcesClient.CreateOrUpdate(groupName, namespace, "", resourceType, resourceName, genericResource, nil)
	onErrorFail(err, "CreateOrUpdate failed")

	return genericResource
}

// updateResource updates a generic resource
func updateResource(gr resources.GenericResource) {
	fmt.Println("Update resource")
	gr.Tags = &map[string]*string{
		"who rocks": to.StringPtr("golang"),
		"where":     to.StringPtr("on azure"),
	}
	_, err := resourcesClient.CreateOrUpdate(groupName, namespace, "", resourceType, resourceName, gr, nil)
	onErrorFail(err, "CreateOrUpdate failed")
}

// listResources lists all resources inside a resource group and prints them.
func listResources() {
	fmt.Println("List resources inside the resource group")
	resourcesList, err := groupsClient.ListResources(groupName, "", "", nil)
	onErrorFail(err, "ListResources failed")
	if resourcesList.Value != nil && len(*resourcesList.Value) > 0 {
		fmt.Printf("Resources in '%s' resource group\n", groupName)
		for _, resource := range *resourcesList.Value {
			tags := "\n"
			if resource.Tags == nil || len(*resource.Tags) <= 0 {
				tags += "\t\t\tNo tags yet\n"
			} else {
				for k, v := range *resource.Tags {
					tags += fmt.Sprintf("\t\t\t%s = %s\n", k, *v)
				}
			}
			fmt.Printf("\tResource '%s'\n", *resource.Name)
			elements := map[string]interface{}{
				"ID":       *resource.ID,
				"Location": *resource.Location,
				"Type":     *resource.Type,
				"Tags":     tags,
			}
			for k, v := range elements {
				fmt.Printf("\t\t%s: %s\n", k, v)
			}
		}
	} else {
		fmt.Printf("There aren't any resources inside '%s' resource group\n", groupName)
	}
}

// exportTemplate saves the resource group template in a json file.
func exportTemplate() {
	fmt.Println("Export resource group template")
	// The asterisk * indicates all resources should be exported.
	expReq := resources.ExportTemplateRequest{
		Resources: &[]string{"*"},
	}
	template, err := groupsClient.ExportTemplate(groupName, expReq)
	onErrorFail(err, "ExportTemplate failed")

	prefix, indent := "", "    "
	exported, err := json.MarshalIndent(template, prefix, indent)
	onErrorFail(err, "MarshalIndent failed")

	fileTemplate := "%s-template.json"
	fileName := fmt.Sprintf(fileTemplate, groupName)
	if _, err := os.Stat(fileName); err == nil {
		onErrorFail(fmt.Errorf("File '%s' already exists", fileName), "Saving JSON file failed")
	}
	ioutil.WriteFile(fileName, exported, 0666)
	onErrorFail(err, "WriteFile failed")

	fmt.Printf("The resource group template has been saved to %s\n", fmt.Sprintf(fileTemplate, groupName))
}

// deleteResource deletes a generic resource
func deleteResource() {
	fmt.Println("Delete a resource")
	_, err := resourcesClient.Delete(groupName, namespace, "", resourceType, resourceName, nil)
	onErrorFail(err, "Delete failed")
}

// deleteResourceGroup deletes a resource group
func deleteResourceGroup() {
	fmt.Println("Delete resource group")
	_, err := groupsClient.Delete(groupName, nil)
	onErrorFail(err, "Delete failed")
}

// getEnvVarOrExit returns the value of specified environment variable or terminates if it's not defined.
func getEnvVarOrExit(varName string) string {
	value := os.Getenv(varName)
	if value == "" {
		fmt.Printf("Missing environment variable %s\n", varName)
		os.Exit(1)
	}

	return value
}

// onErrorFail prints a failure message and exits the program if err is not nil.
// it also deletes the resource group created in the sample
func onErrorFail(err error, message string) {
	if err != nil {
		fmt.Printf("%s: %s\n", message, err)
		groupsClient.Delete(groupName, nil)
		os.Exit(1)
	}
}

func createClients(subscriptionID string, spToken *azure.ServicePrincipalToken) {
	groupsClient = resources.NewGroupsClient(subscriptionID)
	groupsClient.Authorizer = spToken

	resourcesClient = resources.NewClient(subscriptionID)
	resourcesClient.Authorizer = spToken
	resourcesClient.APIVersion = "2015-06-01"
}
