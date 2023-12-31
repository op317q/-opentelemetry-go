// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	api "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/aggregation"
)

const meterName = "github.com/open-telemetry/opentelemetry-go/example/view"

func main() {
	ctx := context.Background()

	// // The exporter embeds a default OpenTelemetry Reader, allowing it to be used in WithReader.
	// exporter, err := otelprom.New()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// Print with a JSON encoder that indents with two spaces.
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	exporter, err := stdoutmetric.New(
		stdoutmetric.WithEncoder(enc),
		stdoutmetric.WithoutTimestamps(),
	)
	if err != nil {
		fmt.Print("failed to initialize stdouttrace exporter", err)
	}

	// exporter, err := otlpmetricgrpc.New(
	// 	ctx,
	// 	otlpmetricgrpc.WithTimeout(10*time.Second),
	// 	otlpmetricgrpc.WithInsecure(),
	// 	otlpmetricgrpc.WithEndpoint(""),
	// )
	// if err != nil {
	// 	fmt.Errorf("failed to initialize otlpmetricgrpc exporter", err)
	// }

	provider := metric.NewMeterProvider(
		metric.WithReader(
			metric.NewPeriodicReader(
				exporter,
				metric.WithInterval(10*time.Second),
			),
		),
		metric.WithView(
			metric.NewView(
				metric.Instrument{
					Name:  "custom_histogram",
					Scope: instrumentation.Scope{Name: meterName},
				},
				metric.Stream{
					Name: "bar",
					Aggregation: aggregation.ExplicitBucketHistogram{
						Boundaries: []float64{64, 128, 256, 512, 1024, 2048, 4096},
					},
				},
			),
			metric.NewView(
				metric.Instrument{
					Name:  "custom_histogram1",
					Scope: instrumentation.Scope{Name: meterName},
				},
				metric.Stream{
					Name: "joe",
					Aggregation: aggregation.ExplicitBucketHistogram{
						Boundaries: []float64{5, 10, 15, 20, 25, 30, 35},
					},
				},
			),
			metric.NewView(
				metric.Instrument{
					Name:  "custom_histogram2",
					Scope: instrumentation.Scope{Name: meterName},
				},
				metric.Stream{
					Name: "tony",
					Aggregation: aggregation.ExplicitBucketHistogram{
						Boundaries: []float64{4, 8, 12, 16, 24, 28, 32},
					},
				},
			),
		),
	)

	// provider = metric.NewMeterProvider(
	// 	metric.WithReader(exporter),
	// 	// View to customize histogram buckets and rename a single histogram instrument.
	// 	metric.WithView(metric.NewView(
	// 		metric.Instrument{
	// 			Name:  "custom_histogram",
	// 			Scope: instrumentation.Scope{Name: meterName},
	// 		},
	// 		metric.Stream{
	// 			Name: "bar",
	// 			Aggregation: aggregation.ExplicitBucketHistogram{
	// 				Boundaries: []float64{64, 128, 256, 512, 1024, 2048, 4096},
	// 			},
	// 		},
	// 	)),
	// )
	meter := provider.Meter(meterName)

	// Start the prometheus HTTP server and pass the exporter Collector to it
	go serveMetrics()

	opt := api.WithAttributes(
		attribute.Key("A").String("B"),
		attribute.Key("C").String("D"),
	)

	// counter, err := meter.Float64Counter("foo", api.WithDescription("a simple counter"))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// counter.Add(ctx, 5, opt)

	histogram, err := meter.Float64Histogram("custom_histogram", api.WithDescription("a histogram with custom buckets and rename"))
	if err != nil {
		log.Fatal(err)
	}
	histogram.Record(ctx, 136, opt)
	histogram.Record(ctx, 64, opt)
	histogram.Record(ctx, 701, opt)
	histogram.Record(ctx, 830, opt)

	// counter1, err := meter.Float64Counter("goo", api.WithDescription("a simple counter"))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// counter1.Add(ctx, 1, opt)

	histogram1, err := meter.Float64Histogram("custom_histogram1", api.WithDescription("a histogram with custom buckets and rename"))
	if err != nil {
		log.Fatal(err)
	}
	histogram1.Record(ctx, 4, opt)
	histogram1.Record(ctx, 8, opt)
	histogram1.Record(ctx, 12, opt)
	histogram1.Record(ctx, 16, opt)

	histogram2, err := meter.Float64Histogram("custom_histogram2", api.WithDescription("a histogram with custom buckets and rename"))
	if err != nil {
		log.Fatal(err)
	}
	histogram2.Record(ctx, 4, opt)
	histogram2.Record(ctx, 8, opt)
	histogram2.Record(ctx, 12, opt)
	histogram2.Record(ctx, 16, opt)

	ctx, _ = signal.NotifyContext(ctx, os.Interrupt)
	<-ctx.Done()
}

func serveMetrics() {
	log.Printf("serving metrics at localhost:2222/metrics")
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(":2222", nil)
	if err != nil {
		fmt.Printf("error serving http: %v", err)
		return
	}
}
