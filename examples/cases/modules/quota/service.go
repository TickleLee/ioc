package quota

// QuotaService 配额服务接口
type QuotaService interface {
	HasQuota(operation string) bool
	UseQuota(operation string) bool
	ResetQuotas()
}
