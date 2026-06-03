package emma

// Deprecated: This data source uses the Statistics API endpoint POST /v1/statistics/retrieve,
// which was marked deprecated on 2026-05-13. It remains available for historical reports
// until a replacement endpoint is provided.

import (
	"context"
	"fmt"
	emmaSdk "github.com/emma-community/emma-go-sdk"
	"github.com/emma-community/terraform-provider-emma/tools"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &statisticalDataDataSource{}

func NewStatisticalDataDataSource() datasource.DataSource {
	return &statisticalDataDataSource{}
}

// statisticalDataDataSource defines the data source implementation.
type statisticalDataDataSource struct {
	apiClient *emmaSdk.APIClient
	token     *emmaSdk.Token
}

// ---------------------------------------------------------------------------
// Top-level data model — exactly one query block should be set.
// ---------------------------------------------------------------------------

type statisticalDataDataSourceModel struct {
	// Query variants (exactly one should be set)
	KubernetesClusterChangingMetrics *k8sChangingMetricsQueryModel `tfsdk:"kubernetes_cluster_changing_metrics"`
	KubernetesClusterCurrentState    *k8sCurrentStateQueryModel    `tfsdk:"kubernetes_cluster_current_state"`
	KubernetesClusterMetrics         *k8sMetricsQueryModel         `tfsdk:"kubernetes_cluster_metrics"`
	KubernetesClusterObjectStates    *k8sObjectStatesQueryModel    `tfsdk:"kubernetes_cluster_object_states"`
	KubernetesClusterObjects         *k8sObjectsQueryModel         `tfsdk:"kubernetes_cluster_objects"`
	ProductStatistics                *productStatisticsQueryModel  `tfsdk:"product_statistics"`
	ProjectSummary                   *projectSummaryQueryModel     `tfsdk:"project_summary"`
	ResourceAnalysis                 *resourceAnalysisQueryModel   `tfsdk:"resource_analysis"`
	VmAnalytics                      *vmAnalyticsQueryModel        `tfsdk:"vm_analytics"`
	VmMonitoring                     *vmMonitoringQueryModel       `tfsdk:"vm_monitoring"`

	// Computed result blocks — only the one matching the query variant will be populated.
	KubernetesClusterChangingMetricsResults types.List `tfsdk:"kubernetes_cluster_changing_metrics_results"`
	KubernetesClusterCurrentStateResults    types.List `tfsdk:"kubernetes_cluster_current_state_results"`
	KubernetesClusterMetricsResults         types.List `tfsdk:"kubernetes_cluster_metrics_results"`
	KubernetesClusterObjectStatesResults    types.List `tfsdk:"kubernetes_cluster_object_states_results"`
	KubernetesClusterObjectsResults         types.List `tfsdk:"kubernetes_cluster_objects_results"`
	ProductStatisticsResults                types.List `tfsdk:"product_statistics_results"`
	ProjectSummaryResults                   types.List `tfsdk:"project_summary_results"`
	ResourceAnalysisResults                 types.List `tfsdk:"resource_analysis_results"`
	VmAnalyticsResults                      types.List `tfsdk:"vm_analytics_results"`
	VmMonitoringResults                     types.List `tfsdk:"vm_monitoring_results"`
}

// ---------------------------------------------------------------------------
// Query input models
// ---------------------------------------------------------------------------

type k8sChangingMetricsQueryModel struct {
	DatasetName   types.String                        `tfsdk:"dataset_name"`
	CoreClusterId types.Int64                         `tfsdk:"core_cluster_id"`
	Filters       *k8sChangingMetricsQueryFiltersModel `tfsdk:"filters"`
}

type k8sChangingMetricsQueryFiltersModel struct {
	ObjectType                  types.String `tfsdk:"object_type"`
	ObjectName                  types.String `tfsdk:"object_name"`
	BreakdownLevel              types.String `tfsdk:"breakdown_level"`
	ChangingMetrics             types.List   `tfsdk:"changing_metrics"`
	Timespan                    types.String `tfsdk:"timespan"`
	CustomFilterState           types.String `tfsdk:"custom_filter_state"`
	CustomFilterAvgCpuRule      types.String `tfsdk:"custom_filter_avg_cpu_rule"`
	CustomFilterAvgCpuValue     types.Float64 `tfsdk:"custom_filter_avg_cpu_value"`
	CustomFilterAvgMemoryRule   types.String `tfsdk:"custom_filter_avg_memory_rule"`
	CustomFilterAvgMemoryValue  types.Float64 `tfsdk:"custom_filter_avg_memory_value"`
	CustomFilterAvgStorageRule  types.String `tfsdk:"custom_filter_avg_storage_rule"`
	CustomFilterAvgStorageValue types.Float64 `tfsdk:"custom_filter_avg_storage_value"`
	CustomFilterSubobjects      types.List   `tfsdk:"custom_filter_subobjects"`
}

type k8sCurrentStateQueryModel struct {
	DatasetName   types.String                       `tfsdk:"dataset_name"`
	CoreClusterId types.Int64                        `tfsdk:"core_cluster_id"`
	Filters       *k8sCurrentStateQueryFiltersModel   `tfsdk:"filters"`
}

type k8sCurrentStateQueryFiltersModel struct {
	ObjectType                  types.String `tfsdk:"object_type"`
	ObjectName                  types.String `tfsdk:"object_name"`
	BreakdownLevel              types.String `tfsdk:"breakdown_level"`
	CurrentStateMetrics         types.List   `tfsdk:"current_state_metrics"`
	CustomFilterState           types.String `tfsdk:"custom_filter_state"`
	CustomFilterAvgCpuRule      types.String `tfsdk:"custom_filter_avg_cpu_rule"`
	CustomFilterAvgCpuValue     types.Float64 `tfsdk:"custom_filter_avg_cpu_value"`
	CustomFilterAvgMemoryRule   types.String `tfsdk:"custom_filter_avg_memory_rule"`
	CustomFilterAvgMemoryValue  types.Float64 `tfsdk:"custom_filter_avg_memory_value"`
	CustomFilterAvgStorageRule  types.String `tfsdk:"custom_filter_avg_storage_rule"`
	CustomFilterAvgStorageValue types.Float64 `tfsdk:"custom_filter_avg_storage_value"`
	CustomFilterSubobjects      types.List   `tfsdk:"custom_filter_subobjects"`
}

type k8sMetricsQueryModel struct {
	DatasetName   types.String               `tfsdk:"dataset_name"`
	CoreClusterId types.Int64                `tfsdk:"core_cluster_id"`
	Filters       *k8sMetricsQueryFiltersModel `tfsdk:"filters"`
}

type k8sMetricsQueryFiltersModel struct {
	ObjectType     types.String `tfsdk:"object_type"`
	ObjectName     types.String `tfsdk:"object_name"`
	BreakdownLevel types.String `tfsdk:"breakdown_level"`
}

type k8sObjectStatesQueryModel struct {
	DatasetName   types.String                    `tfsdk:"dataset_name"`
	CoreClusterId types.Int64                     `tfsdk:"core_cluster_id"`
	Filters       *k8sObjectStatesQueryFiltersModel `tfsdk:"filters"`
}

type k8sObjectStatesQueryFiltersModel struct {
	ObjectType          types.String `tfsdk:"object_type"`
	ObjectName          types.String `tfsdk:"object_name"`
	BreakdownLevel      types.String `tfsdk:"breakdown_level"`
	ObjectStatesMetrics types.List   `tfsdk:"object_states_metrics"`
}

type k8sObjectsQueryModel struct {
	DatasetName   types.String `tfsdk:"dataset_name"`
	CoreClusterId types.Int64  `tfsdk:"core_cluster_id"`
}

type productStatisticsQueryModel struct {
	DatasetName types.String                      `tfsdk:"dataset_name"`
	Filters     *productStatisticsQueryFiltersModel `tfsdk:"filters"`
}

type productStatisticsQueryFiltersModel struct {
	ServiceFilter types.String `tfsdk:"service_filter"`
}

type projectSummaryQueryModel struct {
	DatasetName types.String `tfsdk:"dataset_name"`
}

type resourceAnalysisQueryModel struct {
	DatasetName types.String                     `tfsdk:"dataset_name"`
	Filters     *resourceAnalysisQueryFiltersModel `tfsdk:"filters"`
}

type resourceAnalysisQueryFiltersModel struct {
	PeriodStart types.String `tfsdk:"period_start"`
	PeriodEnd   types.String `tfsdk:"period_end"`
}

type vmAnalyticsQueryModel struct {
	DatasetName types.String `tfsdk:"dataset_name"`
	VmId        types.Int64  `tfsdk:"vm_id"`
}

type vmMonitoringQueryModel struct {
	DatasetName types.String               `tfsdk:"dataset_name"`
	VmId        types.Int64                `tfsdk:"vm_id"`
	Filters     *vmMonitoringQueryFiltersModel `tfsdk:"filters"`
}

type vmMonitoringQueryFiltersModel struct {
	Period types.String `tfsdk:"period"`
}

// ---------------------------------------------------------------------------
// Result item models
// ---------------------------------------------------------------------------

type k8sChangingMetricsResultModel struct {
	SubobjectName   types.String  `tfsdk:"subobject_name"`
	MetricName      types.String  `tfsdk:"metric_name"`
	UiMetricGroup   types.String  `tfsdk:"ui_metric_group"`
	UiColorGroupId  types.Int64   `tfsdk:"ui_color_group_id"`
	UiMetricName    types.String  `tfsdk:"ui_metric_name"`
	HumanLabel      types.String  `tfsdk:"human_label"`
	Timecodes       types.List    `tfsdk:"timecodes"`
	LinechartValues types.List    `tfsdk:"linechart_values"`
	TreemapValues   types.Float64 `tfsdk:"treemap_values"`
}

func (o k8sChangingMetricsResultModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"subobject_name":   types.StringType,
		"metric_name":      types.StringType,
		"ui_metric_group":  types.StringType,
		"ui_color_group_id": types.Int64Type,
		"ui_metric_name":   types.StringType,
		"human_label":      types.StringType,
		"timecodes":        types.ListType{ElemType: types.StringType},
		"linechart_values": types.ListType{ElemType: types.Float64Type},
		"treemap_values":   types.Float64Type,
	}
}

type k8sCurrentStateResultModel struct {
	SubobjectName  types.String  `tfsdk:"subobject_name"`
	MetricName     types.String  `tfsdk:"metric_name"`
	UiMetricGroup  types.String  `tfsdk:"ui_metric_group"`
	UiColorGroupId types.Int64   `tfsdk:"ui_color_group_id"`
	UiMetricName   types.String  `tfsdk:"ui_metric_name"`
	Value          types.Float64 `tfsdk:"value"`
}

func (o k8sCurrentStateResultModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"subobject_name":   types.StringType,
		"metric_name":      types.StringType,
		"ui_metric_group":  types.StringType,
		"ui_color_group_id": types.Int64Type,
		"ui_metric_name":   types.StringType,
		"value":            types.Float64Type,
	}
}

type k8sMetricsResultModel struct {
	MetricName    types.String `tfsdk:"metric_name"`
	UiMetricGroup types.String `tfsdk:"ui_metric_group"`
	UiMetricName  types.String `tfsdk:"ui_metric_name"`
	BlockName     types.String `tfsdk:"block_name"`
}

func (o k8sMetricsResultModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"metric_name":    types.StringType,
		"ui_metric_group": types.StringType,
		"ui_metric_name": types.StringType,
		"block_name":     types.StringType,
	}
}

type k8sObjectStatesResultModel struct {
	MetricName     types.String `tfsdk:"metric_name"`
	UiMetricGroup  types.String `tfsdk:"ui_metric_group"`
	UiColorGroupId types.List   `tfsdk:"ui_color_group_id"`
	UiMetricName   types.String `tfsdk:"ui_metric_name"`
	SubobjectName  types.String `tfsdk:"subobject_name"`
	Value          types.List   `tfsdk:"value"`
}

func (o k8sObjectStatesResultModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"metric_name":      types.StringType,
		"ui_metric_group":  types.StringType,
		"ui_color_group_id": types.ListType{ElemType: types.Int64Type},
		"ui_metric_name":   types.StringType,
		"subobject_name":   types.StringType,
		"value":            types.ListType{ElemType: types.StringType},
	}
}

type k8sObjectsResultModel struct {
	ObjectType types.String `tfsdk:"object_type"`
	ObjectName types.String `tfsdk:"object_name"`
}

func (o k8sObjectsResultModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"object_type": types.StringType,
		"object_name": types.StringType,
	}
}

type productStatisticsResultModel struct {
	Service          types.String  `tfsdk:"service"`
	VmId             types.Int64   `tfsdk:"vm_id"`
	VmName           types.String  `tfsdk:"vm_name"`
	HeadProductId    types.Int64   `tfsdk:"head_product_id"`
	HeadProductName  types.String  `tfsdk:"head_product_name"`
	Currency         types.String  `tfsdk:"currency"`
	Cost             types.Float64 `tfsdk:"cost"`
	ProviderName     types.String  `tfsdk:"provider_name"`
	Country          types.String  `tfsdk:"country"`
	Location         types.String  `tfsdk:"location"`
	Latitude         types.Float64 `tfsdk:"latitude"`
	Longitude        types.Float64 `tfsdk:"longitude"`
	StatusNormalized types.String  `tfsdk:"status_normalized"`
	CpuTotal         types.Float64 `tfsdk:"cpu_total"`
	RamTotal         types.Float64 `tfsdk:"ram_total"`
	DiskUsageTotal   types.Float64 `tfsdk:"disk_usage_total"`
	CpuUsage         types.Float64 `tfsdk:"cpu_usage"`
	RamUsage         types.Float64 `tfsdk:"ram_usage"`
	DiskUsage        types.Float64 `tfsdk:"disk_usage"`
	EmptyValue       types.Int64   `tfsdk:"empty_value"`
}

func (o productStatisticsResultModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"service":           types.StringType,
		"vm_id":             types.Int64Type,
		"vm_name":           types.StringType,
		"head_product_id":   types.Int64Type,
		"head_product_name": types.StringType,
		"currency":          types.StringType,
		"cost":              types.Float64Type,
		"provider_name":     types.StringType,
		"country":           types.StringType,
		"location":          types.StringType,
		"latitude":          types.Float64Type,
		"longitude":         types.Float64Type,
		"status_normalized": types.StringType,
		"cpu_total":         types.Float64Type,
		"ram_total":         types.Float64Type,
		"disk_usage_total":  types.Float64Type,
		"cpu_usage":         types.Float64Type,
		"ram_usage":         types.Float64Type,
		"disk_usage":        types.Float64Type,
		"empty_value":       types.Int64Type,
	}
}

type projectSummaryResultModel struct {
	Service          types.String  `tfsdk:"service"`
	AllInstallations types.Int64   `tfsdk:"all_installations"`
	Cost             types.Float64 `tfsdk:"cost"`
}

func (o projectSummaryResultModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"service":           types.StringType,
		"all_installations": types.Int64Type,
		"cost":              types.Float64Type,
	}
}

type resourceAnalysisResultModel struct {
	Service           types.String  `tfsdk:"service"`
	Timecode          types.String  `tfsdk:"timecode"`
	CpuCoresNumber    types.Float64 `tfsdk:"cpu_cores_number"`
	RamTotalAmountGb  types.Float64 `tfsdk:"ram_total_amount_gb"`
	DiskSpaceTotalGb  types.Float64 `tfsdk:"disk_space_total_gb"`
	Type              types.String  `tfsdk:"type"`
}

func (o resourceAnalysisResultModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"service":             types.StringType,
		"timecode":            types.StringType,
		"cpu_cores_number":    types.Float64Type,
		"ram_total_amount_gb": types.Float64Type,
		"disk_space_total_gb": types.Float64Type,
		"type":                types.StringType,
	}
}

type vmAnalyticsResultModel struct {
	VmId                          types.Int64   `tfsdk:"vm_id"`
	Timecode                      types.String  `tfsdk:"timecode"`
	AvgDateStart                  types.String  `tfsdk:"avg_date_start"`
	AvgDateEnd                    types.String  `tfsdk:"avg_date_end"`
	QuantilesDateStart            types.String  `tfsdk:"quantiles_date_start"`
	QuantilesDateEnd              types.String  `tfsdk:"quantiles_date_end"`
	CpuDataPresent                types.Int64   `tfsdk:"cpu_data_present"`
	CpuUtilizationAverageCores    types.Float64 `tfsdk:"cpu_utilization_average_cores"`
	CpuCoresNumber                types.Int64   `tfsdk:"cpu_cores_number"`
	CpuTotalPercent               types.Int64   `tfsdk:"cpu_total_percent"`
	CpuHumanLabel                 types.String  `tfsdk:"cpu_human_label"`
	RamDataPresent                types.Int64   `tfsdk:"ram_data_present"`
	RamUsageAverageMb             types.Float64 `tfsdk:"ram_usage_average_mb"`
	RamTotalAmountMb              types.Int64   `tfsdk:"ram_total_amount_mb"`
	RamHumanLabel                 types.String  `tfsdk:"ram_human_label"`
	DiskUsedDataPresent           types.Int64   `tfsdk:"disk_used_data_present"`
	DiskSpaceUsedGb               types.Float64 `tfsdk:"disk_space_used_gb"`
	DiskSpaceTotalGb              types.Float64 `tfsdk:"disk_space_total_gb"`
	DiskSpaceHumanLabel           types.String  `tfsdk:"disk_space_human_label"`
	DiskWriteDataPresent          types.Int64   `tfsdk:"disk_write_data_present"`
	DiskWriteBps                  types.Float64 `tfsdk:"disk_write_bps"`
	DiskWriteHuman                types.Float64 `tfsdk:"disk_write_human"`
	DiskWriteHumanLabel           types.String  `tfsdk:"disk_write_human_label"`
	DiskReadDataPresent           types.Int64   `tfsdk:"disk_read_data_present"`
	DiskReadBps                   types.Float64 `tfsdk:"disk_read_bps"`
	DiskReadHuman                 types.Float64 `tfsdk:"disk_read_human"`
	DiskReadHumanLabel            types.String  `tfsdk:"disk_read_human_label"`
	NetworkOutDataPresent         types.Int64   `tfsdk:"network_out_data_present"`
	NetworkOutBps                 types.Float64 `tfsdk:"network_out_bps"`
	NetworkOutHuman               types.Float64 `tfsdk:"network_out_human"`
	NetworkOutHumanLabel          types.String  `tfsdk:"network_out_human_label"`
	NetworkInDataPresent          types.Int64   `tfsdk:"network_in_data_present"`
	NetworkInBps                  types.Float64 `tfsdk:"network_in_bps"`
	NetworkInHuman                types.Float64 `tfsdk:"network_in_human"`
	NetworkInHumanLabel           types.String  `tfsdk:"network_in_human_label"`
	GpuDataPresent                types.Int64   `tfsdk:"gpu_data_present"`
	GpuUtilizationAvgPercent      types.Float64 `tfsdk:"gpu_utilization_avg_percent"`
	GpuTotalPercent               types.Int64   `tfsdk:"gpu_total_percent"`
	GpuHumanLabel                 types.String  `tfsdk:"gpu_human_label"`
	GpuRamDataPresent             types.Int64   `tfsdk:"gpu_ram_data_present"`
	GpuRamUsageAvgGb              types.Float64 `tfsdk:"gpu_ram_usage_avg_gb"`
	GpuRamTotalGb                 types.Float64 `tfsdk:"gpu_ram_total_gb"`
	GpuRamHumanLabel              types.String  `tfsdk:"gpu_ram_human_label"`
	GpuRamUtilizationDataPresent  types.Int64   `tfsdk:"gpu_ram_utilization_data_present"`
	GpuRamUtilizationAvgPercent   types.Float64 `tfsdk:"gpu_ram_utilization_avg_percent"`
	GpuRamUtilizationTotalPercent types.Int64   `tfsdk:"gpu_ram_utilization_total_percent"`
	GpuRamUtilizationHumanLabel   types.String  `tfsdk:"gpu_ram_utilization_human_label"`
	IsShownShort                  types.Int64   `tfsdk:"is_shown_short"`
	Type                          types.String  `tfsdk:"type"`
}

func (o vmAnalyticsResultModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"vm_id":                            types.Int64Type,
		"timecode":                         types.StringType,
		"avg_date_start":                   types.StringType,
		"avg_date_end":                     types.StringType,
		"quantiles_date_start":             types.StringType,
		"quantiles_date_end":               types.StringType,
		"cpu_data_present":                 types.Int64Type,
		"cpu_utilization_average_cores":    types.Float64Type,
		"cpu_cores_number":                 types.Int64Type,
		"cpu_total_percent":                types.Int64Type,
		"cpu_human_label":                  types.StringType,
		"ram_data_present":                 types.Int64Type,
		"ram_usage_average_mb":             types.Float64Type,
		"ram_total_amount_mb":              types.Int64Type,
		"ram_human_label":                  types.StringType,
		"disk_used_data_present":           types.Int64Type,
		"disk_space_used_gb":               types.Float64Type,
		"disk_space_total_gb":              types.Float64Type,
		"disk_space_human_label":           types.StringType,
		"disk_write_data_present":          types.Int64Type,
		"disk_write_bps":                   types.Float64Type,
		"disk_write_human":                 types.Float64Type,
		"disk_write_human_label":           types.StringType,
		"disk_read_data_present":           types.Int64Type,
		"disk_read_bps":                    types.Float64Type,
		"disk_read_human":                  types.Float64Type,
		"disk_read_human_label":            types.StringType,
		"network_out_data_present":         types.Int64Type,
		"network_out_bps":                  types.Float64Type,
		"network_out_human":                types.Float64Type,
		"network_out_human_label":          types.StringType,
		"network_in_data_present":          types.Int64Type,
		"network_in_bps":                   types.Float64Type,
		"network_in_human":                 types.Float64Type,
		"network_in_human_label":           types.StringType,
		"gpu_data_present":                 types.Int64Type,
		"gpu_utilization_avg_percent":      types.Float64Type,
		"gpu_total_percent":                types.Int64Type,
		"gpu_human_label":                  types.StringType,
		"gpu_ram_data_present":             types.Int64Type,
		"gpu_ram_usage_avg_gb":             types.Float64Type,
		"gpu_ram_total_gb":                 types.Float64Type,
		"gpu_ram_human_label":              types.StringType,
		"gpu_ram_utilization_data_present": types.Int64Type,
		"gpu_ram_utilization_avg_percent":  types.Float64Type,
		"gpu_ram_utilization_total_percent": types.Int64Type,
		"gpu_ram_utilization_human_label":  types.StringType,
		"is_shown_short":                   types.Int64Type,
		"type":                             types.StringType,
	}
}

type vmMonitoringResultModel struct {
	Timecode                        types.String  `tfsdk:"timecode"`
	CpuDataPresent                  types.Int64   `tfsdk:"cpu_data_present"`
	CpuUtilizationAverageCores      types.Float64 `tfsdk:"cpu_utilization_average_cores"`
	AvgCpuUtilizationAverageCores   types.Float64 `tfsdk:"avg_cpu_utilization_average_cores"`
	MaxCpuUtilizationAverageCores   types.Float64 `tfsdk:"max_cpu_utilization_average_cores"`
	CpuTotalPercent                 types.Int64   `tfsdk:"cpu_total_percent"`
	CpuHumanLabel                   types.String  `tfsdk:"cpu_human_label"`
	RamDataPresent                  types.Int64   `tfsdk:"ram_data_present"`
	RamUsageAverageGb               types.Float64 `tfsdk:"ram_usage_average_gb"`
	AvgRamUsageAverageGb            types.Float64 `tfsdk:"avg_ram_usage_average_gb"`
	MaxRamUsageAverageGb            types.Float64 `tfsdk:"max_ram_usage_average_gb"`
	RamTotalAmountGb                types.Float64 `tfsdk:"ram_total_amount_gb"`
	RamHumanLabel                   types.String  `tfsdk:"ram_human_label"`
	DiskUsedDataPresent             types.Int64   `tfsdk:"disk_used_data_present"`
	DiskSpaceUsedGb                 types.Float64 `tfsdk:"disk_space_used_gb"`
	AvgDiskSpaceUsedGb              types.Float64 `tfsdk:"avg_disk_space_used_gb"`
	MaxDiskSpaceUsedGb              types.Float64 `tfsdk:"max_disk_space_used_gb"`
	DiskSpaceTotalGb                types.Int64   `tfsdk:"disk_space_total_gb"`
	DiskSpaceHumanLabel             types.String  `tfsdk:"disk_space_human_label"`
	DiskWriteDataPresent            types.Int64   `tfsdk:"disk_write_data_present"`
	DiskWriteCoef                   types.Float64 `tfsdk:"disk_write_coef"`
	DiskWriteHuman                  types.Float64 `tfsdk:"disk_write_human"`
	AvgDiskWriteHuman               types.Float64 `tfsdk:"avg_disk_write_human"`
	MaxDiskWriteHuman               types.Float64 `tfsdk:"max_disk_write_human"`
	DiskWriteHumanLabel             types.String  `tfsdk:"disk_write_human_label"`
	DiskReadDataPresent             types.Int64   `tfsdk:"disk_read_data_present"`
	DiskReadCoef                    types.Float64 `tfsdk:"disk_read_coef"`
	DiskReadHuman                   types.Float64 `tfsdk:"disk_read_human"`
	AvgDiskReadHuman                types.Float64 `tfsdk:"avg_disk_read_human"`
	MaxDiskReadHuman                types.Float64 `tfsdk:"max_disk_read_human"`
	DiskReadHumanLabel              types.String  `tfsdk:"disk_read_human_label"`
	NetworkOutDataPresent           types.Int64   `tfsdk:"network_out_data_present"`
	NetworkOutCoef                  types.Float64 `tfsdk:"network_out_coef"`
	NetworkOutHuman                 types.Float64 `tfsdk:"network_out_human"`
	AvgNetworkOutHuman              types.Float64 `tfsdk:"avg_network_out_human"`
	MaxNetworkOutHuman              types.Float64 `tfsdk:"max_network_out_human"`
	NetworkOutHumanLabel            types.String  `tfsdk:"network_out_human_label"`
	NetworkInDataPresent            types.Int64   `tfsdk:"network_in_data_present"`
	NetworkInCoef                   types.Float64 `tfsdk:"network_in_coef"`
	NetworkInHuman                  types.Float64 `tfsdk:"network_in_human"`
	AvgNetworkInHuman               types.Float64 `tfsdk:"avg_network_in_human"`
	MaxNetworkInHuman               types.Float64 `tfsdk:"max_network_in_human"`
	NetworkInHumanLabel             types.String  `tfsdk:"network_in_human_label"`
	GpuRamUsageDataPresent          types.Int64   `tfsdk:"gpu_ram_usage_data_present"`
	GpuRamUsageAvgGb                types.Float64 `tfsdk:"gpu_ram_usage_avg_gb"`
	AvgGpuRamUsageAvgGb             types.Float64 `tfsdk:"avg_gpu_ram_usage_avg_gb"`
	MaxGpuRamUsageAvgGb             types.Float64 `tfsdk:"max_gpu_ram_usage_avg_gb"`
	GpuRamUsageHumanLabel           types.String  `tfsdk:"gpu_ram_usage_human_label"`
	GpuRamUtilizationAvgPresent     types.Int64   `tfsdk:"gpu_ram_utilization_avg_present"`
	GpuRamUtilizationAvgPercent     types.Float64 `tfsdk:"gpu_ram_utilization_avg_percent"`
	AvgGpuRamUtilizationAvgPercent  types.Float64 `tfsdk:"avg_gpu_ram_utilization_avg_percent"`
	MaxGpuRamUtilizationAvgPercent  types.Float64 `tfsdk:"max_gpu_ram_utilization_avg_percent"`
	GpuRamUtilizationHumanLabel     types.String  `tfsdk:"gpu_ram_utilization_human_label"`
	GpuUtilizationDataPresent       types.Int64   `tfsdk:"gpu_utilization_data_present"`
	GpuUtilizationAvgPercent        types.Float64 `tfsdk:"gpu_utilization_avg_percent"`
	AvgGpuUtilizationAvgPercent     types.Float64 `tfsdk:"avg_gpu_utilization_avg_percent"`
	MaxGpuUtilizationAvgPercent     types.Float64 `tfsdk:"max_gpu_utilization_avg_percent"`
	GpuUtilizationHumanLabel        types.String  `tfsdk:"gpu_utilization_human_label"`
}

func (o vmMonitoringResultModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"timecode":                          types.StringType,
		"cpu_data_present":                  types.Int64Type,
		"cpu_utilization_average_cores":     types.Float64Type,
		"avg_cpu_utilization_average_cores": types.Float64Type,
		"max_cpu_utilization_average_cores": types.Float64Type,
		"cpu_total_percent":                 types.Int64Type,
		"cpu_human_label":                   types.StringType,
		"ram_data_present":                  types.Int64Type,
		"ram_usage_average_gb":              types.Float64Type,
		"avg_ram_usage_average_gb":          types.Float64Type,
		"max_ram_usage_average_gb":          types.Float64Type,
		"ram_total_amount_gb":               types.Float64Type,
		"ram_human_label":                   types.StringType,
		"disk_used_data_present":            types.Int64Type,
		"disk_space_used_gb":                types.Float64Type,
		"avg_disk_space_used_gb":            types.Float64Type,
		"max_disk_space_used_gb":            types.Float64Type,
		"disk_space_total_gb":               types.Int64Type,
		"disk_space_human_label":            types.StringType,
		"disk_write_data_present":           types.Int64Type,
		"disk_write_coef":                   types.Float64Type,
		"disk_write_human":                  types.Float64Type,
		"avg_disk_write_human":              types.Float64Type,
		"max_disk_write_human":              types.Float64Type,
		"disk_write_human_label":            types.StringType,
		"disk_read_data_present":            types.Int64Type,
		"disk_read_coef":                    types.Float64Type,
		"disk_read_human":                   types.Float64Type,
		"avg_disk_read_human":               types.Float64Type,
		"max_disk_read_human":               types.Float64Type,
		"disk_read_human_label":             types.StringType,
		"network_out_data_present":          types.Int64Type,
		"network_out_coef":                  types.Float64Type,
		"network_out_human":                 types.Float64Type,
		"avg_network_out_human":             types.Float64Type,
		"max_network_out_human":             types.Float64Type,
		"network_out_human_label":           types.StringType,
		"network_in_data_present":           types.Int64Type,
		"network_in_coef":                   types.Float64Type,
		"network_in_human":                  types.Float64Type,
		"avg_network_in_human":              types.Float64Type,
		"max_network_in_human":              types.Float64Type,
		"network_in_human_label":            types.StringType,
		"gpu_ram_usage_data_present":        types.Int64Type,
		"gpu_ram_usage_avg_gb":              types.Float64Type,
		"avg_gpu_ram_usage_avg_gb":          types.Float64Type,
		"max_gpu_ram_usage_avg_gb":          types.Float64Type,
		"gpu_ram_usage_human_label":         types.StringType,
		"gpu_ram_utilization_avg_present":   types.Int64Type,
		"gpu_ram_utilization_avg_percent":   types.Float64Type,
		"avg_gpu_ram_utilization_avg_percent": types.Float64Type,
		"max_gpu_ram_utilization_avg_percent": types.Float64Type,
		"gpu_ram_utilization_human_label":   types.StringType,
		"gpu_utilization_data_present":      types.Int64Type,
		"gpu_utilization_avg_percent":       types.Float64Type,
		"avg_gpu_utilization_avg_percent":   types.Float64Type,
		"max_gpu_utilization_avg_percent":   types.Float64Type,
		"gpu_utilization_human_label":       types.StringType,
	}
}

// ---------------------------------------------------------------------------
// Metadata
// ---------------------------------------------------------------------------

func (d *statisticalDataDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_statistical_data"
}

// ---------------------------------------------------------------------------
// Schema
// ---------------------------------------------------------------------------

func (d *statisticalDataDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		DeprecationMessage: "The underlying API endpoint POST /v1/statistics/retrieve was deprecated on 2026-05-13. " +
			"Use this data source only for historical reporting until a replacement becomes available.",
		Description: "Fetches statistical data from emma's DWH. Exactly one query block must be configured. " +
			"The corresponding result list attribute will be populated in state.\n\n" +
			"DEPRECATED: The underlying API endpoint was deprecated on 2026-05-13.",
		Attributes: map[string]schema.Attribute{
			// ---- query inputs ----
			"kubernetes_cluster_changing_metrics": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Query for Kubernetes cluster changing metrics over time.",
				Attributes: map[string]schema.Attribute{
					"dataset_name":    schema.StringAttribute{Required: true, Description: "Dataset name identifier."},
					"core_cluster_id": schema.Int64Attribute{Required: true, Description: "Core cluster ID."},
					"filters": schema.SingleNestedAttribute{
						Required:    true,
						Description: "Filter parameters for the query.",
						Attributes: map[string]schema.Attribute{
							"object_type":                    schema.StringAttribute{Required: true, Description: "Object type to filter."},
							"object_name":                    schema.StringAttribute{Required: true, Description: "Object name to filter."},
							"breakdown_level":                schema.StringAttribute{Required: true, Description: "Breakdown level."},
							"changing_metrics":               schema.ListAttribute{Required: true, ElementType: types.StringType, Description: "List of changing metric names."},
							"timespan":                       schema.StringAttribute{Required: true, Description: "Timespan for the query."},
							"custom_filter_state":            schema.StringAttribute{Required: true, Description: "Custom filter state."},
							"custom_filter_avg_cpu_rule":     schema.StringAttribute{Optional: true, Description: "Custom CPU filter rule (nullable)."},
							"custom_filter_avg_cpu_value":    schema.Float64Attribute{Required: true, Description: "Custom CPU filter value."},
							"custom_filter_avg_memory_rule":  schema.StringAttribute{Required: true, Description: "Custom memory filter rule."},
							"custom_filter_avg_memory_value": schema.Float64Attribute{Required: true, Description: "Custom memory filter value."},
							"custom_filter_avg_storage_rule":  schema.StringAttribute{Required: true, Description: "Custom storage filter rule."},
							"custom_filter_avg_storage_value": schema.Float64Attribute{Required: true, Description: "Custom storage filter value."},
							"custom_filter_subobjects":        schema.ListAttribute{Required: true, ElementType: types.StringType, Description: "List of subobjects to filter."},
						},
					},
				},
			},
			"kubernetes_cluster_current_state": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Query for current state of Kubernetes cluster objects.",
				Attributes: map[string]schema.Attribute{
					"dataset_name":    schema.StringAttribute{Required: true, Description: "Dataset name identifier."},
					"core_cluster_id": schema.Int64Attribute{Required: true, Description: "Core cluster ID."},
					"filters": schema.SingleNestedAttribute{
						Required:    true,
						Description: "Filter parameters for the query.",
						Attributes: map[string]schema.Attribute{
							"object_type":                    schema.StringAttribute{Required: true, Description: "Object type to filter."},
							"object_name":                    schema.StringAttribute{Required: true, Description: "Object name to filter."},
							"breakdown_level":                schema.StringAttribute{Required: true, Description: "Breakdown level."},
							"current_state_metrics":          schema.ListAttribute{Required: true, ElementType: types.StringType, Description: "List of current state metric names."},
							"custom_filter_state":            schema.StringAttribute{Required: true, Description: "Custom filter state."},
							"custom_filter_avg_cpu_rule":     schema.StringAttribute{Optional: true, Description: "Custom CPU filter rule (nullable)."},
							"custom_filter_avg_cpu_value":    schema.Float64Attribute{Required: true, Description: "Custom CPU filter value."},
							"custom_filter_avg_memory_rule":  schema.StringAttribute{Required: true, Description: "Custom memory filter rule."},
							"custom_filter_avg_memory_value": schema.Float64Attribute{Required: true, Description: "Custom memory filter value."},
							"custom_filter_avg_storage_rule":  schema.StringAttribute{Required: true, Description: "Custom storage filter rule."},
							"custom_filter_avg_storage_value": schema.Float64Attribute{Required: true, Description: "Custom storage filter value."},
							"custom_filter_subobjects":        schema.ListAttribute{Required: true, ElementType: types.StringType, Description: "List of subobjects to filter."},
						},
					},
				},
			},
			"kubernetes_cluster_metrics": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Query for Kubernetes cluster metric definitions.",
				Attributes: map[string]schema.Attribute{
					"dataset_name":    schema.StringAttribute{Required: true, Description: "Dataset name identifier."},
					"core_cluster_id": schema.Int64Attribute{Required: true, Description: "Core cluster ID."},
					"filters": schema.SingleNestedAttribute{
						Required:    true,
						Description: "Filter parameters for the query.",
						Attributes: map[string]schema.Attribute{
							"object_type":    schema.StringAttribute{Required: true, Description: "Object type to filter."},
							"object_name":    schema.StringAttribute{Required: true, Description: "Object name to filter."},
							"breakdown_level": schema.StringAttribute{Required: true, Description: "Breakdown level."},
						},
					},
				},
			},
			"kubernetes_cluster_object_states": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Query for Kubernetes cluster object state distributions.",
				Attributes: map[string]schema.Attribute{
					"dataset_name":    schema.StringAttribute{Required: true, Description: "Dataset name identifier."},
					"core_cluster_id": schema.Int64Attribute{Required: true, Description: "Core cluster ID."},
					"filters": schema.SingleNestedAttribute{
						Required:    true,
						Description: "Filter parameters for the query.",
						Attributes: map[string]schema.Attribute{
							"object_type":          schema.StringAttribute{Required: true, Description: "Object type to filter."},
							"object_name":          schema.StringAttribute{Required: true, Description: "Object name to filter."},
							"breakdown_level":      schema.StringAttribute{Required: true, Description: "Breakdown level."},
							"object_states_metrics": schema.ListAttribute{Required: true, ElementType: types.StringType, Description: "List of object state metric names."},
						},
					},
				},
			},
			"kubernetes_cluster_objects": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Query for Kubernetes cluster object inventory.",
				Attributes: map[string]schema.Attribute{
					"dataset_name":    schema.StringAttribute{Required: true, Description: "Dataset name identifier."},
					"core_cluster_id": schema.Int64Attribute{Required: true, Description: "Core cluster ID."},
				},
			},
			"product_statistics": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Query for product-level usage and cost statistics.",
				Attributes: map[string]schema.Attribute{
					"dataset_name": schema.StringAttribute{Required: true, Description: "Dataset name identifier."},
					"filters": schema.SingleNestedAttribute{
						Required:    true,
						Description: "Filter parameters for the query.",
						Attributes: map[string]schema.Attribute{
							"service_filter": schema.StringAttribute{Required: true, Description: "Service name filter."},
						},
					},
				},
			},
			"project_summary": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Query for project-level cost and usage summary.",
				Attributes: map[string]schema.Attribute{
					"dataset_name": schema.StringAttribute{Required: true, Description: "Dataset name identifier."},
				},
			},
			"resource_analysis": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Query for resource analysis over a time period.",
				Attributes: map[string]schema.Attribute{
					"dataset_name": schema.StringAttribute{Required: true, Description: "Dataset name identifier."},
					"filters": schema.SingleNestedAttribute{
						Required:    true,
						Description: "Filter parameters for the query.",
						Attributes: map[string]schema.Attribute{
							"period_start": schema.StringAttribute{Required: true, Description: "Start of the analysis period (ISO 8601)."},
							"period_end":   schema.StringAttribute{Required: true, Description: "End of the analysis period (ISO 8601)."},
						},
					},
				},
			},
			"vm_analytics": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Query for VM-level analytics and utilization statistics.",
				Attributes: map[string]schema.Attribute{
					"dataset_name": schema.StringAttribute{Required: true, Description: "Dataset name identifier."},
					"vm_id":        schema.Int64Attribute{Required: true, Description: "ID of the virtual machine."},
				},
			},
			"vm_monitoring": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Query for VM monitoring time-series data.",
				Attributes: map[string]schema.Attribute{
					"dataset_name": schema.StringAttribute{Required: true, Description: "Dataset name identifier."},
					"vm_id":        schema.Int64Attribute{Required: true, Description: "ID of the virtual machine."},
					"filters": schema.SingleNestedAttribute{
						Required:    true,
						Description: "Filter parameters for the query.",
						Attributes: map[string]schema.Attribute{
							"period": schema.StringAttribute{Required: true, Description: "Monitoring period (e.g. '1h', '24h', '7d')."},
						},
					},
				},
			},

			// ---- computed results ----
			"kubernetes_cluster_changing_metrics_results": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Results when kubernetes_cluster_changing_metrics query is used.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"subobject_name":   schema.StringAttribute{Computed: true, Description: "Subobject name."},
						"metric_name":      schema.StringAttribute{Computed: true, Description: "Metric name."},
						"ui_metric_group":  schema.StringAttribute{Computed: true, Description: "UI metric group."},
						"ui_color_group_id": schema.Int64Attribute{Computed: true, Description: "UI color group ID."},
						"ui_metric_name":   schema.StringAttribute{Computed: true, Description: "UI display name for the metric."},
						"human_label":      schema.StringAttribute{Computed: true, Description: "Human-readable label."},
						"timecodes":        schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "List of timecodes."},
						"linechart_values": schema.ListAttribute{Computed: true, ElementType: types.Float64Type, Description: "Values for line chart."},
						"treemap_values":   schema.Float64Attribute{Computed: true, Description: "Value for treemap visualization."},
					},
				},
			},
			"kubernetes_cluster_current_state_results": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Results when kubernetes_cluster_current_state query is used.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"subobject_name":   schema.StringAttribute{Computed: true, Description: "Subobject name."},
						"metric_name":      schema.StringAttribute{Computed: true, Description: "Metric name."},
						"ui_metric_group":  schema.StringAttribute{Computed: true, Description: "UI metric group."},
						"ui_color_group_id": schema.Int64Attribute{Computed: true, Description: "UI color group ID."},
						"ui_metric_name":   schema.StringAttribute{Computed: true, Description: "UI display name for the metric."},
						"value":            schema.Float64Attribute{Computed: true, Description: "Current state metric value."},
					},
				},
			},
			"kubernetes_cluster_metrics_results": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Results when kubernetes_cluster_metrics query is used.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"metric_name":    schema.StringAttribute{Computed: true, Description: "Metric name."},
						"ui_metric_group": schema.StringAttribute{Computed: true, Description: "UI metric group."},
						"ui_metric_name": schema.StringAttribute{Computed: true, Description: "UI display name for the metric."},
						"block_name":     schema.StringAttribute{Computed: true, Description: "UI block name."},
					},
				},
			},
			"kubernetes_cluster_object_states_results": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Results when kubernetes_cluster_object_states query is used.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"metric_name":      schema.StringAttribute{Computed: true, Description: "Metric name."},
						"ui_metric_group":  schema.StringAttribute{Computed: true, Description: "UI metric group."},
						"ui_color_group_id": schema.ListAttribute{Computed: true, ElementType: types.Int64Type, Description: "UI color group IDs."},
						"ui_metric_name":   schema.StringAttribute{Computed: true, Description: "UI display name for the metric."},
						"subobject_name":   schema.StringAttribute{Computed: true, Description: "Subobject name."},
						"value":            schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "State values."},
					},
				},
			},
			"kubernetes_cluster_objects_results": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Results when kubernetes_cluster_objects query is used.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"object_type": schema.StringAttribute{Computed: true, Description: "Type of the Kubernetes object."},
						"object_name": schema.StringAttribute{Computed: true, Description: "Name of the Kubernetes object."},
					},
				},
			},
			"product_statistics_results": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Results when product_statistics query is used.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"service":           schema.StringAttribute{Computed: true, Description: "Service name."},
						"vm_id":             schema.Int64Attribute{Computed: true, Description: "VM ID."},
						"vm_name":           schema.StringAttribute{Computed: true, Description: "VM name."},
						"head_product_id":   schema.Int64Attribute{Computed: true, Description: "Head product ID."},
						"head_product_name": schema.StringAttribute{Computed: true, Description: "Head product name."},
						"currency":          schema.StringAttribute{Computed: true, Description: "Cost currency."},
						"cost":              schema.Float64Attribute{Computed: true, Description: "Cost value."},
						"provider_name":     schema.StringAttribute{Computed: true, Description: "Cloud provider name."},
						"country":           schema.StringAttribute{Computed: true, Description: "Country."},
						"location":          schema.StringAttribute{Computed: true, Description: "Location."},
						"latitude":          schema.Float64Attribute{Computed: true, Description: "Geographic latitude."},
						"longitude":         schema.Float64Attribute{Computed: true, Description: "Geographic longitude."},
						"status_normalized": schema.StringAttribute{Computed: true, Description: "Normalized status."},
						"cpu_total":         schema.Float64Attribute{Computed: true, Description: "Total CPU."},
						"ram_total":         schema.Float64Attribute{Computed: true, Description: "Total RAM."},
						"disk_usage_total":  schema.Float64Attribute{Computed: true, Description: "Total disk usage."},
						"cpu_usage":         schema.Float64Attribute{Computed: true, Description: "CPU usage."},
						"ram_usage":         schema.Float64Attribute{Computed: true, Description: "RAM usage."},
						"disk_usage":        schema.Float64Attribute{Computed: true, Description: "Disk usage."},
						"empty_value":       schema.Int64Attribute{Computed: true, Description: "Empty value placeholder."},
					},
				},
			},
			"project_summary_results": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Results when project_summary query is used.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"service":           schema.StringAttribute{Computed: true, Description: "Service name."},
						"all_installations": schema.Int64Attribute{Computed: true, Description: "Total number of installations."},
						"cost":              schema.Float64Attribute{Computed: true, Description: "Total cost."},
					},
				},
			},
			"resource_analysis_results": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Results when resource_analysis query is used.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"service":             schema.StringAttribute{Computed: true, Description: "Service name."},
						"timecode":            schema.StringAttribute{Computed: true, Description: "Time code."},
						"cpu_cores_number":    schema.Float64Attribute{Computed: true, Description: "Number of CPU cores."},
						"ram_total_amount_gb": schema.Float64Attribute{Computed: true, Description: "Total RAM in GB."},
						"disk_space_total_gb": schema.Float64Attribute{Computed: true, Description: "Total disk space in GB."},
						"type":                schema.StringAttribute{Computed: true, Description: "Resource type."},
					},
				},
			},
			"vm_analytics_results": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Results when vm_analytics query is used.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"vm_id":                            schema.Int64Attribute{Computed: true, Description: "VM ID."},
						"timecode":                         schema.StringAttribute{Computed: true, Description: "Time code."},
						"avg_date_start":                   schema.StringAttribute{Computed: true, Description: "Average period start date."},
						"avg_date_end":                     schema.StringAttribute{Computed: true, Description: "Average period end date."},
						"quantiles_date_start":             schema.StringAttribute{Computed: true, Description: "Quantiles period start date."},
						"quantiles_date_end":               schema.StringAttribute{Computed: true, Description: "Quantiles period end date."},
						"cpu_data_present":                 schema.Int64Attribute{Computed: true, Description: "CPU data present flag."},
						"cpu_utilization_average_cores":    schema.Float64Attribute{Computed: true, Description: "Average CPU utilization in cores."},
						"cpu_cores_number":                 schema.Int64Attribute{Computed: true, Description: "Total CPU cores."},
						"cpu_total_percent":                schema.Int64Attribute{Computed: true, Description: "CPU total percent."},
						"cpu_human_label":                  schema.StringAttribute{Computed: true, Description: "Human-readable CPU label."},
						"ram_data_present":                 schema.Int64Attribute{Computed: true, Description: "RAM data present flag."},
						"ram_usage_average_mb":             schema.Float64Attribute{Computed: true, Description: "Average RAM usage in MB."},
						"ram_total_amount_mb":              schema.Int64Attribute{Computed: true, Description: "Total RAM in MB."},
						"ram_human_label":                  schema.StringAttribute{Computed: true, Description: "Human-readable RAM label."},
						"disk_used_data_present":           schema.Int64Attribute{Computed: true, Description: "Disk used data present flag."},
						"disk_space_used_gb":               schema.Float64Attribute{Computed: true, Description: "Disk space used in GB."},
						"disk_space_total_gb":              schema.Float64Attribute{Computed: true, Description: "Total disk space in GB."},
						"disk_space_human_label":           schema.StringAttribute{Computed: true, Description: "Human-readable disk label."},
						"disk_write_data_present":          schema.Int64Attribute{Computed: true, Description: "Disk write data present flag."},
						"disk_write_bps":                   schema.Float64Attribute{Computed: true, Description: "Disk write speed in bytes/s."},
						"disk_write_human":                 schema.Float64Attribute{Computed: true, Description: "Disk write human value."},
						"disk_write_human_label":           schema.StringAttribute{Computed: true, Description: "Human-readable disk write label."},
						"disk_read_data_present":           schema.Int64Attribute{Computed: true, Description: "Disk read data present flag."},
						"disk_read_bps":                    schema.Float64Attribute{Computed: true, Description: "Disk read speed in bytes/s."},
						"disk_read_human":                  schema.Float64Attribute{Computed: true, Description: "Disk read human value."},
						"disk_read_human_label":            schema.StringAttribute{Computed: true, Description: "Human-readable disk read label."},
						"network_out_data_present":         schema.Int64Attribute{Computed: true, Description: "Network out data present flag."},
						"network_out_bps":                  schema.Float64Attribute{Computed: true, Description: "Network out speed in bytes/s."},
						"network_out_human":                schema.Float64Attribute{Computed: true, Description: "Network out human value."},
						"network_out_human_label":          schema.StringAttribute{Computed: true, Description: "Human-readable network out label."},
						"network_in_data_present":          schema.Int64Attribute{Computed: true, Description: "Network in data present flag."},
						"network_in_bps":                   schema.Float64Attribute{Computed: true, Description: "Network in speed in bytes/s."},
						"network_in_human":                 schema.Float64Attribute{Computed: true, Description: "Network in human value."},
						"network_in_human_label":           schema.StringAttribute{Computed: true, Description: "Human-readable network in label."},
						"gpu_data_present":                 schema.Int64Attribute{Computed: true, Description: "GPU data present flag."},
						"gpu_utilization_avg_percent":      schema.Float64Attribute{Computed: true, Description: "Average GPU utilization percent."},
						"gpu_total_percent":                schema.Int64Attribute{Computed: true, Description: "GPU total percent."},
						"gpu_human_label":                  schema.StringAttribute{Computed: true, Description: "Human-readable GPU label."},
						"gpu_ram_data_present":             schema.Int64Attribute{Computed: true, Description: "GPU RAM data present flag."},
						"gpu_ram_usage_avg_gb":             schema.Float64Attribute{Computed: true, Description: "Average GPU RAM usage in GB."},
						"gpu_ram_total_gb":                 schema.Float64Attribute{Computed: true, Description: "Total GPU RAM in GB."},
						"gpu_ram_human_label":              schema.StringAttribute{Computed: true, Description: "Human-readable GPU RAM label."},
						"gpu_ram_utilization_data_present": schema.Int64Attribute{Computed: true, Description: "GPU RAM utilization data present flag."},
						"gpu_ram_utilization_avg_percent":  schema.Float64Attribute{Computed: true, Description: "Average GPU RAM utilization percent."},
						"gpu_ram_utilization_total_percent": schema.Int64Attribute{Computed: true, Description: "GPU RAM utilization total percent."},
						"gpu_ram_utilization_human_label":  schema.StringAttribute{Computed: true, Description: "Human-readable GPU RAM utilization label."},
						"is_shown_short":                   schema.Int64Attribute{Computed: true, Description: "Is shown short flag."},
						"type":                             schema.StringAttribute{Computed: true, Description: "Record type."},
					},
				},
			},
			"vm_monitoring_results": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Results when vm_monitoring query is used.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"timecode":                          schema.StringAttribute{Computed: true, Description: "Time code."},
						"cpu_data_present":                  schema.Int64Attribute{Computed: true, Description: "CPU data present flag."},
						"cpu_utilization_average_cores":     schema.Float64Attribute{Computed: true, Description: "Average CPU utilization in cores."},
						"avg_cpu_utilization_average_cores": schema.Float64Attribute{Computed: true, Description: "Period average CPU utilization in cores."},
						"max_cpu_utilization_average_cores": schema.Float64Attribute{Computed: true, Description: "Period max CPU utilization in cores."},
						"cpu_total_percent":                 schema.Int64Attribute{Computed: true, Description: "CPU total percent."},
						"cpu_human_label":                   schema.StringAttribute{Computed: true, Description: "Human-readable CPU label."},
						"ram_data_present":                  schema.Int64Attribute{Computed: true, Description: "RAM data present flag."},
						"ram_usage_average_gb":              schema.Float64Attribute{Computed: true, Description: "Average RAM usage in GB."},
						"avg_ram_usage_average_gb":          schema.Float64Attribute{Computed: true, Description: "Period average RAM usage in GB."},
						"max_ram_usage_average_gb":          schema.Float64Attribute{Computed: true, Description: "Period max RAM usage in GB."},
						"ram_total_amount_gb":               schema.Float64Attribute{Computed: true, Description: "Total RAM in GB."},
						"ram_human_label":                   schema.StringAttribute{Computed: true, Description: "Human-readable RAM label."},
						"disk_used_data_present":            schema.Int64Attribute{Computed: true, Description: "Disk used data present flag."},
						"disk_space_used_gb":                schema.Float64Attribute{Computed: true, Description: "Disk space used in GB."},
						"avg_disk_space_used_gb":            schema.Float64Attribute{Computed: true, Description: "Period average disk space used in GB."},
						"max_disk_space_used_gb":            schema.Float64Attribute{Computed: true, Description: "Period max disk space used in GB."},
						"disk_space_total_gb":               schema.Int64Attribute{Computed: true, Description: "Total disk space in GB."},
						"disk_space_human_label":            schema.StringAttribute{Computed: true, Description: "Human-readable disk label."},
						"disk_write_data_present":           schema.Int64Attribute{Computed: true, Description: "Disk write data present flag."},
						"disk_write_coef":                   schema.Float64Attribute{Computed: true, Description: "Disk write coefficient."},
						"disk_write_human":                  schema.Float64Attribute{Computed: true, Description: "Disk write human value."},
						"avg_disk_write_human":              schema.Float64Attribute{Computed: true, Description: "Period average disk write human value."},
						"max_disk_write_human":              schema.Float64Attribute{Computed: true, Description: "Period max disk write human value."},
						"disk_write_human_label":            schema.StringAttribute{Computed: true, Description: "Human-readable disk write label."},
						"disk_read_data_present":            schema.Int64Attribute{Computed: true, Description: "Disk read data present flag."},
						"disk_read_coef":                    schema.Float64Attribute{Computed: true, Description: "Disk read coefficient."},
						"disk_read_human":                   schema.Float64Attribute{Computed: true, Description: "Disk read human value."},
						"avg_disk_read_human":               schema.Float64Attribute{Computed: true, Description: "Period average disk read human value."},
						"max_disk_read_human":               schema.Float64Attribute{Computed: true, Description: "Period max disk read human value."},
						"disk_read_human_label":             schema.StringAttribute{Computed: true, Description: "Human-readable disk read label."},
						"network_out_data_present":          schema.Int64Attribute{Computed: true, Description: "Network out data present flag."},
						"network_out_coef":                  schema.Float64Attribute{Computed: true, Description: "Network out coefficient."},
						"network_out_human":                 schema.Float64Attribute{Computed: true, Description: "Network out human value."},
						"avg_network_out_human":             schema.Float64Attribute{Computed: true, Description: "Period average network out human value."},
						"max_network_out_human":             schema.Float64Attribute{Computed: true, Description: "Period max network out human value."},
						"network_out_human_label":           schema.StringAttribute{Computed: true, Description: "Human-readable network out label."},
						"network_in_data_present":           schema.Int64Attribute{Computed: true, Description: "Network in data present flag."},
						"network_in_coef":                   schema.Float64Attribute{Computed: true, Description: "Network in coefficient."},
						"network_in_human":                  schema.Float64Attribute{Computed: true, Description: "Network in human value."},
						"avg_network_in_human":              schema.Float64Attribute{Computed: true, Description: "Period average network in human value."},
						"max_network_in_human":              schema.Float64Attribute{Computed: true, Description: "Period max network in human value."},
						"network_in_human_label":            schema.StringAttribute{Computed: true, Description: "Human-readable network in label."},
						"gpu_ram_usage_data_present":        schema.Int64Attribute{Computed: true, Description: "GPU RAM usage data present flag."},
						"gpu_ram_usage_avg_gb":              schema.Float64Attribute{Computed: true, Description: "Average GPU RAM usage in GB."},
						"avg_gpu_ram_usage_avg_gb":          schema.Float64Attribute{Computed: true, Description: "Period average GPU RAM usage in GB."},
						"max_gpu_ram_usage_avg_gb":          schema.Float64Attribute{Computed: true, Description: "Period max GPU RAM usage in GB."},
						"gpu_ram_usage_human_label":         schema.StringAttribute{Computed: true, Description: "Human-readable GPU RAM usage label."},
						"gpu_ram_utilization_avg_present":   schema.Int64Attribute{Computed: true, Description: "GPU RAM utilization avg present flag."},
						"gpu_ram_utilization_avg_percent":   schema.Float64Attribute{Computed: true, Description: "Average GPU RAM utilization percent."},
						"avg_gpu_ram_utilization_avg_percent": schema.Float64Attribute{Computed: true, Description: "Period average GPU RAM utilization percent."},
						"max_gpu_ram_utilization_avg_percent": schema.Float64Attribute{Computed: true, Description: "Period max GPU RAM utilization percent."},
						"gpu_ram_utilization_human_label":   schema.StringAttribute{Computed: true, Description: "Human-readable GPU RAM utilization label."},
						"gpu_utilization_data_present":      schema.Int64Attribute{Computed: true, Description: "GPU utilization data present flag."},
						"gpu_utilization_avg_percent":       schema.Float64Attribute{Computed: true, Description: "Average GPU utilization percent."},
						"avg_gpu_utilization_avg_percent":   schema.Float64Attribute{Computed: true, Description: "Period average GPU utilization percent."},
						"max_gpu_utilization_avg_percent":   schema.Float64Attribute{Computed: true, Description: "Period max GPU utilization percent."},
						"gpu_utilization_human_label":       schema.StringAttribute{Computed: true, Description: "Human-readable GPU utilization label."},
					},
				},
			},
		},
	}
}

// ---------------------------------------------------------------------------
// Configure
// ---------------------------------------------------------------------------

func (d *statisticalDataDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *Client, got: %T. Please report this issue to the provider developers.",
				req.ProviderData))
		return
	}
	d.apiClient = client.apiClient
	d.token = client.token
}

// ---------------------------------------------------------------------------
// Read
// ---------------------------------------------------------------------------

func (d *statisticalDataDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data statisticalDataDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate exactly one query variant is set
	queryCount := 0
	if data.KubernetesClusterChangingMetrics != nil {
		queryCount++
	}
	if data.KubernetesClusterCurrentState != nil {
		queryCount++
	}
	if data.KubernetesClusterMetrics != nil {
		queryCount++
	}
	if data.KubernetesClusterObjectStates != nil {
		queryCount++
	}
	if data.KubernetesClusterObjects != nil {
		queryCount++
	}
	if data.ProductStatistics != nil {
		queryCount++
	}
	if data.ProjectSummary != nil {
		queryCount++
	}
	if data.ResourceAnalysis != nil {
		queryCount++
	}
	if data.VmAnalytics != nil {
		queryCount++
	}
	if data.VmMonitoring != nil {
		queryCount++
	}
	if queryCount != 1 {
		resp.Diagnostics.AddError("Configuration Error",
			fmt.Sprintf("Exactly one query block must be set, got %d.", queryCount))
		return
	}

	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *d.token.AccessToken)

	// Build the SDK request body based on which query block is set.
	var body emmaSdk.GetStatisticalDataRequest

	switch {
	case data.KubernetesClusterChangingMetrics != nil:
		q := data.KubernetesClusterChangingMetrics
		f := q.Filters
		changingMetrics := make([]string, 0)
		f.ChangingMetrics.ElementsAs(ctx, &changingMetrics, false)
		subobjects := make([]string, 0)
		f.CustomFilterSubobjects.ElementsAs(ctx, &subobjects, false)
		cpuRule := emmaSdk.NullableString{}
		if !f.CustomFilterAvgCpuRule.IsNull() {
			v := f.CustomFilterAvgCpuRule.ValueString()
			cpuRule.Set(&v)
		}
		filters := emmaSdk.KubernetesClusterChangingMetricsQueryFilters{
			ObjectType:                  f.ObjectType.ValueString(),
			ObjectName:                  f.ObjectName.ValueString(),
			BreakdownLevel:              f.BreakdownLevel.ValueString(),
			ChangingMetrics:             changingMetrics,
			Timespan:                    f.Timespan.ValueString(),
			CustomFilterState:           f.CustomFilterState.ValueString(),
			CustomFilterAvgCpuRule:      cpuRule,
			CustomFilterAvgCpuValue:     float32(f.CustomFilterAvgCpuValue.ValueFloat64()),
			CustomFilterAvgMemoryRule:   f.CustomFilterAvgMemoryRule.ValueString(),
			CustomFilterAvgMemoryValue:  float32(f.CustomFilterAvgMemoryValue.ValueFloat64()),
			CustomFilterAvgStorageRule:  f.CustomFilterAvgStorageRule.ValueString(),
			CustomFilterAvgStorageValue: float32(f.CustomFilterAvgStorageValue.ValueFloat64()),
			CustomFilterSubobjects:      subobjects,
		}
		body = emmaSdk.KubernetesClusterChangingMetricsQueryAsGetStatisticalDataRequest(
			emmaSdk.NewKubernetesClusterChangingMetricsQuery(
				q.DatasetName.ValueString(),
				int32(q.CoreClusterId.ValueInt64()),
				filters,
			),
		)

	case data.KubernetesClusterCurrentState != nil:
		q := data.KubernetesClusterCurrentState
		f := q.Filters
		currentStateMetrics := make([]string, 0)
		f.CurrentStateMetrics.ElementsAs(ctx, &currentStateMetrics, false)
		subobjects := make([]string, 0)
		f.CustomFilterSubobjects.ElementsAs(ctx, &subobjects, false)
		cpuRule := emmaSdk.NullableString{}
		if !f.CustomFilterAvgCpuRule.IsNull() {
			v := f.CustomFilterAvgCpuRule.ValueString()
			cpuRule.Set(&v)
		}
		filters := emmaSdk.KubernetesClusterCurrentStateQueryFilters{
			ObjectType:                  f.ObjectType.ValueString(),
			ObjectName:                  f.ObjectName.ValueString(),
			BreakdownLevel:              f.BreakdownLevel.ValueString(),
			CurrentStateMetrics:         currentStateMetrics,
			CustomFilterState:           f.CustomFilterState.ValueString(),
			CustomFilterAvgCpuRule:      cpuRule,
			CustomFilterAvgCpuValue:     float32(f.CustomFilterAvgCpuValue.ValueFloat64()),
			CustomFilterAvgMemoryRule:   f.CustomFilterAvgMemoryRule.ValueString(),
			CustomFilterAvgMemoryValue:  float32(f.CustomFilterAvgMemoryValue.ValueFloat64()),
			CustomFilterAvgStorageRule:  f.CustomFilterAvgStorageRule.ValueString(),
			CustomFilterAvgStorageValue: float32(f.CustomFilterAvgStorageValue.ValueFloat64()),
			CustomFilterSubobjects:      subobjects,
		}
		body = emmaSdk.KubernetesClusterCurrentStateQueryAsGetStatisticalDataRequest(
			emmaSdk.NewKubernetesClusterCurrentStateQuery(
				q.DatasetName.ValueString(),
				int32(q.CoreClusterId.ValueInt64()),
				filters,
			),
		)

	case data.KubernetesClusterMetrics != nil:
		q := data.KubernetesClusterMetrics
		f := q.Filters
		filters := emmaSdk.KubernetesClusterMetricsQueryFilters{
			ObjectType:     f.ObjectType.ValueString(),
			ObjectName:     f.ObjectName.ValueString(),
			BreakdownLevel: f.BreakdownLevel.ValueString(),
		}
		body = emmaSdk.KubernetesClusterMetricsQueryAsGetStatisticalDataRequest(
			emmaSdk.NewKubernetesClusterMetricsQuery(
				q.DatasetName.ValueString(),
				int32(q.CoreClusterId.ValueInt64()),
				filters,
			),
		)

	case data.KubernetesClusterObjectStates != nil:
		q := data.KubernetesClusterObjectStates
		f := q.Filters
		objectStatesMetrics := make([]string, 0)
		f.ObjectStatesMetrics.ElementsAs(ctx, &objectStatesMetrics, false)
		filters := emmaSdk.KubernetesClusterObjectStatesQueryFilters{
			ObjectType:          f.ObjectType.ValueString(),
			ObjectName:          f.ObjectName.ValueString(),
			BreakdownLevel:      f.BreakdownLevel.ValueString(),
			ObjectStatesMetrics: objectStatesMetrics,
		}
		body = emmaSdk.KubernetesClusterObjectStatesQueryAsGetStatisticalDataRequest(
			emmaSdk.NewKubernetesClusterObjectStatesQuery(
				q.DatasetName.ValueString(),
				int32(q.CoreClusterId.ValueInt64()),
				filters,
			),
		)

	case data.KubernetesClusterObjects != nil:
		q := data.KubernetesClusterObjects
		body = emmaSdk.KubernetesClusterObjectsQueryAsGetStatisticalDataRequest(
			emmaSdk.NewKubernetesClusterObjectsQuery(
				q.DatasetName.ValueString(),
				int32(q.CoreClusterId.ValueInt64()),
			),
		)

	case data.ProductStatistics != nil:
		q := data.ProductStatistics
		filters := emmaSdk.ProductStatisticsQueryFilters{
			ServiceFilter: q.Filters.ServiceFilter.ValueString(),
		}
		body = emmaSdk.ProductStatisticsQueryAsGetStatisticalDataRequest(
			emmaSdk.NewProductStatisticsQuery(q.DatasetName.ValueString(), filters),
		)

	case data.ProjectSummary != nil:
		q := data.ProjectSummary
		body = emmaSdk.ProjectSummaryQueryAsGetStatisticalDataRequest(
			emmaSdk.NewProjectSummaryQuery(q.DatasetName.ValueString()),
		)

	case data.ResourceAnalysis != nil:
		q := data.ResourceAnalysis
		filters := emmaSdk.ResourceAnalysisQueryFilters{
			PeriodStart: q.Filters.PeriodStart.ValueString(),
			PeriodEnd:   q.Filters.PeriodEnd.ValueString(),
		}
		body = emmaSdk.ResourceAnalysisQueryAsGetStatisticalDataRequest(
			emmaSdk.NewResourceAnalysisQuery(q.DatasetName.ValueString(), filters),
		)

	case data.VmAnalytics != nil:
		q := data.VmAnalytics
		body = emmaSdk.VmAnalyticsQueryAsGetStatisticalDataRequest(
			emmaSdk.NewVmAnalyticsQuery(q.DatasetName.ValueString(), int32(q.VmId.ValueInt64())),
		)

	case data.VmMonitoring != nil:
		q := data.VmMonitoring
		filters := emmaSdk.VmMonitoringQueryFilters{
			Period: q.Filters.Period.ValueString(),
		}
		body = emmaSdk.VmMonitoringQueryAsGetStatisticalDataRequest(
			emmaSdk.NewVmMonitoringQuery(q.DatasetName.ValueString(), int32(q.VmId.ValueInt64()), filters),
		)
	}

	result, response, err := d.apiClient.StatisticsAPI.GetStatisticalData(auth).GetStatisticalDataRequest(body).Execute() //nolint:staticcheck
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to retrieve statistical data, got error: %s",
				tools.ExtractErrorMessage(response)))
		return
	}

	// Initialize all result lists to empty; populate only the matching one.
	var diags diag.Diagnostics

	emptyChangingMetrics, d1 := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: k8sChangingMetricsResultModel{}.attrTypes()}, []k8sChangingMetricsResultModel{})
	diags.Append(d1...)
	emptyCurrentState, d2 := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: k8sCurrentStateResultModel{}.attrTypes()}, []k8sCurrentStateResultModel{})
	diags.Append(d2...)
	emptyMetrics, d3 := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: k8sMetricsResultModel{}.attrTypes()}, []k8sMetricsResultModel{})
	diags.Append(d3...)
	emptyObjectStates, d4 := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: k8sObjectStatesResultModel{}.attrTypes()}, []k8sObjectStatesResultModel{})
	diags.Append(d4...)
	emptyObjects, d5 := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: k8sObjectsResultModel{}.attrTypes()}, []k8sObjectsResultModel{})
	diags.Append(d5...)
	emptyProductStats, d6 := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: productStatisticsResultModel{}.attrTypes()}, []productStatisticsResultModel{})
	diags.Append(d6...)
	emptyProjectSummary, d7 := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: projectSummaryResultModel{}.attrTypes()}, []projectSummaryResultModel{})
	diags.Append(d7...)
	emptyResourceAnalysis, d8 := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: resourceAnalysisResultModel{}.attrTypes()}, []resourceAnalysisResultModel{})
	diags.Append(d8...)
	emptyVmAnalytics, d9 := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: vmAnalyticsResultModel{}.attrTypes()}, []vmAnalyticsResultModel{})
	diags.Append(d9...)
	emptyVmMonitoring, d10 := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: vmMonitoringResultModel{}.attrTypes()}, []vmMonitoringResultModel{})
	diags.Append(d10...)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	data.KubernetesClusterChangingMetricsResults = emptyChangingMetrics
	data.KubernetesClusterCurrentStateResults = emptyCurrentState
	data.KubernetesClusterMetricsResults = emptyMetrics
	data.KubernetesClusterObjectStatesResults = emptyObjectStates
	data.KubernetesClusterObjectsResults = emptyObjects
	data.ProductStatisticsResults = emptyProductStats
	data.ProjectSummaryResults = emptyProjectSummary
	data.ResourceAnalysisResults = emptyResourceAnalysis
	data.VmAnalyticsResults = emptyVmAnalytics
	data.VmMonitoringResults = emptyVmMonitoring

	// Map the populated response variant.
	switch {
	case result.ArrayOfKubernetesClusterChangingMetricsResponse != nil:
		list, listDiags := convertK8sChangingMetricsResults(ctx, *result.ArrayOfKubernetesClusterChangingMetricsResponse)
		resp.Diagnostics.Append(listDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.KubernetesClusterChangingMetricsResults = list

	case result.ArrayOfKubernetesClusterCurrentStateResponse != nil:
		list, listDiags := convertK8sCurrentStateResults(ctx, *result.ArrayOfKubernetesClusterCurrentStateResponse)
		resp.Diagnostics.Append(listDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.KubernetesClusterCurrentStateResults = list

	case result.ArrayOfKubernetesClusterMetricsResponse != nil:
		list, listDiags := convertK8sMetricsResults(ctx, *result.ArrayOfKubernetesClusterMetricsResponse)
		resp.Diagnostics.Append(listDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.KubernetesClusterMetricsResults = list

	case result.ArrayOfKubernetesClusterObjectStatesResponse != nil:
		list, listDiags := convertK8sObjectStatesResults(ctx, *result.ArrayOfKubernetesClusterObjectStatesResponse)
		resp.Diagnostics.Append(listDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.KubernetesClusterObjectStatesResults = list

	case result.ArrayOfKubernetesClusterObjectsResponse != nil:
		list, listDiags := convertK8sObjectsResults(ctx, *result.ArrayOfKubernetesClusterObjectsResponse)
		resp.Diagnostics.Append(listDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.KubernetesClusterObjectsResults = list

	case result.ArrayOfProductStatisticsResponse != nil:
		list, listDiags := convertProductStatisticsResults(ctx, *result.ArrayOfProductStatisticsResponse)
		resp.Diagnostics.Append(listDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.ProductStatisticsResults = list

	case result.ArrayOfProjectSummaryResponse != nil:
		list, listDiags := convertProjectSummaryResults(ctx, *result.ArrayOfProjectSummaryResponse)
		resp.Diagnostics.Append(listDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.ProjectSummaryResults = list

	case result.ArrayOfResourceAnalysisResponse != nil:
		list, listDiags := convertResourceAnalysisResults(ctx, *result.ArrayOfResourceAnalysisResponse)
		resp.Diagnostics.Append(listDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.ResourceAnalysisResults = list

	case result.ArrayOfVmAnalyticsResponse != nil:
		list, listDiags := convertVmAnalyticsResults(ctx, *result.ArrayOfVmAnalyticsResponse)
		resp.Diagnostics.Append(listDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.VmAnalyticsResults = list

	case result.ArrayOfVmMonitoringResponse != nil:
		list, listDiags := convertVmMonitoringResults(ctx, *result.ArrayOfVmMonitoringResponse)
		resp.Diagnostics.Append(listDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.VmMonitoringResults = list
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------------------------------------------------------------------------
// Conversion helpers
// ---------------------------------------------------------------------------

func convertK8sChangingMetricsResults(ctx context.Context, items []emmaSdk.KubernetesClusterChangingMetricsResponse) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	models := make([]k8sChangingMetricsResultModel, 0, len(items))
	for _, item := range items {
		m := k8sChangingMetricsResultModel{}
		if item.SubobjectName != nil {
			m.SubobjectName = types.StringValue(*item.SubobjectName)
		} else {
			m.SubobjectName = types.StringNull()
		}
		if item.MetricName != nil {
			m.MetricName = types.StringValue(*item.MetricName)
		} else {
			m.MetricName = types.StringNull()
		}
		if item.UiMetricGroup != nil {
			m.UiMetricGroup = types.StringValue(*item.UiMetricGroup)
		} else {
			m.UiMetricGroup = types.StringNull()
		}
		if item.UiColorGroupId != nil {
			m.UiColorGroupId = types.Int64Value(int64(*item.UiColorGroupId))
		} else {
			m.UiColorGroupId = types.Int64Null()
		}
		if item.UiMetricName != nil {
			m.UiMetricName = types.StringValue(*item.UiMetricName)
		} else {
			m.UiMetricName = types.StringNull()
		}
		if item.HumanLabel != nil {
			m.HumanLabel = types.StringValue(*item.HumanLabel)
		} else {
			m.HumanLabel = types.StringNull()
		}
		timecodesList, ld := types.ListValueFrom(ctx, types.StringType, item.Timecodes)
		diags.Append(ld...)
		m.Timecodes = timecodesList

		// linechart values are []float32 — convert to []float64 for Terraform
		float64Vals := make([]float64, len(item.LinechartValues))
		for i, v := range item.LinechartValues {
			float64Vals[i] = float64(v)
		}
		linechartList, ld2 := types.ListValueFrom(ctx, types.Float64Type, float64Vals)
		diags.Append(ld2...)
		m.LinechartValues = linechartList

		if item.TreemapValues != nil {
			m.TreemapValues = types.Float64Value(float64(*item.TreemapValues))
		} else {
			m.TreemapValues = types.Float64Null()
		}
		models = append(models, m)
	}
	list, ld := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: k8sChangingMetricsResultModel{}.attrTypes()}, models)
	diags.Append(ld...)
	return list, diags
}

func convertK8sCurrentStateResults(ctx context.Context, items []emmaSdk.KubernetesClusterCurrentStateResponse) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	models := make([]k8sCurrentStateResultModel, 0, len(items))
	for _, item := range items {
		m := k8sCurrentStateResultModel{}
		if item.SubobjectName != nil {
			m.SubobjectName = types.StringValue(*item.SubobjectName)
		} else {
			m.SubobjectName = types.StringNull()
		}
		if item.MetricName != nil {
			m.MetricName = types.StringValue(*item.MetricName)
		} else {
			m.MetricName = types.StringNull()
		}
		if item.UiMetricGroup != nil {
			m.UiMetricGroup = types.StringValue(*item.UiMetricGroup)
		} else {
			m.UiMetricGroup = types.StringNull()
		}
		if item.UiColorGroupId != nil {
			m.UiColorGroupId = types.Int64Value(int64(*item.UiColorGroupId))
		} else {
			m.UiColorGroupId = types.Int64Null()
		}
		if item.UiMetricName != nil {
			m.UiMetricName = types.StringValue(*item.UiMetricName)
		} else {
			m.UiMetricName = types.StringNull()
		}
		if item.Value != nil {
			m.Value = types.Float64Value(float64(*item.Value))
		} else {
			m.Value = types.Float64Null()
		}
		models = append(models, m)
	}
	list, ld := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: k8sCurrentStateResultModel{}.attrTypes()}, models)
	diags.Append(ld...)
	return list, diags
}

func convertK8sMetricsResults(ctx context.Context, items []emmaSdk.KubernetesClusterMetricsResponse) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	models := make([]k8sMetricsResultModel, 0, len(items))
	for _, item := range items {
		m := k8sMetricsResultModel{}
		if item.MetricName != nil {
			m.MetricName = types.StringValue(*item.MetricName)
		} else {
			m.MetricName = types.StringNull()
		}
		if item.UiMetricGroup != nil {
			m.UiMetricGroup = types.StringValue(*item.UiMetricGroup)
		} else {
			m.UiMetricGroup = types.StringNull()
		}
		if item.UiMetricName != nil {
			m.UiMetricName = types.StringValue(*item.UiMetricName)
		} else {
			m.UiMetricName = types.StringNull()
		}
		if item.BlockName != nil {
			m.BlockName = types.StringValue(*item.BlockName)
		} else {
			m.BlockName = types.StringNull()
		}
		models = append(models, m)
	}
	list, ld := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: k8sMetricsResultModel{}.attrTypes()}, models)
	diags.Append(ld...)
	return list, diags
}

func convertK8sObjectStatesResults(ctx context.Context, items []emmaSdk.KubernetesClusterObjectStatesResponse) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	models := make([]k8sObjectStatesResultModel, 0, len(items))
	for _, item := range items {
		m := k8sObjectStatesResultModel{}
		if item.MetricName != nil {
			m.MetricName = types.StringValue(*item.MetricName)
		} else {
			m.MetricName = types.StringNull()
		}
		if item.UiMetricGroup != nil {
			m.UiMetricGroup = types.StringValue(*item.UiMetricGroup)
		} else {
			m.UiMetricGroup = types.StringNull()
		}
		// UiColorGroupId is []int32 — convert to []int64
		int64ColorIds := make([]int64, len(item.UiColorGroupId))
		for i, v := range item.UiColorGroupId {
			int64ColorIds[i] = int64(v)
		}
		colorList, ld := types.ListValueFrom(ctx, types.Int64Type, int64ColorIds)
		diags.Append(ld...)
		m.UiColorGroupId = colorList

		if item.UiMetricName != nil {
			m.UiMetricName = types.StringValue(*item.UiMetricName)
		} else {
			m.UiMetricName = types.StringNull()
		}
		if item.SubobjectName != nil {
			m.SubobjectName = types.StringValue(*item.SubobjectName)
		} else {
			m.SubobjectName = types.StringNull()
		}
		valueList, ld2 := types.ListValueFrom(ctx, types.StringType, item.Value)
		diags.Append(ld2...)
		m.Value = valueList

		models = append(models, m)
	}
	list, ld := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: k8sObjectStatesResultModel{}.attrTypes()}, models)
	diags.Append(ld...)
	return list, diags
}

func convertK8sObjectsResults(ctx context.Context, items []emmaSdk.KubernetesClusterObjectsResponse) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	models := make([]k8sObjectsResultModel, 0, len(items))
	for _, item := range items {
		m := k8sObjectsResultModel{}
		if item.ObjectType != nil {
			m.ObjectType = types.StringValue(*item.ObjectType)
		} else {
			m.ObjectType = types.StringNull()
		}
		if item.ObjectName != nil {
			m.ObjectName = types.StringValue(*item.ObjectName)
		} else {
			m.ObjectName = types.StringNull()
		}
		models = append(models, m)
	}
	list, ld := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: k8sObjectsResultModel{}.attrTypes()}, models)
	diags.Append(ld...)
	return list, diags
}

func convertProductStatisticsResults(ctx context.Context, items []emmaSdk.ProductStatisticsResponse) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	models := make([]productStatisticsResultModel, 0, len(items))
	for _, item := range items {
		m := productStatisticsResultModel{}
		ptrString := func(p *string) types.String {
			if p != nil {
				return types.StringValue(*p)
			}
			return types.StringNull()
		}
		ptrInt32 := func(p *int32) types.Int64 {
			if p != nil {
				return types.Int64Value(int64(*p))
			}
			return types.Int64Null()
		}
		ptrFloat32 := func(p *float32) types.Float64 {
			if p != nil {
				return types.Float64Value(float64(*p))
			}
			return types.Float64Null()
		}
		m.Service = ptrString(item.Service)
		m.VmId = ptrInt32(item.VmId)
		m.VmName = ptrString(item.VmName)
		m.HeadProductId = ptrInt32(item.HeadProductId)
		m.HeadProductName = ptrString(item.HeadProductName)
		m.Currency = ptrString(item.Currency)
		m.Cost = ptrFloat32(item.Cost)
		m.ProviderName = ptrString(item.ProviderName)
		m.Country = ptrString(item.Country)
		m.Location = ptrString(item.Location)
		m.Latitude = ptrFloat32(item.Latitude)
		m.Longitude = ptrFloat32(item.Longitude)
		m.StatusNormalized = ptrString(item.StatusNormalized)
		m.CpuTotal = ptrFloat32(item.CpuTotal)
		m.RamTotal = ptrFloat32(item.RamTotal)
		m.DiskUsageTotal = ptrFloat32(item.DiskUsageTotal)
		m.CpuUsage = ptrFloat32(item.CpuUsage)
		m.RamUsage = ptrFloat32(item.RamUsage)
		m.DiskUsage = ptrFloat32(item.DiskUsage)
		m.EmptyValue = ptrInt32(item.EmptyValue)
		models = append(models, m)
	}
	list, ld := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: productStatisticsResultModel{}.attrTypes()}, models)
	diags.Append(ld...)
	return list, diags
}

func convertProjectSummaryResults(ctx context.Context, items []emmaSdk.ProjectSummaryResponse) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	models := make([]projectSummaryResultModel, 0, len(items))
	for _, item := range items {
		m := projectSummaryResultModel{}
		if item.Service != nil {
			m.Service = types.StringValue(*item.Service)
		} else {
			m.Service = types.StringNull()
		}
		if item.AllInstallations != nil {
			m.AllInstallations = types.Int64Value(int64(*item.AllInstallations))
		} else {
			m.AllInstallations = types.Int64Null()
		}
		if item.Cost != nil {
			m.Cost = types.Float64Value(float64(*item.Cost))
		} else {
			m.Cost = types.Float64Null()
		}
		models = append(models, m)
	}
	list, ld := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: projectSummaryResultModel{}.attrTypes()}, models)
	diags.Append(ld...)
	return list, diags
}

func convertResourceAnalysisResults(ctx context.Context, items []emmaSdk.ResourceAnalysisResponse) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	models := make([]resourceAnalysisResultModel, 0, len(items))
	for _, item := range items {
		m := resourceAnalysisResultModel{}
		if item.Service != nil {
			m.Service = types.StringValue(*item.Service)
		} else {
			m.Service = types.StringNull()
		}
		if item.Timecode != nil {
			m.Timecode = types.StringValue(*item.Timecode)
		} else {
			m.Timecode = types.StringNull()
		}
		if item.CpuCoresNumber != nil {
			m.CpuCoresNumber = types.Float64Value(float64(*item.CpuCoresNumber))
		} else {
			m.CpuCoresNumber = types.Float64Null()
		}
		if item.RamTotalAmountGb != nil {
			m.RamTotalAmountGb = types.Float64Value(float64(*item.RamTotalAmountGb))
		} else {
			m.RamTotalAmountGb = types.Float64Null()
		}
		if item.DiskSpaceTotalGb != nil {
			m.DiskSpaceTotalGb = types.Float64Value(float64(*item.DiskSpaceTotalGb))
		} else {
			m.DiskSpaceTotalGb = types.Float64Null()
		}
		if item.Type != nil {
			m.Type = types.StringValue(*item.Type)
		} else {
			m.Type = types.StringNull()
		}
		models = append(models, m)
	}
	list, ld := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: resourceAnalysisResultModel{}.attrTypes()}, models)
	diags.Append(ld...)
	return list, diags
}

func convertVmAnalyticsResults(ctx context.Context, items []emmaSdk.VmAnalyticsResponse) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	models := make([]vmAnalyticsResultModel, 0, len(items))

	pS := func(p *string) types.String {
		if p != nil {
			return types.StringValue(*p)
		}
		return types.StringNull()
	}
	pI32 := func(p *int32) types.Int64 {
		if p != nil {
			return types.Int64Value(int64(*p))
		}
		return types.Int64Null()
	}
	pF32 := func(p *float32) types.Float64 {
		if p != nil {
			return types.Float64Value(float64(*p))
		}
		return types.Float64Null()
	}

	for _, item := range items {
		m := vmAnalyticsResultModel{
			VmId:                          pI32(item.VmId),
			Timecode:                      pS(item.Timecode),
			AvgDateStart:                  pS(item.AvgDateStart),
			AvgDateEnd:                    pS(item.AvgDateEnd),
			QuantilesDateStart:            pS(item.QuantilesDateStart),
			QuantilesDateEnd:              pS(item.QuantilesDateEnd),
			CpuDataPresent:                pI32(item.CpuDataPresent),
			CpuUtilizationAverageCores:    pF32(item.CpuUtilizationAverageCores),
			CpuCoresNumber:                pI32(item.CpuCoresNumber),
			CpuTotalPercent:               pI32(item.CpuTotalPercent),
			CpuHumanLabel:                 pS(item.CpuHumanLabel),
			RamDataPresent:                pI32(item.RamDataPresent),
			RamUsageAverageMb:             pF32(item.RamUsageAverageMb),
			RamTotalAmountMb:              pI32(item.RamTotalAmountMb),
			RamHumanLabel:                 pS(item.RamHumanLabel),
			DiskUsedDataPresent:           pI32(item.DiskUsedDataPresent),
			DiskSpaceUsedGb:               pF32(item.DiskSpaceUsedGb),
			DiskSpaceTotalGb:              pF32(item.DiskSpaceTotalGb),
			DiskSpaceHumanLabel:           pS(item.DiskSpaceHumanLabel),
			DiskWriteDataPresent:          pI32(item.DiskWriteDataPresent),
			DiskWriteBps:                  pF32(item.DiskWriteBps),
			DiskWriteHuman:                pF32(item.DiskWriteHuman),
			DiskWriteHumanLabel:           pS(item.DiskWriteHumanLabel),
			DiskReadDataPresent:           pI32(item.DiskReadDataPresent),
			DiskReadBps:                   pF32(item.DiskReadBps),
			DiskReadHuman:                 pF32(item.DiskReadHuman),
			DiskReadHumanLabel:            pS(item.DiskReadHumanLabel),
			NetworkOutDataPresent:         pI32(item.NetworkOutDataPresent),
			NetworkOutBps:                 pF32(item.NetworkOutBps),
			NetworkOutHuman:               pF32(item.NetworkOutHuman),
			NetworkOutHumanLabel:          pS(item.NetworkOutHumanLabel),
			NetworkInDataPresent:          pI32(item.NetworkInDataPresent),
			NetworkInBps:                  pF32(item.NetworkInBps),
			NetworkInHuman:                pF32(item.NetworkInHuman),
			NetworkInHumanLabel:           pS(item.NetworkInHumanLabel),
			GpuDataPresent:                pI32(item.GpuDataPresent),
			GpuUtilizationAvgPercent:      pF32(item.GpuUtilizationAvgPercent),
			GpuTotalPercent:               pI32(item.GpuTotalPercent),
			GpuHumanLabel:                 pS(item.GpuHumanLabel),
			GpuRamDataPresent:             pI32(item.GpuRamDataPresent),
			GpuRamUsageAvgGb:              pF32(item.GpuRamUsageAvgGb),
			GpuRamTotalGb:                 pF32(item.GpuRamTotalGb),
			GpuRamHumanLabel:              pS(item.GpuRamHumanLabel),
			GpuRamUtilizationDataPresent:  pI32(item.GpuRamUtilizationDataPresent),
			GpuRamUtilizationAvgPercent:   pF32(item.GpuRamUtilizationAvgPercent),
			GpuRamUtilizationTotalPercent: pI32(item.GpuRamUtilizationTotalPercent),
			GpuRamUtilizationHumanLabel:   pS(item.GpuRamUtilizationHumanLabel),
			IsShownShort:                  pI32(item.IsShownShort),
			Type:                          pS(item.Type),
		}
		models = append(models, m)
	}
	list, ld := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: vmAnalyticsResultModel{}.attrTypes()}, models)
	diags.Append(ld...)
	return list, diags
}

func convertVmMonitoringResults(ctx context.Context, items []emmaSdk.VmMonitoringResponse) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	models := make([]vmMonitoringResultModel, 0, len(items))

	pS := func(p *string) types.String {
		if p != nil {
			return types.StringValue(*p)
		}
		return types.StringNull()
	}
	pI32 := func(p *int32) types.Int64 {
		if p != nil {
			return types.Int64Value(int64(*p))
		}
		return types.Int64Null()
	}
	pF32 := func(p *float32) types.Float64 {
		if p != nil {
			return types.Float64Value(float64(*p))
		}
		return types.Float64Null()
	}

	for _, item := range items {
		m := vmMonitoringResultModel{
			Timecode:                        pS(item.Timecode),
			CpuDataPresent:                  pI32(item.CpuDataPresent),
			CpuUtilizationAverageCores:      pF32(item.CpuUtilizationAverageCores),
			AvgCpuUtilizationAverageCores:   pF32(item.AvgCpuUtilizationAverageCores),
			MaxCpuUtilizationAverageCores:   pF32(item.MaxCpuUtilizationAverageCores),
			CpuTotalPercent:                 pI32(item.CpuTotalPercent),
			CpuHumanLabel:                   pS(item.CpuHumanLabel),
			RamDataPresent:                  pI32(item.RamDataPresent),
			RamUsageAverageGb:               pF32(item.RamUsageAverageGb),
			AvgRamUsageAverageGb:            pF32(item.AvgRamUsageAverageGb),
			MaxRamUsageAverageGb:            pF32(item.MaxRamUsageAverageGb),
			RamTotalAmountGb:                pF32(item.RamTotalAmountGb),
			RamHumanLabel:                   pS(item.RamHumanLabel),
			DiskUsedDataPresent:             pI32(item.DiskUsedDataPresent),
			DiskSpaceUsedGb:                 pF32(item.DiskSpaceUsedGb),
			AvgDiskSpaceUsedGb:              pF32(item.AvgDiskSpaceUsedGb),
			MaxDiskSpaceUsedGb:              pF32(item.MaxDiskSpaceUsedGb),
			DiskSpaceTotalGb:                pI32(item.DiskSpaceTotalGb),
			DiskSpaceHumanLabel:             pS(item.DiskSpaceHumanLabel),
			DiskWriteDataPresent:            pI32(item.DiskWriteDataPresent),
			DiskWriteCoef:                   pF32(item.DiskWriteCoef),
			DiskWriteHuman:                  pF32(item.DiskWriteHuman),
			AvgDiskWriteHuman:               pF32(item.AvgDiskWriteHuman),
			MaxDiskWriteHuman:               pF32(item.MaxDiskWriteHuman),
			DiskWriteHumanLabel:             pS(item.DiskWriteHumanLabel),
			DiskReadDataPresent:             pI32(item.DiskReadDataPresent),
			DiskReadCoef:                    pF32(item.DiskReadCoef),
			DiskReadHuman:                   pF32(item.DiskReadHuman),
			AvgDiskReadHuman:                pF32(item.AvgDiskReadHuman),
			MaxDiskReadHuman:                pF32(item.MaxDiskReadHuman),
			DiskReadHumanLabel:              pS(item.DiskReadHumanLabel),
			NetworkOutDataPresent:           pI32(item.NetworkOutDataPresent),
			NetworkOutCoef:                  pF32(item.NetworkOutCoef),
			NetworkOutHuman:                 pF32(item.NetworkOutHuman),
			AvgNetworkOutHuman:              pF32(item.AvgNetworkOutHuman),
			MaxNetworkOutHuman:              pF32(item.MaxNetworkOutHuman),
			NetworkOutHumanLabel:            pS(item.NetworkOutHumanLabel),
			NetworkInDataPresent:            pI32(item.NetworkInDataPresent),
			NetworkInCoef:                   pF32(item.NetworkInCoef),
			NetworkInHuman:                  pF32(item.NetworkInHuman),
			AvgNetworkInHuman:               pF32(item.AvgNetworkInHuman),
			MaxNetworkInHuman:               pF32(item.MaxNetworkInHuman),
			NetworkInHumanLabel:             pS(item.NetworkInHumanLabel),
			GpuRamUsageDataPresent:          pI32(item.GpuRamUsageDataPresent),
			GpuRamUsageAvgGb:                pF32(item.GpuRamUsageAvgGb),
			AvgGpuRamUsageAvgGb:             pF32(item.AvgGpuRamUsageAvgGb),
			MaxGpuRamUsageAvgGb:             pF32(item.MaxGpuRamUsageAvgGb),
			GpuRamUsageHumanLabel:           pS(item.GpuRamUsageHumanLabel),
			GpuRamUtilizationAvgPresent:     pI32(item.GpuRamUtilizationAvgPresent),
			GpuRamUtilizationAvgPercent:     pF32(item.GpuRamUtilizationAvgPercent),
			AvgGpuRamUtilizationAvgPercent:  pF32(item.AvgGpuRamUtilizationAvgPercent),
			MaxGpuRamUtilizationAvgPercent:  pF32(item.MaxGpuRamUtilizationAvgPercent),
			GpuRamUtilizationHumanLabel:     pS(item.GpuRamUtilizationHumanLabel),
			GpuUtilizationDataPresent:       pI32(item.GpuUtilizationDataPresent),
			GpuUtilizationAvgPercent:        pF32(item.GpuUtilizationAvgPercent),
			AvgGpuUtilizationAvgPercent:     pF32(item.AvgGpuUtilizationAvgPercent),
			MaxGpuUtilizationAvgPercent:     pF32(item.MaxGpuUtilizationAvgPercent),
			GpuUtilizationHumanLabel:        pS(item.GpuUtilizationHumanLabel),
		}
		models = append(models, m)
	}
	list, ld := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: vmMonitoringResultModel{}.attrTypes()}, models)
	diags.Append(ld...)
	return list, diags
}
