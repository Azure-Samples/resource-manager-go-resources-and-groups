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

func main() {
	fmt.Println("Get credentials and token...")
	credentials := map[string]string{
		"AZURE_CLIENT_ID":       os.Getenv("AZURE_CLIENT_ID"),
		"AZURE_CLIENT_SECRET":   os.Getenv("AZURE_CLIENT_SECRET"),
		"AZURE_SUBSCRIPTION_ID": os.Getenv("AZURE_SUBSCRIPTION_ID"),
		"AZURE_TENANT_ID":       os.Getenv("AZURE_TENANT_ID")}
	if err := checkEnvVar(&credentials); err != nil {
		printError(err)
		return
	}
	oauthConfig, err := azure.PublicCloud.OAuthConfigForTenant(credentials["AZURE_TENANT_ID"])
	if err != nil {
		printError(err)
		return
	}
	token, err := azure.NewServicePrincipalToken(*oauthConfig, credentials["AZURE_CLIENT_ID"], credentials["AZURE_CLIENT_SECRET"], azure.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		printError(err)
		return
	}

	groupClient := resources.NewGroupsClient(credentials["AZURE_SUBSCRIPTION_ID"])
	groupClient.Authorizer = token

	location := "westus"
	resourceGroupName := "azure-sample-group"
	tags := map[string]*string{
		"who rocks": to.StringPtr("golang"),
		"where":     to.StringPtr("on azure")}

	fmt.Println("Create resource group...")
	resourceGroupParameters := resources.ResourceGroup{
		Location: &location}
	if _, err := groupClient.CreateOrUpdate(resourceGroupName, resourceGroupParameters); err != nil {
		printError(err)
		return
	}

	fmt.Println("Update resource group...")
	resourceGroupParameters.Tags = &tags
	_, err = groupClient.CreateOrUpdate(resourceGroupName, resourceGroupParameters)
	printError(err)

	fmt.Println("List resource groups...")
	printError(listResourceGroups(&groupClient))

	namespace, resourceType, vaultName := "Microsoft.KeyVault", "vaults", "azureSampleVault"

	fmt.Println("Create a Key Vault via Generic Resource Put...")
	resourceClient := resources.NewClient(credentials["AZURE_SUBSCRIPTION_ID"])
	resourceClient.Authorizer = token
	resourceClient.APIVersion = "2015-06-01"
	keyVaultParameters := resources.GenericResource{
		Location: &location,
		Properties: &map[string]interface{}{
			"sku": map[string]string{
				"Family": "A",
				"Name":   "standard"},
			"tenantID":             credentials["AZURE_TENANT_ID"],
			"accessPolicies":       []string{},
			"enabledForDeployment": true}}
	if _, err = resourceClient.CreateOrUpdate(resourceGroupName, namespace, "", resourceType, vaultName, keyVaultParameters); err != nil {
		printError(err)
		return
	}

	fmt.Println("Update resource...")
	keyVaultParameters.Tags = &tags
	_, err = resourceClient.CreateOrUpdate(resourceGroupName, namespace, "", resourceType, vaultName, keyVaultParameters)
	printError(err)

	fmt.Println("List resources inside the resource group...")
	printError(listResources(&groupClient, resourceGroupName))

	fmt.Println("Export resource group template...")
	printError(exportTemplate(&groupClient, resourceGroupName))

	fmt.Println("Delete a resource...")
	_, err = resourceClient.Delete(resourceGroupName, namespace, "", resourceType, vaultName)
	printError(err)

	fmt.Println("Delete resource group...")
	if _, err := groupClient.Delete(resourceGroupName, nil); err != nil {
		printError(err)
		return
	}
}

// checkEnvVar checks if the environment variables are actually set.
func checkEnvVar(envVars *map[string]string) error {
	var missingVars []string
	for varName, value := range *envVars {
		if value == "" {
			missingVars = append(missingVars, varName)
		}
	}
	if len(missingVars) > 0 {
		return fmt.Errorf("Missing environment variables %v", missingVars)
	}
	return nil
}

// listResourceGroups lists all resource groups and prints them.
func listResourceGroups(groupClient *resources.GroupsClient) error {
	groupsList, err := groupClient.List("", to.Int32Ptr(10))
	if err != nil {
		return err
	}
	if len(*groupsList.Value) > 0 {
		fmt.Println("Resource groups in subscription")
		for _, group := range *groupsList.Value {
			tags := "\n"
			if group.Tags == nil || len(*group.Tags) <= 0 {
				tags += "\t\tNo tags yet\n"
			} else {
				for k, v := range *group.Tags {
					tags += fmt.Sprintf("\t\t%v = %v\n", k, *v)
				}
			}
			fmt.Printf("Resource group '%v'\n", *group.Name)
			elements := map[string]interface{}{
				"ID":                 *group.ID,
				"Location":           *group.Location,
				"Provisioning state": *group.Properties.ProvisioningState,
				"Tags":               tags}
			for k, v := range elements {
				fmt.Printf("\t%v: %v\n", k, v)
			}
		}
	} else {
		fmt.Println("There aren't any resource groups")
	}
	return nil
}

// listResources lists all resources inside a resource group and prints them.
func listResources(groupClient *resources.GroupsClient, resourceGroupName string) error {
	resourcesList, err := groupClient.ListResources(resourceGroupName, "", "", to.Int32Ptr(100))
	if err != nil {
		return err
	}
	if len(*resourcesList.Value) > 0 {
		fmt.Printf("Resources in '%v' resource group\n", resourceGroupName)
		for _, resource := range *resourcesList.Value {
			tags := "\n"
			if resource.Tags == nil || len(*resource.Tags) <= 0 {
				tags += "\t\t\tNo tags yet\n"
			} else {
				for k, v := range *resource.Tags {
					tags += fmt.Sprintf("\t\t\t%v = %v\n", k, *v)
				}
			}
			fmt.Printf("\tResource '%v'\n", *resource.Name)
			elements := map[string]interface{}{
				"ID":       *resource.ID,
				"Location": *resource.Location,
				"Type":     *resource.Type,
				"Tags":     tags}
			for k, v := range elements {
				fmt.Printf("\t\t%v: %v\n", k, v)
			}
		}
	} else {
		fmt.Printf("There aren't any resources inside '%v' resource group\n", resourceGroupName)
	}
	return nil
}

// exportTemplate saves the resource group template in a json file.
func exportTemplate(groupClient *resources.GroupsClient, resourceGroupName string) error {
	// The asterisk * indicates all resources should be exported.
	expReq := resources.ExportTemplateRequest{
		Resources: &[]string{"*"}}
	template, err := groupClient.ExportTemplate(resourceGroupName, expReq)
	if err != nil {
		return err
	}
	exported, err := json.MarshalIndent(template, "", "    ")
	if err != nil {
		return err
	}
	fileName := fmt.Sprintf("%v-template.json", resourceGroupName)
	if _, err := os.Stat(fileName); err == nil {
		return fmt.Errorf("File '%v' already exists", fileName)
	}
	return ioutil.WriteFile(fileName, exported, 0666)

}

// printError prints non nil errors.
func printError(err error) {
	if err != nil {
		fmt.Printf("Error! %v\n", err)
	}
}
