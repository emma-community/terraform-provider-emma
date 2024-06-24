package emma

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

var invalidPortValues = []struct {
	in string
}{
	{"-1"},
	{"65536"},
	{"123A"},
	{"A"},
	{"123-a"},
	{"a-123"},
	{"-1-123"},
	{"1--123"},
	{"-1--123"},
	{"3008-300"},
}

var validPortValues = []struct {
	in string
}{
	{"0"},
	{"65535"},
	{"3000-3008"},
	{"all"},
	{"3000"},
}

var invalidIpValues = []struct {
	in string
}{
	{"256.1.1.1"},
	{"1.256.1.1"},
	{"1.1.256.1"},
	{"1.1.1.256"},
	{"1,1.1.1"},
	{"-1.1.1.1"},
	{"1.-1.1.1"},
	{"1.1.-1.1"},
	{"1.1.1.-1"},
	{"1.1.1.1/"},
	{"1.1.1.1/0"},
	{"1.1.1.1/33"},
	{"1.1.1.1/-1"},
	{"1.1.1.1\\32"},
	{"a"},
}

var validIpValues = []struct {
	in string
}{
	{"1.1.1.1"},
	{"0.0.0.0"},
	{"1.1.1.1/32"},
	{"1.1.1.1/1"},
	{"255.255.255.255/32"},
	{"255.255.255.255/1"},
	{"255.255.255.255"},
}

func TestPortRange_ValidateString_InvalidValues(t *testing.T) {
	for _, invalidPortValue := range invalidPortValues {
		v := PortRange{}
		var resp validator.StringResponse
		var req validator.StringRequest

		req.ConfigValue = types.StringValue(invalidPortValue.in)
		req.Path = path.Root("test")

		v.ValidateString(context.Background(), req, &resp)

		assert.Equal(t, 1, resp.Diagnostics.ErrorsCount())
		if resp.Diagnostics.HasError() {
			actualMsg := resp.Diagnostics.Errors()[0].Detail()
			assert.Equal(t, "test is invalid, may contain next values: all, 3000 or 1-3010", actualMsg)
		} else {
			assert.Fail(t, "Is valid port value: "+invalidPortValue.in)
		}
	}
}

func TestPortRange_ValidateString_ValidValues(t *testing.T) {
	for _, validPortValue := range validPortValues {
		v := PortRange{}
		var resp validator.StringResponse
		var req validator.StringRequest

		req.ConfigValue = types.StringValue(validPortValue.in)
		req.Path = path.Root("test")

		v.ValidateString(context.Background(), req, &resp)

		assert.False(t, resp.Diagnostics.HasError(), "Is invalid port value: "+validPortValue.in)
	}
}

func TestIpRange_ValidateString_InvalidValues(t *testing.T) {
	for _, invalidIpValue := range invalidIpValues {
		v := IpRange{}
		var resp validator.StringResponse
		var req validator.StringRequest

		req.ConfigValue = types.StringValue(invalidIpValue.in)
		req.Path = path.Root("test")

		v.ValidateString(context.Background(), req, &resp)

		assert.Equal(t, 1, resp.Diagnostics.ErrorsCount())
		if resp.Diagnostics.HasError() {
			actualMsg := resp.Diagnostics.Errors()[0].Detail()
			assert.Equal(t, "test is invalid, may contain next values: 0.0.0.0, 1.1.1.1 or 1.1.1.1/32", actualMsg)
		} else {
			assert.Fail(t, "Is valid ip value: "+invalidIpValue.in)
		}
	}
}

func TestIpRange_ValidateString_ValidValues(t *testing.T) {
	for _, validIpValue := range validIpValues {
		v := IpRange{}
		var resp validator.StringResponse
		var req validator.StringRequest

		req.ConfigValue = types.StringValue(validIpValue.in)
		req.Path = path.Root("test")

		v.ValidateString(context.Background(), req, &resp)

		assert.False(t, resp.Diagnostics.HasError(), "Is invalid ip value: "+validIpValue.in)
	}
}

func TestDirection_ValidateString_InvalidValue(t *testing.T) {
	v := Direction{}
	var resp validator.StringResponse
	var req validator.StringRequest

	req.ConfigValue = types.StringValue("test")
	req.Path = path.Root("test")

	v.ValidateString(context.Background(), req, &resp)

	assert.Equal(t, 1, resp.Diagnostics.ErrorsCount())
	if resp.Diagnostics.HasError() {
		actualMsg := resp.Diagnostics.Errors()[0].Detail()
		assert.Equal(t, "test can contain INBOUND or OUTBOUND", actualMsg)
	} else {
		assert.Fail(t, "Is valid direction value: test")
	}
}

func TestDirection_ValidateString_ValidValues(t *testing.T) {
	for _, validDirectionValue := range []string{"INBOUND", "OUTBOUND"} {
		v := Direction{}
		var resp validator.StringResponse
		var req validator.StringRequest

		req.ConfigValue = types.StringValue(validDirectionValue)
		req.Path = path.Root("test")

		v.ValidateString(context.Background(), req, &resp)

		assert.False(t, resp.Diagnostics.HasError(), "Is invalid direction value: "+validDirectionValue)
	}
}

func TestProtocol_ValidateString_InvalidValue(t *testing.T) {
	v := Protocol{}
	var resp validator.StringResponse
	var req validator.StringRequest

	req.ConfigValue = types.StringValue("test")
	req.Path = path.Root("test")

	v.ValidateString(context.Background(), req, &resp)

	assert.Equal(t, 1, resp.Diagnostics.ErrorsCount())
	if resp.Diagnostics.HasError() {
		actualMsg := resp.Diagnostics.Errors()[0].Detail()
		assert.Equal(t, "test can contain next values: all, TCP, SCTP, GRE, ESP, AH, UDP and ICMP", actualMsg)
	} else {
		assert.Fail(t, "Is valid protocol value: test")
	}
}

func TestProtocol_ValidateString_ValidValues(t *testing.T) {
	for _, validDirectionValue := range []string{"all", "TCP", "SCTP", "GRE", "ESP", "AH", "UDP", "ICMP"} {
		v := Protocol{}
		var resp validator.StringResponse
		var req validator.StringRequest

		req.ConfigValue = types.StringValue(validDirectionValue)
		req.Path = path.Root("test")

		v.ValidateString(context.Background(), req, &resp)

		assert.False(t, resp.Diagnostics.HasError(), "Is invalid protocol value: "+validDirectionValue)
	}
}
