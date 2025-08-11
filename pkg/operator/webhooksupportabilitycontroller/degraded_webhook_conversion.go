package webhooksupportabilitycontroller

import (
	"context"

	operatorv1 "github.com/openshift/api/operator/v1"
	"github.com/openshift/library-go/pkg/operator/v1helpers"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func (c *webhookSupportabilityController) updateCRDConversionWebhookConfigurationDegraded(ctx context.Context) v1helpers.UpdateStatusFunc {
	tracer := otel.GetTracerProvider().Tracer("ckao")
	ctx, span := tracer.Start(ctx, "webhook.updateCRDConversionWebhookConfigurationDegraded", trace.WithAttributes(
		attribute.String("name", c.Name()),
	))
	defer span.End()

	condition := operatorv1.OperatorCondition{
		Type:   CRDConversionWebhookConfigurationErrorType,
		Status: operatorv1.ConditionUnknown,
	}
	crds, err := c.crdLister.List(ctx, labels.Everything())
	if err != nil {
		condition.Message = err.Error()
		return v1helpers.UpdateConditionFn(condition)
	}
	var webhookInfos []webhookInfo
	for _, crd := range crds {
		if !hasCRDConversionWebhookConfiguration(crd) {
			continue
		}
		info := webhookInfo{
			Name:                   crd.Name,
			CABundle:               crd.Spec.Conversion.Webhook.ClientConfig.CABundle,
			HasServiceCaAnnotation: hasServiceCaAnnotation(crd.Annotations),
			Service: &serviceReference{
				Namespace: crd.Spec.Conversion.Webhook.ClientConfig.Service.Namespace,
				Name:      crd.Spec.Conversion.Webhook.ClientConfig.Service.Name,
				Port:      crd.Spec.Conversion.Webhook.ClientConfig.Service.Port,
			},
		}
		webhookInfos = append(webhookInfos, info)
	}
	return c.updateWebhookConfigurationDegraded(ctx, condition, webhookInfos)
}

func hasCRDConversionWebhookConfiguration(crd *apiextensionsv1.CustomResourceDefinition) bool {
	conversion := crd.Spec.Conversion
	if conversion == nil || conversion.Strategy != v1.WebhookConverter {
		return false
	}
	clientConfig := conversion.Webhook.ClientConfig
	if clientConfig == nil || clientConfig.Service == nil {
		return false
	}
	return true
}
