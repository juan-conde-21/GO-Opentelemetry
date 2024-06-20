# GO-Opentelemetry
GO-Opentelemetry

Instalar librerias de Opentelemetry

  Trazas
  
    go get go.opentelemetry.io/otel
    go get go.opentelemetry.io/otel/sdk
    go get go.opentelemetry.io/otel/sdk/trace
    go get go.opentelemetry.io/otel/exporters/stdout/stdouttrace
    go get go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp
  
  Metricas
  
    go get go.opentelemetry.io/otel/metric
    go get go.opentelemetry.io/otel/sdk/metric
    go get go.opentelemetry.io/otel/exporters/metric/stdout/stdoutmetric
    go get go.opentelemetry.io/otel/exporters/stdout/stdoutmetric
  
  Exporters
  
    go get go.opentelemetry.io/otel/exporters/otlp/otlptrace
    go get go.opentelemetry.io/otel/exporters/otlp/otlpmetric
    go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc
    go get go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc

