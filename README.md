#Manage Azure resources and resource groups with Go

This package demonstrates how to manage [resources and resource groups](bhttps://azure.microsoft.com/documentation/articles/resource-group-overview/#resource-groups) using Azure SDK for Go.

If you don't have a Microsoft Azure subscription you can get a FREE trial account [here](https://azure.microsoft.com/pricing/free-trial).

##Instructions

1. Create a [service principal](https://azure.microsoft.com/documentation/articles/resource-group-authenticate-service-principal-cli/). You will need the Tenant ID, Client ID and Client Secret for [authentication](https://github.com/Azure/azure-sdk-for-go/tree/master/arm#first-a-sidenote-authentication-and-the-azure-resource-manager), so keep them as soon as you get them.
2. Get your Azure Subscription ID using either of the methods mentioned below:
  - Get it through the [portal](portal.azure.com) in the subscriptions section.
  - Get it using the [Azure CLI](https://azure.microsoft.com/documentation/articles/xplat-cli-install/) with command `azure account show`.
  - Get it using [Azure Powershell](https://azure.microsoft.com/documentation/articles/powershell-install-configure/) whit cmdlet `Get-AzureRmSubscription`.
3. Set environment variables `AZURE_TENANT_ID = <TENANT_ID>`, `AZURE_CLIENT_ID = <CLIENT_ID>`, `AZURE_CLIENT_SECRET = <CLIENT_SECRET>` and `AZURE_SUBSCRIPTION_ID = <SUBSCRIPTION_ID>`.
4. Get the [Azure SDK for Go](https://github.com/Azure/azure-sdk-for-go) using command `go get -u github.com/Azure/azure-sdk-for-go`.
5. Get this sample using command `go get -u github.com/Azure-Samples/resource-manager-go-resources-and-groups`.
6. Compile and run the sample.

##What does resourceSample.go do?

The sample gets an authorization token, creates a new resource group, updates it, lists the resource groups, creates a resource using a generic template, updates the resource, lists the resources inside the resource group, exports all resources into a json template, deletes a resource, and deletes the resource group.

##Get credentials and token

The sample starts by getting an authorization token from the service principal using your credentials. This token should be included in clients.

```
	credentials := map[string]string{
		"AZURE_CLIENT_ID":       os.Getenv("AZURE_CLIENT_ID"),
		"AZURE_CLIENT_SECRET":   os.Getenv("AZURE_CLIENT_SECRET"),
		"AZURE_SUBSCRIPTION_ID": os.Getenv("AZURE_SUBSCRIPTION_ID"),
		"AZURE_TENANT_ID":       os.Getenv("AZURE_TENANT_ID")}
	if err := checkEnvVar(&credentials); err != nil {
		return err
	}
	oauthConfig, err := azure.PublicCloud.OAuthConfigForTenant(credentials["AZURE_TENANT_ID"])
	if err != nil {
		return err
	}
	token, err := azure.NewServicePrincipalToken(*oauthConfig, credentials["AZURE_CLIENT_ID"], credentials["AZURE_CLIENT_SECRET"], azure.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		return err
	}
```

###Create resource group

Then, the sample creates a GroupsClient, which creates a resource group.

```
groupClient := resources.NewGroupsClient(credentials["AZURE_SUBSCRIPTION_ID"])
groupClient.Authorizer = token
location := "westus"
resourceGroupName := "azure-sample-group"	resourceGroupParameters := resources.ResourceGroup{
	Location: &location}
if _, err := groupClient.CreateOrUpdate(resourceGroupName, resourceGroupParameters); err != nil {
	return err
}
```

###Update the resource group

The sample updates the resource group with tags.

```
tags := map[string]*string{
		"who rocks": to.StringPtr("golang"),
		"where":     to.StringPtr("on azure")}
resourceGroupParameters.Tags = &tags
_, err = groupClient.CreateOrUpdate(resourceGroupName, resourceGroupParameters)
```

###List resource groups in subscription

This list shows just 10 resource groups, but the result is pageable.

```
groupsList, err := groupClient.List("", to.Int32Ptr(10))
```

###Create a generic resource

In this sample, a Key Vault is created, but it can be any resource.

```
resourceClient := resources.NewClient(credentials["AZURE_SUBSCRIPTION_ID"])
resourceClient.Authorizer = token
resourceClient.APIVersion = "2015-06-01"
namespace, resourceType, vaultName := "Microsoft.KeyVault", "vaults", "azureSampleVault"
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
	return err
}
```

###Update the resource with tags

```
keyVaultParameters.Tags = &tags
_, err = resourceClient.CreateOrUpdate(resourceGroupName, namespace, "", resourceType, vaultName, keyVaultParameters)
```

###List resources inside the resource group

This list shows just 100 resources, but the result is pageable.

```
resourcesList, err := groupClient.ListResources(resourceGroupName, "", to.Int32Ptr(100))
```

###Export resource group template to a json file

Resources can be exported into a json file. The asterisk * indicates all resources should be exported. Later, the json file can be used for [template deployment](https://github.com/Azure-Samples/resource-manager-go-template-deployment).

```
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
```

###Delete a generic resource

```
_, err = resourceClient.Delete(resourceGroupName, namespace, "", resourceType, vaultName)
```

###Delete the resource group

```
_, err = groupClient.Delete(resourceGroupName, nil)
```

##Find documentation
- [First a Sidenote: Authentication and the Azure Resource Manager](https://github.com/Azure/azure-sdk-for-go/tree/master/arm#first-a-sidenote-authentication-and-the-azure-resource-manager)
- [Azure Resource Manager overview](https://azure.microsoft.com/documentation/articles/resource-group-overview/)

***

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/). For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.