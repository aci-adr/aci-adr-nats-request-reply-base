package request

type ForexRequest struct {
	TenantId       int    `json:"tenantId"`
	BankId         int    `json:"bankId"`
	BaseCurrency   string `json:"baseCurrency"`
	TargetCurrency string `json:"targetCurrency"`
	Tier           string `json:"tier"`
}
