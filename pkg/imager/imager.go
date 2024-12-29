package imager

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	helmcli "helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/registry"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/kubectl/pkg/scheme"
)

var settings = helmcli.New()

type ImageDetails struct {
	Name   string
	Layers int
	Size   int
}

type Imager struct {
	client *action.Install
}

func NewImager() *Imager {
	cfg := new(action.Configuration)

	registryClient, err := registry.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	cfg.RegistryClient = registryClient

	client := action.NewInstall(cfg)
	client.DryRun = true
	client.ReleaseName = "imager-utility"
	client.ClientOnly = true
	client.SkipSchemaValidation = true

	return &Imager{
		client: client,
	}
}

// extractImages retrieves images defined in a chart
//
// It looks for images in resources that have a container in their definition i.e Deployments
func (i Imager) extractImages(ctx context.Context, chart *chart.Chart) ([]string, error) {
	release, err := i.client.RunWithContext(ctx, chart, map[string]interface{}{})
	if err != nil {
		return nil, fmt.Errorf("Error rendering templates: %v", err)
	}

	reader := strings.NewReader(release.Manifest)

	decoder := yaml.NewYAMLOrJSONDecoder(reader, 1024)

	containerImages := []string{}

	for {
		rawObj := &runtime.Unknown{}
		if err := decoder.Decode(rawObj); err != nil {
			if err.Error() == "EOF" {
				break
			}

			return nil, fmt.Errorf("Failed to decode manifest: %v", err)
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
		case *corev1.Pod:
			for _, container := range o.Spec.Containers {
				containerImages = append(containerImages, container.Image)
			}
		default:
			continue
		}
	}

	return containerImages, nil
}

// getImageDetails returns details about a provided image from the registry
func (i Imager) getImageDetails(_ context.Context, containerImage string) (*ImageDetails, error) {
	ref, err := name.ParseReference(containerImage)
	if err != nil {
		return nil, fmt.Errorf("failed to parse image reference: %v", err)

	}

	image, err := remote.Image(ref)
	if err != nil {
		return nil, fmt.Errorf("failed to pull image: %v", err)
	}

	layers, err := image.Layers()
	if err != nil {
		return nil, fmt.Errorf("failed to get image layers: %v", err)
	}

	size, err := image.Size()
	if err != nil {
		return nil, fmt.Errorf("failed to get image size: %v", err)
	}

	details := ImageDetails{
		Name:   containerImage,
		Layers: len(layers),
		Size:   int(size),
	}

	return &details, nil
}

// GetChartImagesDetails returns details about images that are defined in a helm chart
func (i Imager) GetChartImagesDetails(ctx context.Context, chartURL string) ([]*ImageDetails, error) {
	chartPath, err := i.client.ChartPathOptions.LocateChart(chartURL, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to locate chart from provided path: %v", err)
	}

	chart, err := loader.Load(chartPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load chart: %v", err)

	}

	containerImages, err := i.extractImages(ctx, chart)
	if err != nil {
		return nil, fmt.Errorf("failed to get images from chart: %v", err)
	}

	if len(containerImages) == 0 {
		return nil, fmt.Errorf("no images in the provided chart :-)")
	}

	// removes duplicate images
	slices.Sort(containerImages)
	containerImages = slices.Compact(containerImages)

	images := []*ImageDetails{}

	for _, containerImage := range containerImages {
		imageInfo, err := i.getImageDetails(ctx, containerImage)
		if err != nil {
			return nil, fmt.Errorf("failed to get images from main chart: %v", err)
		}

		images = append(images, imageInfo)
	}

	return images, nil
}
