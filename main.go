package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
)

var (
	filename string
)

func main() {
	flag.StringVar(&filename, "f", "", "")
	flag.Parse()

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	decoder := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader(b), 1000)
	for {
		var rawObj runtime.RawExtension
		if err = decoder.Decode(&rawObj); err != nil {
			break
		}

		// decode the yaml into the object and gvk of the CRD.
		// The gvk can be used to determine whether the crd is using the correct
		// version of the CustomResourceDefinition resource
		// (i.e. the proper "v1" version or the deprecated "v1beta1" version)
		obj, gvk, err := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(rawObj.Raw, nil, nil)
		if err != nil {
			log.Fatal(err)
		}
		apiVersion := gvk.Version
		unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		if err != nil {
			log.Fatal(err)
		}

		// retrieve the CRDs group from .spec.group
		spec, ok := unstructuredMap["spec"]
		if !ok {
			log.Fatal(errors.New("no spec in unstructuredMap"))
		}
		specMap := spec.(map[string]interface{})
		group, ok := (specMap["group"]).(string)
		if !ok {
			log.Fatal(errors.New("group is not a string"))
		}

		// retrieve the CRD's kind from .spec.names.kind
		names := specMap["names"]
		namesMap := names.(map[string]interface{})
		kind := namesMap["kind"]

		// determine if the source of the CRD is 1P or OSS based on whether any
		// google specific terms exist in the CRD's group
		src := "OSS"
		if strings.Contains(group, "gke") || strings.Contains(group, "google") {
			src = "1P"
		}

		// for every version check if at least some openAPIV3Schema exists
		// at .spec.versions[*].schema.openAPIV3Schema
		versions := specMap["versions"]
		vSlice := versions.([]interface{})
		for _, v := range vSlice {
			vMap := v.(map[string]interface{})
			vName := vMap["name"]
			schema := vMap["schema"]
			hasSchema := false
			if schema != nil {
				schemaMap := schema.(map[string]interface{})
				openAPIV3Schema := schemaMap["openAPIV3Schema"]
				if openAPIV3Schema != nil {
					hasSchema = true
				}
			}
			// print the csv line
			fmt.Printf("%s, %s, %s, %s,,, %s, %t\n", group, vName, kind, src, apiVersion, hasSchema)
		}
	}
	if err != io.EOF {
		log.Fatal("eof ", err)
	}
}
