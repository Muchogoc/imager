# Imager

A simple utility that uses a link to a Helm chart to search for container images defined in the chart, downloads the respective images and outputs info about the images i.e., their sizes and number of layers.

## Features

- Extracts container images from Helm charts.
- Downloads the specified images.
- Displays image size.
- Displays the number of image layers.

## Installation

### Prerequisites

- Go 1.23 or higher

### Building and Running

1. Build the `imager` binary

   ```bash
   go build -o imager ./cmd/main.go
   ```
   This command will create an executable file named imager in your current directory.

2. Run imager with the desired option
    ```bash
    ./imager <subcommand> <arguments>
    ```
    Available subcommands:
    - `list`: A CLI to show image information.
    Example
    ```bash
    ./imager list https://github.com/openfga/helm-charts/releases/download/openfga-0.2.19/openfga-0.2.19.tgz
    ```
    - `http-server`: Starts a HTTP server that exposes an API to show image information.
    ```bash
    ./imager http-server
    ```
    Example request:
    ```bash
    curl --location 'http://localhost:8080' --header 'Content-Type: application/json' --data '{"chart_url": "oci://registry-1.docker.io/bitnamicharts/airflow"}'
    ```

    