package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	rfv1 "git.v87.us/formicary/refunc/pkg/apis/refunc/v1"
	"github.com/rancher/norman/api"
	"github.com/rancher/norman/pkg/subscribe"
	"github.com/rancher/norman/store/proxy"
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/factory"
	"github.com/rancher/norman/types/mapper"
	"github.com/rancher/norman/types/values"
	"github.com/rancher/norman/urlbuilder"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// VERSION of current binary
var VERSION = "v0.1.0-dev"

func main() {
	app := cli.NewApp()
	app.Name = "refunc-rancher"
	app.Version = VERSION
	app.Usage = "You need help!"
	app.Action = func(c *cli.Context) error {
		ctx := context.Background()

		if os.Getenv("REFUNC_DEBUG") == "true" {
			logrus.SetLevel(logrus.DebugLevel)
		}

		cfgPath := os.Getenv("KUBECONFIG")
		if cfgPath == "" {
			cfgPath = filepath.Join(homedir.HomeDir(), ".kube/config")
			if _, err := os.Stat(cfgPath); err != nil {
				// fallback to guess config using InClusterConfig
				cfgPath = ""
			}
		}
		kubeConfig, err := clientcmd.BuildConfigFromFlags("", cfgPath)
		if err != nil {
			panic(err)
		}

		k8sClient, err := proxy.NewClientGetterFromConfig(*kubeConfig)
		if err != nil {
			panic(err)
		}

		version := types.APIVersion{
			Version: rfv1.SchemeGroupVersion.Version,
			Group:   rfv1.SchemeGroupVersion.Group,
			Path:    "/refunc/v1",
		}

		schemas := newSchemas(&version)
		subscribe.Register(&version, schemas)

		// funceves
		schemas.MustImport(&version, rfv1.FuncdefSpec{}, struct {
			Meta map[string]interface{} `json:"meta"`
		}{}).MustImportAndCustomize(&version, rfv1.Funcdef{}, func(schema *types.Schema) {
			schema.PluralName = rfv1.FuncdefPluralName
			if err := assignStores(ctx, k8sClient, types.DefaultStorageContext, schema, rfv1.CRDs[0].CRD); err != nil {
				panic(err)
			}
		}, namespacedType)

		// xenvs
		schemas.AddMapperForType(&version, rfv1.XenvSpec{},
			mapper.Move{From: "type", To: "xenvType"},
		).MustImportAndCustomize(&version, rfv1.Xenv{}, func(schema *types.Schema) {
			if err := assignStores(ctx, k8sClient, types.DefaultStorageContext, schema, rfv1.CRDs[1].CRD); err != nil {
				panic(err)
			}
		}, namespacedType)

		// triggers
		schemas.AddMapperForType(&version, rfv1.TriggerSpec{},
			mapper.Enum{Field: "type", Options: []string{
				"eventgateway",
				"cron",
				"http",
			}},
			mapper.Move{From: "type", To: "triggerType"},
		).MustImportAndCustomize(&version, rfv1.Trigger{}, func(schema *types.Schema) {
			if err := assignStores(ctx, k8sClient, types.DefaultStorageContext, schema, rfv1.CRDs[2].CRD); err != nil {
				panic(err)
			}
		}, namespacedType)

		// funcinsts
		schemas.AddMapperForType(&version, rfv1.Funcinst{},
			labelToField{LabelField: "name", Field: "funcdefName"},
			fillFundefIDField{},
		).MustImportAndCustomize(&version, rfv1.Funcinst{}, func(schema *types.Schema) {
			schema.CollectionMethods = []string{http.MethodGet}
			schema.ResourceMethods = []string{http.MethodGet, http.MethodDelete}
			if err := assignStores(ctx, k8sClient, types.DefaultStorageContext, schema, rfv1.CRDs[3].CRD); err != nil {
				panic(err)
			}
		}, namespacedType, struct {
			FuncdefName string `json:"funcdefName"`
			FuncdefID   string `json:"funcdefId"`
		}{})

		server := api.NewAPIServer()
		if err := server.AddSchemas(schemas); err != nil {
			panic(err)
		}

		fmt.Println("Listening on 0.0.0.0:1234")
		http.ListenAndServe("0.0.0.0:1234", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			// handle k8s proxy forward X-Forwarded-Uri
			if fwdURI := req.Header.Get("X-Forwarded-Uri"); fwdURI != "" {
				apiPrefix := strings.TrimSuffix(fwdURI, req.URL.Path)
				if clusterID := os.Getenv("RANCHER_CLUSTER_ID"); clusterID != "" {
					apiPrefix = path.Join("/k8s/clusters", clusterID, apiPrefix)
				}
				req.Header.Add(urlbuilder.PrefixHeader, apiPrefix)
			}
			server.ServeHTTP(rw, req)
		}))
		return nil
	}

	app.Run(os.Args)
}

func assignStores(ctx context.Context, ClientGetter proxy.ClientGetter, storageContext types.StorageContext, schema *types.Schema, crd *apiextensionsv1beta1.CustomResourceDefinition) error {
	schema.Store = proxy.NewProxyStore(ctx, ClientGetter,
		storageContext,
		[]string{"apis"},
		crd.Spec.Group,
		crd.Spec.Version,
		crd.Spec.Names.Kind,
		crd.Spec.Names.Plural)

	return nil
}

func newSchemas(version *types.APIVersion) *types.Schemas {
	schemas := factory.Schemas(version)
	baseFunc := schemas.DefaultMappers
	schemas.DefaultMappers = func() []types.Mapper {
		mappers := append(baseFunc(), &mapper.Scope{
			If: types.NamespaceScope,
			Mappers: []types.Mapper{
				&mapper.Move{
					From: "namespace",
					To:   "namespaceId",
				},
				&mapper.ReadOnly{
					Field:    "namespaceId",
					Optional: false,
				},
			},
		})
		return mappers
	}
	return schemas
}

type namespacedOverride struct {
	types.Namespaced
}

var namespacedType = namespacedOverride{}

type labelToField struct {
	LabelField string
	Field      string
}

func (e labelToField) FromInternal(data map[string]interface{}) {
	if e.LabelField == "" {
		e.LabelField = e.Field
	}
	v, ok := values.RemoveValue(data, "labels", "refunc.io/"+e.LabelField)
	if ok {
		data[e.Field] = v
	}
}

func (e labelToField) ToInternal(data map[string]interface{}) error {
	v, ok := data[e.Field]
	if ok {
		if e.LabelField == "" {
			e.LabelField = e.Field
		}
		values.PutValue(data, v, "labels", "refunc.io/"+e.LabelField)
	}
	return nil
}

func (e labelToField) ModifySchema(schema *types.Schema, schemas *types.Schemas) error {
	return mapper.ValidateField(e.Field, schema)
}

type fillFundefIDField struct{}

func (e fillFundefIDField) FromInternal(data map[string]interface{}) {
	data["funcdefId"] = fmt.Sprintf("%s:%s", data["namespaceId"], data["funcdefName"])
}

func (e fillFundefIDField) ToInternal(data map[string]interface{}) error {
	values.RemoveValue(data, "funcdefId")
	return nil
}

func (e fillFundefIDField) ModifySchema(schema *types.Schema, schemas *types.Schemas) error {
	return nil
}
