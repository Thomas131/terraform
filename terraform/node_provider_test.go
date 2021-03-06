package terraform

import (
	"testing"

	"github.com/hashicorp/terraform/addrs"
	"github.com/hashicorp/terraform/configs"
	"github.com/zclconf/go-cty/cty"
)

func TestNodeApplyableProviderExecute(t *testing.T) {
	config := &configs.Provider{
		Name: "foo",
		Config: configs.SynthBody("", map[string]cty.Value{
			"test_string": cty.StringVal("hello"),
		}),
	}
	provider := mockProviderWithConfigSchema(simpleTestSchema())
	providerAddr := addrs.AbsProviderConfig{
		Module:   addrs.RootModule,
		Provider: addrs.NewDefaultProvider("foo"),
	}

	n := &NodeApplyableProvider{&NodeAbstractProvider{
		Addr:   providerAddr,
		Config: config,
	}}

	ctx := &MockEvalContext{ProviderProvider: provider}
	ctx.installSimpleEval()
	if err := n.Execute(ctx, walkApply); err != nil {
		t.Fatalf("err: %s", err)
	}

	if !ctx.ConfigureProviderCalled {
		t.Fatal("should be called")
	}

	gotObj := ctx.ConfigureProviderConfig
	if !gotObj.Type().HasAttribute("test_string") {
		t.Fatal("configuration object does not have \"test_string\" attribute")
	}
	if got, want := gotObj.GetAttr("test_string"), cty.StringVal("hello"); !got.RawEquals(want) {
		t.Errorf("wrong configuration value\ngot:  %#v\nwant: %#v", got, want)
	}
}

func TestNodeApplyableProviderExecute_unknownImport(t *testing.T) {
	config := &configs.Provider{
		Name: "foo",
		Config: configs.SynthBody("", map[string]cty.Value{
			"test_string": cty.UnknownVal(cty.String),
		}),
	}
	provider := mockProviderWithConfigSchema(simpleTestSchema())
	providerAddr := addrs.AbsProviderConfig{
		Module:   addrs.RootModule,
		Provider: addrs.NewDefaultProvider("foo"),
	}
	n := &NodeApplyableProvider{&NodeAbstractProvider{
		Addr:   providerAddr,
		Config: config,
	}}

	ctx := &MockEvalContext{ProviderProvider: provider}
	ctx.installSimpleEval()

	err := n.Execute(ctx, walkImport)
	if err == nil {
		t.Fatal("expected error, got success")
	}

	detail := `Invalid provider configuration: The configuration for provider["registry.terraform.io/hashicorp/foo"] depends on values that cannot be determined until apply.`
	if got, want := err.Error(), detail; got != want {
		t.Errorf("wrong diagnostic detail\n got: %q\nwant: %q", got, want)
	}

	if ctx.ConfigureProviderCalled {
		t.Fatal("should not be called")
	}
}

func TestNodeApplyableProviderExecute_unknownApply(t *testing.T) {
	config := &configs.Provider{
		Name: "foo",
		Config: configs.SynthBody("", map[string]cty.Value{
			"test_string": cty.UnknownVal(cty.String),
		}),
	}
	provider := mockProviderWithConfigSchema(simpleTestSchema())
	providerAddr := addrs.AbsProviderConfig{
		Module:   addrs.RootModule,
		Provider: addrs.NewDefaultProvider("foo"),
	}
	n := &NodeApplyableProvider{&NodeAbstractProvider{
		Addr:   providerAddr,
		Config: config,
	}}
	ctx := &MockEvalContext{ProviderProvider: provider}
	ctx.installSimpleEval()

	if err := n.Execute(ctx, walkApply); err != nil {
		t.Fatalf("err: %s", err)
	}

	if !ctx.ConfigureProviderCalled {
		t.Fatal("should be called")
	}

	gotObj := ctx.ConfigureProviderConfig
	if !gotObj.Type().HasAttribute("test_string") {
		t.Fatal("configuration object does not have \"test_string\" attribute")
	}
	if got, want := gotObj.GetAttr("test_string"), cty.UnknownVal(cty.String); !got.RawEquals(want) {
		t.Errorf("wrong configuration value\ngot:  %#v\nwant: %#v", got, want)
	}
}

func TestNodeApplyableProviderExecute_sensitive(t *testing.T) {
	config := &configs.Provider{
		Name: "foo",
		Config: configs.SynthBody("", map[string]cty.Value{
			"test_string": cty.StringVal("hello").Mark("sensitive"),
		}),
	}
	provider := mockProviderWithConfigSchema(simpleTestSchema())
	providerAddr := addrs.AbsProviderConfig{
		Module:   addrs.RootModule,
		Provider: addrs.NewDefaultProvider("foo"),
	}

	n := &NodeApplyableProvider{&NodeAbstractProvider{
		Addr:   providerAddr,
		Config: config,
	}}

	ctx := &MockEvalContext{ProviderProvider: provider}
	ctx.installSimpleEval()
	if err := n.Execute(ctx, walkApply); err != nil {
		t.Fatalf("err: %s", err)
	}

	if !ctx.ConfigureProviderCalled {
		t.Fatal("should be called")
	}

	gotObj := ctx.ConfigureProviderConfig
	if !gotObj.Type().HasAttribute("test_string") {
		t.Fatal("configuration object does not have \"test_string\" attribute")
	}
	if got, want := gotObj.GetAttr("test_string"), cty.StringVal("hello"); !got.RawEquals(want) {
		t.Errorf("wrong configuration value\ngot:  %#v\nwant: %#v", got, want)
	}
}

func TestNodeApplyableProviderExecute_sensitiveValidate(t *testing.T) {
	config := &configs.Provider{
		Name: "foo",
		Config: configs.SynthBody("", map[string]cty.Value{
			"test_string": cty.StringVal("hello").Mark("sensitive"),
		}),
	}
	provider := mockProviderWithConfigSchema(simpleTestSchema())
	providerAddr := addrs.AbsProviderConfig{
		Module:   addrs.RootModule,
		Provider: addrs.NewDefaultProvider("foo"),
	}

	n := &NodeApplyableProvider{&NodeAbstractProvider{
		Addr:   providerAddr,
		Config: config,
	}}

	ctx := &MockEvalContext{ProviderProvider: provider}
	ctx.installSimpleEval()
	if err := n.Execute(ctx, walkValidate); err != nil {
		t.Fatalf("err: %s", err)
	}

	if !provider.PrepareProviderConfigCalled {
		t.Fatal("should be called")
	}

	gotObj := provider.PrepareProviderConfigRequest.Config
	if !gotObj.Type().HasAttribute("test_string") {
		t.Fatal("configuration object does not have \"test_string\" attribute")
	}
	if got, want := gotObj.GetAttr("test_string"), cty.StringVal("hello"); !got.RawEquals(want) {
		t.Errorf("wrong configuration value\ngot:  %#v\nwant: %#v", got, want)
	}
}
