It is a simple utility that uses a link to a helm chart to search for container images defined in the chart, downloads the respective Docker images from their repositories, and prints info about the images i.e their sizes and number of layers.

```bash
go run main.go https://github.com/openfga/helm-charts/releases/download/openfga-0.2.19/openfga-0.2.19.tgz
```
