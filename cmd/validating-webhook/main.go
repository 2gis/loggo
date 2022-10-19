package main

import (
	"context"
	"log"
	"net/http"
	"os"

	corev1 "k8s.io/api/core/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/2gis/loggo/components/k8s"
	"github.com/2gis/loggo/configuration"
	"github.com/2gis/loggo/logging"
)

type validator struct {
	Client  client.Client
	Config  configuration.Config
	decoder *admission.Decoder
}

func (v *validator) Handle(ctx context.Context, req admission.Request) admission.Response {
	service := &corev1.Service{}

	err := v.decoder.Decode(req, service)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if _, err := k8s.CreateService(v.Config.SLIExporterConfig, service.Annotations); err != nil {
		return admission.Denied(err.Error())
	}

	return admission.Allowed("")
}

func (v *validator) InjectDecoder(d *admission.Decoder) error {
	v.decoder = d
	return nil
}

func main() {
	c := configuration.GetConfig()
	log.Printf("Starting with configuration: %s", c.ToString())
	logger := logging.NewLogger("json", c.LogLevel, os.Stdout)

	logger.Printf("Setting up controller manager")
	restconfig, err := config.GetConfig()
	if err != nil {
		logger.Fatalln(err)
	}
	mgr, err := manager.New(restconfig, manager.Options{
		HealthProbeBindAddress: ":8090",
		MetricsBindAddress:     ":8080",
		Port:                   9443,
	})
	if err != nil {
		logger.Fatalln(err)
	}

	logger.Printf("Registering healthz and readyz checkers")
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		logger.Fatalln(err)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		logger.Fatalln(err)
	}

	logger.Printf("Setting up webhook server")
	hookServer := mgr.GetWebhookServer()

	logger.Printf("Registering validating-webhook to the webhook server")
	hookServer.Register("/validate", &webhook.Admission{
		Handler: &validator{
			Client: mgr.GetClient(),
			Config: c,
		},
	})

	logger.Printf("Starting manager")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		logger.Fatalln(err)
	}
}
