package webhooksupportabilitycontroller

import (
	"context"

	operatorv1 "github.com/openshift/api/operator/v1"
	"github.com/openshift/library-go/pkg/operator/v1helpers"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	v1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func (c *webhookSupportabilityController) updateMutatingAdmissionWebhookConfigurationDegraded(ctx context.Context) v1helpers.UpdateStatusFunc {
	tracer := otel.GetTracerProvider().Tracer("ckao")
	ctx, span := tracer.Start(ctx, "webhook.updateMutatingAdmissionWebhookConfigurationDegraded", trace.WithAttributes(
		attribute.String("name", c.Name()),
	))
	defer span.End()

	condition := operatorv1.OperatorCondition{
		Type:   MutatingAdmissionWebhookConfigurationErrorType,
		Status: operatorv1.ConditionUnknown,
	}
	webhookConfigurations, err := c.mutatingWebhookLister.List(ctx, labels.Everything())
	if err != nil {
		condition.Message = err.Error()
		return v1helpers.UpdateConditionFn(condition)
	}
	var webhookInfos []webhookInfo
	for _, webhookConfiguration := range webhookConfigurations {
		for _, webhook := range webhookConfiguration.Webhooks {
			info := webhookInfo{
				Name:                   webhook.Name,
				CABundle:               webhook.ClientConfig.CABundle,
				HasServiceCaAnnotation: hasServiceCaAnnotation(webhookConfiguration.Annotations),
				FailurePolicyIsIgnore:  webhook.FailurePolicy != nil && *webhook.FailurePolicy == admissionregistrationv1.Ignore,
				TimeoutSeconds:         webhook.TimeoutSeconds,
			}
			if webhook.ClientConfig.Service != nil {
				info.Service = &serviceReference{
					Namespace: webhook.ClientConfig.Service.Namespace,
					Name:      webhook.ClientConfig.Service.Name,
					Port:      webhook.ClientConfig.Service.Port,
				}
			}
			webhookInfos = append(webhookInfos, info)
		}
	}
	return c.updateWebhookConfigurationDegraded(ctx, condition, webhookInfos)
}

func (c *webhookSupportabilityController) updateValidatingAdmissionWebhookConfigurationDegradedStatus(ctx context.Context) v1helpers.UpdateStatusFunc {
	tracer := otel.GetTracerProvider().Tracer("ckao")
	ctx, span := tracer.Start(ctx, "webhook.updateValidatingAdmissionWebhookConfigurationDegradedStatus", trace.WithAttributes(
		attribute.String("name", c.Name()),
	))
	defer span.End()
	condition := operatorv1.OperatorCondition{
		Type:   ValidatingAdmissionWebhookConfigurationErrorType,
		Status: operatorv1.ConditionUnknown,
	}
	webhookConfigurations, err := c.validatingWebhookLister.List(ctx, labels.Everything())
	if err != nil {
		condition.Message = err.Error()
		return v1helpers.UpdateConditionFn(condition)
	}
	var webhookInfos []webhookInfo
	for _, webhookConfiguration := range webhookConfigurations {
		for _, webhook := range webhookConfiguration.Webhooks {
			info := webhookInfo{
				Name:                   webhook.Name,
				CABundle:               webhook.ClientConfig.CABundle,
				HasServiceCaAnnotation: hasServiceCaAnnotation(webhookConfiguration.Annotations),
				FailurePolicyIsIgnore:  webhook.FailurePolicy != nil && (*webhook.FailurePolicy == v1.Ignore),
				TimeoutSeconds:         webhook.TimeoutSeconds,
			}

			if webhook.ClientConfig.Service != nil {
				info.Service = &serviceReference{
					Namespace: webhook.ClientConfig.Service.Namespace,
					Name:      webhook.ClientConfig.Service.Name,
					Port:      webhook.ClientConfig.Service.Port,
				}
			}
			webhookInfos = append(webhookInfos, info)
		}
	}
	return c.updateWebhookConfigurationDegraded(ctx, condition, webhookInfos)
}
