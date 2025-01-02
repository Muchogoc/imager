package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Muchogoc/imager/pkg/imager"
	"github.com/urfave/cli/v2"
)

func main() {
	i := imager.NewImager()

	app := &cli.App{
		Name:  "Imager",
		Usage: "Show details about images in a helm chart",
		Commands: []*cli.Command{
			{
				Name:  "list",
				Usage: "Run imager CLI by providing a link to the helm chart",
				Action: func(cCtx *cli.Context) error {
					if cCtx.Args().Len() == 0 {
						return cli.Exit("Expected a link to a helm chart", 1)
					}

					chartURL := cCtx.Args().First()

					images, err := i.GetChartImagesDetails(cCtx.Context, chartURL)
					if err != nil {
						return cli.Exit(err.Error(), 1)
					}

					for _, imager := range images {
						fmt.Printf("Name: %s\tLayers: %d\tSize: %f\n", imager.Name, imager.Layers, imager.Size)
					}

					return nil
				},
			},
			{
				Name:  "http-server",
				Usage: "Run imager http server",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "port",
						Aliases: []string{"p"},
						Usage:   "Port to run the HTTP server on",
						Value:   "8080",
						EnvVars: []string{"SERVER_PORT"},
					},
				},
				Action: func(cCtx *cli.Context) error {
					type input struct {
						ChartURL string `json:"chart_url"`
					}

					http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set("Content-Type", "application/json")

						switch r.Method {
						case http.MethodGet:
							w.WriteHeader(http.StatusOK)
							msg := struct {
								Message string `json:"message"`
							}{
								Message: "Welcome to Imager!! 🥳",
							}
							_ = json.NewEncoder(w).Encode(msg)
						case http.MethodPost:
							var payload input
							_ = json.NewDecoder(r.Body).Decode(&payload)

							if payload.ChartURL == "" {
								w.WriteHeader(http.StatusBadRequest)
								msg := struct {
									Message string `json:"message"`
								}{
									Message: "Expected a link to a helm chart",
								}
								_ = json.NewEncoder(w).Encode(msg)
								return
							}

							images, err := i.GetChartImagesDetails(r.Context(), payload.ChartURL)
							if err != nil {
								w.WriteHeader(http.StatusInternalServerError)
								msg := struct {
									Message string `json:"message"`
								}{
									Message: err.Error(),
								}
								_ = json.NewEncoder(w).Encode(msg)
								return
							}

							w.WriteHeader(http.StatusOK)
							_ = json.NewEncoder(w).Encode(images)
						default:
							w.WriteHeader(http.StatusBadRequest)
							msg := struct {
								Message string `json:"message"`
							}{
								Message: "Invalid HTTP method.",
							}
							_ = json.NewEncoder(w).Encode(msg)
						}

					})

					port := cCtx.String("port")
					log.Printf("Starting server on port %s...\n", port)

					if err := http.ListenAndServe(":"+port, nil); err != nil {
						log.Fatalf("Failed to start server: %v", err)
					}

					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
