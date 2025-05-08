module github.com/Azure/karpenter-provider-azure

go 1.24.2

require (
	github.com/Azure/aks-middleware v0.0.34
	github.com/Azure/azure-kusto-go v0.16.1
	github.com/Azure/azure-sdk-for-go v68.0.0+incompatible
	github.com/Azure/azure-sdk-for-go-extensions v0.1.8
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.18.0
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.9.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute v1.0.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5 v5.7.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v4 v4.8.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork v1.1.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resourcegraph/armresourcegraph v0.9.0
	github.com/Azure/go-autorest/autorest v0.11.30
	github.com/Azure/go-autorest/autorest/to v0.4.1
	github.com/Azure/skewer v0.0.19
	github.com/Pallinder/go-randomdata v1.2.0
	github.com/alecthomas/units v0.0.0-20211218093645-b94a6e3cc137
	github.com/awslabs/operatorpkg v0.0.0-20250320000002-b05af0f15c68
	github.com/blang/semver/v4 v4.0.0
	github.com/go-openapi/errors v0.22.1
	github.com/go-openapi/runtime v0.28.0
	github.com/go-openapi/strfmt v0.23.0
	github.com/go-openapi/swag v0.23.1
	github.com/go-openapi/validate v0.24.0
	github.com/go-playground/validator/v10 v10.26.0
	github.com/google/uuid v1.6.0
	github.com/imdario/mergo v0.3.16
	github.com/jongio/azidext/go/azidext v0.5.0
	github.com/mitchellh/hashstructure/v2 v2.0.2
	github.com/onsi/ginkgo/v2 v2.23.4
	github.com/onsi/gomega v1.37.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/prometheus/client_golang v1.22.0
	github.com/samber/lo v1.50.0
	github.com/stretchr/testify v1.10.0
	go.uber.org/multierr v1.11.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.32.3
	k8s.io/apiextensions-apiserver v0.32.3
	k8s.io/apimachinery v0.32.3
	k8s.io/client-go v0.32.3
	k8s.io/klog/v2 v2.130.1
	k8s.io/utils v0.0.0-20250321185631-1f6e0b77f77e
	sigs.k8s.io/cloud-provider-azure v1.32.4
	sigs.k8s.io/controller-runtime v0.20.4
	sigs.k8s.io/karpenter v1.4.0
)

require (
	github.com/Azure/go-autorest/autorest/adal v0.9.24 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.2 // indirect
)
