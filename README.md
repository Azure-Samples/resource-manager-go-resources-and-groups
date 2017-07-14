---
services: azure-resource-manager
platforms: go
author: mcardosos
---

# Manage Azure resources and resource groups with Go

This package demonstrates how to manage [resources and resource groups](bhttps://azure.microsoft.com/documentation/articles/resource-group-overview/#resource-groups) using Azure SDK for Go.

If you don't have a Microsoft Azure subscription you can get a FREE trial account [here](https://azure.microsoft.com/pricing/free-trial).

**On this page**

- [Run this sample](#run)
- [What does example.go do?](#sample)
- [More information](#info)

<a id="run"></a>

## Run this sample

1. If you don't already have it, [install Go 1.7](https://golang.org/dl/).

1. Get the sample. You can use either

	```
	go get github.com/Azure-Samples/resource-manager-go-resources-and-groups
	```

	or

    ```
    git clone https://github.com:Azure-Samples/virtual-machines-go-manage.git
    ```

1. Install the dependencies using [glide](https://github.com/Masterminds/glide).

    ```
    cd virtual-machines-go-manage
    glide install
    ```

1. Create an Azure service principal either through
    [Azure CLI](https://azure.microsoft.com/documentation/articles/resource-group-authenticate-service-principal-cli/),
    [PowerShell](https://azure.microsoft.com/documentation/articles/resource-group-authenticate-service-principal/)
    or [the portal](https://azure.microsoft.com/documentation/articles/resource-group-create-service-principal-portal/).

1. Set the following environment variables using the information from the service principle that you created.

    ```
    export AZURE_TENANT_ID={your tenant id}
    export AZURE_CLIENT_ID={your client id}
    export AZURE_CLIENT_SECRET={your client secret}
    export AZURE_SUBSCRIPTION_ID={your subscription id}
    ```

    > [AZURE.NOTE] On Windows, use `set` instead of `export`.

1. Run the sample.

    ```
    go run example.go
    ```


<a id="sample"></a>

## What does example.go do?

### Create resource group

```go
rg.Tags = &map[string]*string{
	"who rocks": to.StringPtr("golang"),
	"where":     to.StringPtr("on azure"),
}
_, err := groupsClient.CreateOrUpdate(groupName, rg)
```

### Update the resource group

The sample updates the resource group with tags.

```go
rg.Tags = &map[string]*string{
	"who rocks": to.StringPtr("golang"),
	"where":     to.StringPtr("on azure"),
}
_, err := groupsClient.CreateOrUpdate(groupName, rg)
```

### List resource groups in subscription

```go
groupsList, err := groupClient.List("", nil)
```

### Create a generic resource

In this sample, a Key Vault is created, but it can be any resource.

```go
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
```

### Update the resource with tags

```go
gr.Tags = &map[string]*string{
	"who rocks": to.StringPtr("golang"),
	"where":     to.StringPtr("on azure"),
}
_, err := resourcesClient.CreateOrUpdate(groupName, namespace, "", resourceType, resourceName, gr, nil)
```

### List resources inside the resource group

```go
resourcesList, err := groupsClient.ListResources(groupName, "", "", nil)
```

### Export resource group template to a json file

Resources can be exported into a json file. The asterisk * indicates all resources should be exported. Later, the json file can be used for [template deployment](https://github.com/Azure-Samples/resource-manager-go-template-deployment).

```go
// The asterisk * indicates all resources should be exported.
expReq := resources.ExportTemplateRequest{
	ResourcesProperty: &[]string{"*"},
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
```

### Delete a generic resource

```go
resourcesClient.Delete(groupName, namespace, "", resourceType, resourceName, nil)
```

### Delete the resource group

```go
groupsClient.Delete(groupName, nil)
```

<a id="info"></a>

## More information

- [First a Sidenote: Authentication and the Azure Resource Manager](https://github.com/Azure/azure-sdk-for-go/tree/master/arm#first-a-sidenote-authentication-and-the-azure-resource-manager)
- [Azure Resource Manager overview](https://azure.microsoft.com/documentation/articles/resource-group-overview/)

***

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/). For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.