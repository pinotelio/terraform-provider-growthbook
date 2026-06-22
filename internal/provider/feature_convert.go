package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pinotelio/terraform-provider-growthbook/internal/client"
)

// buildEnvironments converts the Terraform environment map into the API input
// shape. A nil/empty map yields nil so the field is omitted from the request.
func buildEnvironments(ctx context.Context, envs map[string]featureEnvModel, diags *diag.Diagnostics) map[string]client.FeatureEnvironmentInput {
	if len(envs) == 0 {
		return nil
	}
	out := make(map[string]client.FeatureEnvironmentInput, len(envs))
	for name, env := range envs {
		// Start from a non-nil slice so an environment with no rules marshals as
		// `[]` (which GrowthBook requires) instead of being omitted.
		in := client.FeatureEnvironmentInput{
			Enabled: optBool(env.Enabled),
			Rules:   make([]client.FeatureRule, 0, len(env.Rules)),
		}
		for _, rule := range env.Rules {
			in.Rules = append(in.Rules, buildRule(ctx, rule, diags))
		}
		out[name] = in
	}
	return out
}

func buildRule(ctx context.Context, m featureRuleModel, diags *diag.Diagnostics) client.FeatureRule {
	rule := client.FeatureRule{
		Type:                   m.Type.ValueString(),
		Description:            optString(m.Description),
		Enabled:                optBool(m.Enabled),
		Condition:              optString(m.Condition),
		Value:                  optString(m.Value),
		Coverage:               optFloat64(m.Coverage),
		HashAttribute:          optString(m.HashAttribute),
		Seed:                   optString(m.Seed),
		HashVersion:            optInt64(m.HashVersion),
		TrackingKey:            optString(m.TrackingKey),
		FallbackAttribute:      optString(m.FallbackAttribute),
		DisableStickyBucketing: optBool(m.DisableStickyBucketing),
		BucketVersion:          optInt64(m.BucketVersion),
		MinBucketVersion:       optInt64(m.MinBucketVersion),
		ExperimentID:           optString(m.ExperimentID),
		Sparse:                 optBool(m.Sparse),
	}
	for _, sg := range m.SavedGroupTargeting {
		rule.SavedGroupTargeting = append(rule.SavedGroupTargeting, client.SavedGroupTargeting{
			MatchType:   sg.MatchType.ValueString(),
			SavedGroups: stringListToSlice(ctx, sg.SavedGroups, diags),
		})
	}
	for _, pr := range m.Prerequisites {
		rule.Prerequisites = append(rule.Prerequisites, client.RulePrerequisite{
			ID:        pr.ID.ValueString(),
			Condition: pr.Condition.ValueString(),
		})
	}
	for _, sr := range m.ScheduleRules {
		rule.ScheduleRules = append(rule.ScheduleRules, client.ScheduleRule{
			Enabled:   sr.Enabled.ValueBool(),
			Timestamp: optString(sr.Timestamp),
		})
	}
	if m.Namespace != nil {
		rule.Namespace = &client.RuleNamespace{
			Enabled: m.Namespace.Enabled.ValueBool(),
			Name:    m.Namespace.Name.ValueString(),
			Range:   []float64{m.Namespace.RangeMin.ValueFloat64(), m.Namespace.RangeMax.ValueFloat64()},
		}
	}
	for _, v := range m.Values {
		rule.Values = append(rule.Values, client.ExperimentValue{
			Value:  v.Value.ValueString(),
			Weight: v.Weight.ValueFloat64(),
			Name:   optString(v.Name),
		})
	}
	for _, v := range m.Variations {
		rule.Variations = append(rule.Variations, client.ExperimentRefVariation{
			Value:       v.Value.ValueString(),
			VariationID: v.VariationID.ValueString(),
		})
	}
	return rule
}

// flattenEnvironments rebuilds the Terraform environment map from an API
// feature. It is used to seed state on import; during normal Read the
// configured environments are preserved instead (see featureResource.Read).
func flattenEnvironments(f *client.Feature) map[string]featureEnvModel {
	if len(f.Environments) == 0 {
		return nil
	}
	out := make(map[string]featureEnvModel)
	for name, env := range f.Environments {
		model := featureEnvModel{Enabled: types.BoolValue(env.Enabled)}
		for _, rule := range env.Rules {
			model.Rules = append(model.Rules, flattenRule(rule))
		}
		out[name] = model
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func flattenRule(r client.FeatureRule) featureRuleModel {
	m := featureRuleModel{
		Type:                   types.StringValue(r.Type),
		Description:            stringPtrValue(r.Description),
		Enabled:                boolPtrValue(r.Enabled),
		Condition:              stringPtrValue(r.Condition),
		Value:                  stringPtrValue(r.Value),
		Coverage:               float64PtrValue(r.Coverage),
		HashAttribute:          stringPtrValue(r.HashAttribute),
		Seed:                   stringPtrValue(r.Seed),
		HashVersion:            int64PtrValue(r.HashVersion),
		TrackingKey:            stringPtrValue(r.TrackingKey),
		FallbackAttribute:      stringPtrValue(r.FallbackAttribute),
		DisableStickyBucketing: boolPtrValue(r.DisableStickyBucketing),
		BucketVersion:          int64PtrValue(r.BucketVersion),
		MinBucketVersion:       int64PtrValue(r.MinBucketVersion),
		ExperimentID:           stringPtrValue(r.ExperimentID),
		Sparse:                 boolPtrValue(r.Sparse),
	}
	for _, sg := range r.SavedGroupTargeting {
		m.SavedGroupTargeting = append(m.SavedGroupTargeting, savedGroupTargetingModel{
			MatchType:   types.StringValue(sg.MatchType),
			SavedGroups: sliceToStringList(sg.SavedGroups),
		})
	}
	for _, pr := range r.Prerequisites {
		m.Prerequisites = append(m.Prerequisites, rulePrereqModel{
			ID:        types.StringValue(pr.ID),
			Condition: types.StringValue(pr.Condition),
		})
	}
	for _, sr := range r.ScheduleRules {
		m.ScheduleRules = append(m.ScheduleRules, scheduleRuleModel{
			Enabled:   types.BoolValue(sr.Enabled),
			Timestamp: stringPtrValue(sr.Timestamp),
		})
	}
	if r.Namespace != nil {
		ns := &ruleNamespaceModel{
			Enabled: types.BoolValue(r.Namespace.Enabled),
			Name:    types.StringValue(r.Namespace.Name),
		}
		if len(r.Namespace.Range) == 2 {
			ns.RangeMin = types.Float64Value(r.Namespace.Range[0])
			ns.RangeMax = types.Float64Value(r.Namespace.Range[1])
		}
		m.Namespace = ns
	}
	for _, v := range r.Values {
		m.Values = append(m.Values, experimentValueModel{
			Value:  types.StringValue(v.Value),
			Weight: types.Float64Value(v.Weight),
			Name:   stringPtrValue(v.Name),
		})
	}
	for _, v := range r.Variations {
		m.Variations = append(m.Variations, experimentRefVariationModel{
			Value:       types.StringValue(v.Value),
			VariationID: types.StringValue(v.VariationID),
		})
	}
	return m
}
