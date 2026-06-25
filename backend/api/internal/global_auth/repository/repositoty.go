package repository

import productrepo "omnilogs-api/internal/product/repository"

type Repository = productrepo.Repository

var NewRepository = productrepo.NewRepository

type Resource string
type Action string
type Effect string
type ScopeLevel string

const (
	ResourceProduct     Resource = "PRODUCT"
	ResourceProject     Resource = "PROJECT"
	ResourceFeature     Resource = "FEATURE"
	ResourceCategory    Resource = "CATEGORY"
	ResourceRole        Resource = "ROLE"
	ResourceAccess      Resource = "ACCESS"
	ResourceAPIKey      Resource = "API_KEY"
	ResourceEnvironment Resource = "ENVIRONMENT"
	ResourceLog         Resource = "LOG"
)

const (
	ActionCreate        Action = "CREATE"
	ActionRead          Action = "READ"
	ActionUpdate        Action = "UPDATE"
	ActionDelete        Action = "DELETE"
	ActionGrant         Action = "GRANT"
	ActionRevoke        Action = "REVOKE"
	ActionExport        Action = "EXPORT"
	ActionViewSensitive Action = "VIEW_SENSITIVE"
)

const (
	EffectAllow Effect = "ALLOW"
	EffectDeny  Effect = "DENY"
)

const (
	ScopeGlobal   ScopeLevel = "GLOBAL"
	ScopeProduct  ScopeLevel = "PRODUCT"
	ScopeProject  ScopeLevel = "PROJECT"
	ScopeCategory ScopeLevel = "CATEGORY"
)
