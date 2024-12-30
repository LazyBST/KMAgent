package main

import (
	"context"
	"log"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/provider/envprovider"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"
	"go.opentelemetry.io/collector/confmap/provider/httpprovider"
	"go.opentelemetry.io/collector/confmap/provider/httpsprovider"
	"go.opentelemetry.io/collector/confmap/provider/yamlprovider"
	"go.opentelemetry.io/collector/otelcol"
)

func startCollector() {
	if collector != nil {
		collector.Shutdown()
		collector = nil
		log.Println("old collector instance stopped")
	}

	info := component.BuildInfo{
		Command:     "kmagent",
		Description: "Agent-as-collector POC",
		Version:     "0.0.1",
	}

	set := otelcol.CollectorSettings{
		BuildInfo: info,
		Factories: components,
		ConfigProviderSettings: otelcol.ConfigProviderSettings{
			ResolverSettings: confmap.ResolverSettings{
				URIs: []string{CONFIG_PATH},
				ProviderFactories: []confmap.ProviderFactory{
					envprovider.NewFactory(),
					fileprovider.NewFactory(),
					httpprovider.NewFactory(),
					httpsprovider.NewFactory(),
					yamlprovider.NewFactory(),
				},
			},
		},
	}

	runInteractive(set)
	log.Println("collector started")
}

func runInteractive(params otelcol.CollectorSettings) {
	go func(params otelcol.CollectorSettings) {
		var err error
		collector, err = otelcol.NewCollector(params)

		if err != nil {
			log.Fatalf("collector instance creation failed with error: %v", err)
		}

		err = collector.Run(context.Background())

		if err != nil {
			log.Fatalf("collector server run finished with error: %v", err)
		}
	}(params)
}

func stopCollector() {
	if collector == nil {
		return
	}

	collector.Shutdown()
}
