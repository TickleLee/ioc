package impl

import (
	"fmt"
	"log"
	"time"

	"github.com/TickleLee/ioc/examples/cases/config"
	"github.com/TickleLee/ioc/examples/cases/modules/product"
	"github.com/TickleLee/ioc/pkg/ioc"
)

// QuotaServiceImpl 配额服务实现
type QuotaServiceImpl struct {
	// 存储每个操作的配额使用情况
	quotas         map[string]int
	Config         *config.Config `inject:"appConfig"`
	lastReset      time.Time
	ProductService product.ProductService `inject:"productService"`
}

// PostConstruct 初始化方法
func (q *QuotaServiceImpl) PostConstruct() error {
	fmt.Println("初始化 QuotaService，最大配额:", q.Config.MaxQuota)
	q.quotas = make(map[string]int)
	q.lastReset = time.Now()
	return nil
}

func (q *QuotaServiceImpl) checkReset() {
	// 检查是否需要重置配额
	if time.Since(q.lastReset) > q.Config.QuotaResetTime {
		q.ResetQuotas()
	}
}

func (q *QuotaServiceImpl) HasQuota(operation string) bool {
	q.checkReset()
	return q.quotas[operation] < q.Config.MaxQuota
}

func (q *QuotaServiceImpl) UseQuota(operation string) bool {
	if !q.HasQuota(operation) {
		return false
	}

	q.quotas[operation]++
	return true
}

func (q *QuotaServiceImpl) ResetQuotas() {
	q.quotas = make(map[string]int)
	q.lastReset = time.Now()
}

func init() {
	err := ioc.Register("quotaService", &QuotaServiceImpl{}, ioc.Singleton)
	if err != nil {
		log.Fatalf("注册 quotaService 失败: %v", err)
	}
}
