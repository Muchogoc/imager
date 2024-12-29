package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/kubectl/pkg/scheme"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Println("Expected a chart URL to be provided")
		os.Exit(1)
	}

	chartURL := args[0]

	settings := cli.New()
	actionConfig := new(action.Configuration)
	client := action.NewInstall(actionConfig)
	client.DryRun = true
	client.ReleaseName = "imager-utility"
	client.ClientOnly = true

	chartPath, err := client.ChartPathOptions.LocateChart(chartURL, settings)
	if err != nil {
		fmt.Printf("Error locating chart: %v\n", err)
		os.Exit(1)
	}

	chart, err := loader.Load(chartPath)
	if err != nil {
		fmt.Printf("Error loading chart: %v\n", err)
		os.Exit(1)
	}

	rel, err := client.Run(chart, map[string]interface{}{})
	if err != nil {
		fmt.Printf("Error rendering templates: %v\n", err)
		os.Exit(1)
	}

	reader := strings.NewReader(rel.Manifest)

	decoder := yaml.NewYAMLOrJSONDecoder(reader, 1024)

	containerImages := []string{}

	for {
		rawObj := &runtime.Unknown{}
		if err := decoder.Decode(rawObj); err != nil {
			if err.Error() == "EOF" {
				break
			}

			fmt.Printf("Failed to decode manifest: %v\n", err)
			os.Exit(1)
		}

		obj, _, err := scheme.Codecs.UniversalDeserializer().Decode(rawObj.Raw, nil, nil)
		if err != nil {
			continue
		}

		switch o := obj.(type) {
		case *appsv1.Deployment:
			for _, container := range o.Spec.Template.Spec.Containers {
				containerImages = append(containerImages, container.Image)
			}
		default:
			continue
		}
	}

	if len(containerImages) == 0 {
		fmt.Printf("Imager found no images :-)")
		os.Exit(1)
	}

	type imager struct {
		Name   string
		Layers int
		Size   int
	}

	imagers := []imager{}

	for _, containerImage := range containerImages {
		ref, err := name.ParseReference(containerImage)
		if err != nil {
			fmt.Printf("Failed to parse image reference: %v\n", err)
			os.Exit(1)
		}

		image, err := remote.Image(ref)
		if err != nil {
			fmt.Printf("Failed to pull image: %v\n", err)
			os.Exit(1)
		}

		layers, _ := image.Layers()
		size, _ := image.Size()

		imagers = append(imagers, imager{Name: containerImage, Layers: len(layers), Size: int(size)})
	}

	for _, imager := range imagers {
		fmt.Printf("Name: %s\tLayers: %d\tSize: %d", imager.Name, imager.Layers, imager.Size)
	}
}
