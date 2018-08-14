package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	rfv1 "git.v87.us/formicary/refunc/pkg/apis/refunc/v1"
	"github.com/rancher/norman/api"
	"github.com/rancher/norman/pkg/subscribe"
	"github.com/rancher/norman/store/proxy"
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/factory"
	"github.com/rancher/norman/types/mapper"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// VERSION of current binary
var VERSION = "v0.1.0-dev"

var (
	version = types.APIVersion{
		Version: rfv1.SchemeGroupVersion.Version,
		Group:   rfv1.SchemeGroupVersion.Group,
		Path:    fmt.Sprintf("/refunc/v1"),
	}

	schemas = factory.Schemas(&version)
)

func init() {
	// Only log the warning severity or above.
	logrus.SetLevel(logrus.DebugLevel)
}

func main() {
	app := cli.NewApp()
	app.Name = "refunc-rancher"
	app.Version = VERSION
	app.Usage = "You need help!"
	app.Action = func(c *cli.Context) error {
		ctx := context.Background()

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

		subscribe.Register(&version, schemas)

		// funceves
		schemas.MustImport(&version, rfv1.FuncdefSpec{}, struct {
			Meta map[string]interface{} `json:"meta"`
		}{}).MustImportAndCustomize(&version, rfv1.Funcdef{}, func(schema *types.Schema) {
			schema.Scope = types.NamespaceScope
			schema.PluralName = rfv1.FuncdefPluralName
			if err := assignStores(ctx, k8sClient, types.DefaultStorageContext, schema, rfv1.CRDs[0].CRD); err != nil {
				panic(err)
			}
		})

		// xenvs
		schemas.AddMapperForType(&version, rfv1.XenvSpec{},
			mapper.Move{From: "type", To: "xenvType"},
		).MustImportAndCustomize(&version, rfv1.Xenv{}, func(schema *types.Schema) {
			schema.Scope = types.NamespaceScope
			if err := assignStores(ctx, k8sClient, types.DefaultStorageContext, schema, rfv1.CRDs[1].CRD); err != nil {
				panic(err)
			}
		})

		// triggers
		schemas.AddMapperForType(&version, rfv1.TriggerSpec{},
			mapper.Enum{Field: "type", Options: []string{
				"eventgateway",
				"cron",
				"http",
			}},
			mapper.Move{From: "type", To: "triggerType"},
		).MustImportAndCustomize(&version, rfv1.Trigger{}, func(schema *types.Schema) {
			schema.Scope = types.NamespaceScope
			if err := assignStores(ctx, k8sClient, types.DefaultStorageContext, schema, rfv1.CRDs[2].CRD); err != nil {
				panic(err)
			}
		})

		// funcinsts
		schemas.MustImportAndCustomize(&version, rfv1.Funcinst{}, func(schema *types.Schema) {
			schema.Scope = types.NamespaceScope
			schema.CollectionMethods = []string{http.MethodGet}
			schema.ResourceMethods = []string{http.MethodGet, http.MethodDelete}
			if err := assignStores(ctx, k8sClient, types.DefaultStorageContext, schema, rfv1.CRDs[3].CRD); err != nil {
				panic(err)
			}
		})

		server := api.NewAPIServer()
		if err := server.AddSchemas(schemas); err != nil {
			panic(err)
		}

		fmt.Println("Listening on 0.0.0.0:1234")
		http.ListenAndServe("0.0.0.0:1234", server)
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
