package emma

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"regexp"
	"strconv"
	"strings"
)

const (
	Slash          = "/"
	AnyPort        = "all"
	AnyIp          = "0.0.0.0"
	Ipv4Regex      = `^(\d{1,3})\.(\d{1,3})\.(\d{1,3})\.(\d{1,3})$`
	Ipv4RangeRegex = `^(\d{1,3})\.(\d{1,3})\.(\d{1,3})\.(\d{1,3})/(\d{1,2})$`
)

var (
	Ipv4Pattern      = regexp.MustCompile(Ipv4Regex)
	Ipv4RangePattern = regexp.MustCompile(Ipv4RangeRegex)
)

type PortRange struct {
}

func (v PortRange) Description(ctx context.Context) string {
	return "ports is invalid"
}

func (v PortRange) MarkdownDescription(ctx context.Context) string {
	return "ports is invalid"
}

func (v PortRange) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}
	if !isValidPortsValue(req.ConfigValue.ValueString()) {
		resp.Diagnostics.AddError("Validation Error", req.Path.String()+" is invalid, may contain next values: all, 3000 or 1-3010")
	}
}

type IpRange struct {
}

func (v IpRange) Description(ctx context.Context) string {
	return "ip_range is invalid"
}

func (v IpRange) MarkdownDescription(ctx context.Context) string {
	return "ip_range is invalid"
}

func (v IpRange) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}
	if !isValidIpRangeValue(req.ConfigValue.ValueString()) {
		resp.Diagnostics.AddError("Validation Error", req.Path.String()+" is invalid, may contain next values: 0.0.0.0, 1.1.1.1 or 1.1.1.1/32")
	}
}

type Direction struct {
}

func (v Direction) Description(ctx context.Context) string {
	return "direction can contain INBOUND or OUTBOUND"
}

func (v Direction) MarkdownDescription(ctx context.Context) string {
	return "direction can contain INBOUND or OUTBOUND"
}

func (v Direction) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}
	if req.ConfigValue.ValueString() != "INBOUND" && req.ConfigValue.ValueString() != "OUTBOUND" {
		resp.Diagnostics.AddError("Validation Error", req.Path.String()+" can contain INBOUND or OUTBOUND")
	}
}

type Protocol struct {
}

func (v Protocol) Description(ctx context.Context) string {
	return "protocol can contain next values: all, TCP, SCTP, GRE, ESP, AH, UDP and ICMP"
}

func (v Protocol) MarkdownDescription(ctx context.Context) string {
	return "protocol can contain next values: all, TCP, SCTP, GRE, ESP, AH, UDP and ICMP"
}

func (v Protocol) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}
	if req.ConfigValue.ValueString() != "all" && req.ConfigValue.ValueString() != "TCP" &&
		req.ConfigValue.ValueString() != "SCTP" && req.ConfigValue.ValueString() != "GRE" &&
		req.ConfigValue.ValueString() != "ESP" && req.ConfigValue.ValueString() != "AH" &&
		req.ConfigValue.ValueString() != "UDP" && req.ConfigValue.ValueString() != "ICMP" {
		resp.Diagnostics.AddError("Validation Error", req.Path.String()+" can contain next values: all, TCP, SCTP, GRE, ESP, AH, UDP and ICMP")
	}
}

func isValidPortsValue(ports string) bool {
	if ports == AnyPort {
		return true
	}
	if strings.Contains(ports, "-") {
		parts := strings.Split(ports, "-")
		if !isValidSinglePortValue(parts[0]) || !isValidSinglePortValue(parts[1]) {
			return false
		}

		start, _ := strconv.Atoi(parts[0])
		end, _ := strconv.Atoi(parts[1])
		if end < start {
			return false
		}
	}
	return isValidSinglePortValue(ports)
}

func isValidSinglePortValue(port string) bool {
	portValue, err := strconv.Atoi(port)
	if err != nil {
		return false
	}
	if portValue < 0 || portValue > 65535 {
		return false
	}
	return true
}

func isValidIpRangeValue(ipRange string) bool {
	if !isValidIpRange(ipRange) && !isValidIp(ipRange) {
		return false
	}
	if strings.HasPrefix(ipRange, AnyIp) && strings.Contains(ipRange, Slash) {
		splitted := strings.Split(ipRange, Slash)
		if len(splitted) < 2 {
			return false
		}
		_, err := strconv.Atoi(splitted[1])
		if err != nil {
			return false
		}
	}
	return true
}

func isValidIp(ip string) bool {
	if !Ipv4Pattern.MatchString(ip) {
		return false
	}
	octets := strings.Split(ip, ".")
	for _, octetStr := range octets {
		octet, err := strconv.Atoi(octetStr)
		if err != nil {
			return false
		}
		if !isValidOctetValue(octet) {
			return false
		}
	}
	return true
}

func isValidIpRange(ipRange string) bool {
	if !Ipv4RangePattern.MatchString(ipRange) {
		return false
	}
	ip := strings.Split(ipRange, "/")[0]
	if !isValidIp(ip) {
		return false
	}
	prefix, err := strconv.Atoi(strings.Split(ipRange, "/")[1])
	if err != nil {
		return false
	}
	if !isValidPrefixLength(prefix) {
		return false
	}
	return true
}

func isValidOctetValue(octetValue int) bool {
	if octetValue < 0 || octetValue > 255 {
		return false
	}
	return true
}

func isValidPrefixLength(prefixLength int) bool {
	if prefixLength < 1 || prefixLength > 32 {
		return false
	}
	return true
}
