package usecase

import (
	"strconv"
	"strings"

	"omnilogs-api/dto"
	"omnilogs-api/models"
)

func (u *usecase) GetSetupStatus(actor Actor, productID int) (*dto.SetupStatusResponse, error) {
	prod, err := u.GetProduct(actor, productID)
	if err != nil {
		return nil, err
	}

	var projectsCount int64
	u.repository.DB().Model(&models.Project{}).Where("product_id = ?", productID).Count(&projectsCount)

	var features []models.ProjectFeature
	u.repository.DB().Where("product_id = ?", productID).Find(&features)

	featuresCount := 0
	subFeaturesCount := 0
	for _, f := range features {
		if f.ParentID == nil {
			featuresCount++
		} else {
			subFeaturesCount++
		}
	}
	totalNodesCount := len(features)

	var apiKeysCount int64
	u.repository.DB().Model(&models.ProductAPIKey{}).Where("product_id = ? AND is_active = TRUE", productID).Count(&apiKeysCount)

	var testLogCount int64
	u.repository.DB().Model(&models.LogQueueItem{}).Where("product_id = ?", productID).Count(&testLogCount)
	testLogReceived := testLogCount > 0

	valid, valErrs, _ := u.ValidateSetupStructure(actor, productID)

	currentStep := 1
	progressPercent := 15
	status := prod.SetupStatus
	if status == "" {
		status = "DRAFT"
	}

	if projectsCount > 0 {
		currentStep = 2
		progressPercent = 30
	}
	if totalNodesCount > 0 {
		currentStep = 3
		progressPercent = 50
	}
	if valid {
		currentStep = 4
		progressPercent = 70
	}
	if apiKeysCount > 0 {
		currentStep = 5
		progressPercent = 85
	}
	if testLogReceived {
		currentStep = 6
		progressPercent = 95
	}
	if status == "ACTIVE" {
		currentStep = 7
		progressPercent = 100
	}

	envs := make([]any, 0, len(prod.ProductEnvironments))
	for _, e := range prod.ProductEnvironments {
		envs = append(envs, e)
	}

	return &dto.SetupStatusResponse{
		ProductID:        prod.ProductID,
		ProductName:      prod.ProductName,
		ProductCode:      prod.ProductCode,
		Description:      prod.Description,
		SetupStatus:      status,
		CurrentStep:      currentStep,
		ProgressPercent:  progressPercent,
		Environments:     envs,
		ProjectsCount:    int(projectsCount),
		FeaturesCount:    featuresCount,
		SubFeaturesCount: subFeaturesCount,
		TotalNodesCount:  totalNodesCount,
		ApiKeysCount:     int(apiKeysCount),
		TestLogReceived:  testLogReceived,
		ValidationErrors: valErrs,
	}, nil
}

func (u *usecase) ValidateSetupStructure(actor Actor, productID int) (bool, []string, error) {
	prod, err := u.GetProduct(actor, productID)
	if err != nil {
		return false, nil, err
	}

	var errs []string

	// 1. A product must have at least one active environment. The Product Setup
	// wizard intentionally configures one environment at a time.
	hasActiveEnvironment := false
	for _, env := range prod.ProductEnvironments {
		if env.IsActive && strings.TrimSpace(env.EnvironmentCode) != "" {
			hasActiveEnvironment = true
			break
		}
	}
	if !hasActiveEnvironment {
		errs = append(errs, "Must create at least 1 active Environment")
	}

	// 2. Must have at least 1 project
	var projects []models.Project
	u.repository.DB().Where("product_id = ?", productID).Find(&projects)
	if len(projects) == 0 {
		errs = append(errs, "Must create at least 1 Project")
	}

	// 3. Must have at least 1 feature
	var features []models.ProjectFeature
	u.repository.DB().Where("product_id = ?", productID).Find(&features)
	if len(features) == 0 {
		errs = append(errs, "Must create at least 1 Feature")
	}

	// 4. Duplicate project codes check
	projCodes := map[string]bool{}
	for _, p := range projects {
		if strings.TrimSpace(p.ProjectCode) == "" || strings.TrimSpace(p.ProjectName) == "" {
			errs = append(errs, "Project name and code must not be empty")
		}
		code := strings.ToUpper(strings.TrimSpace(p.ProjectCode))
		if projCodes[code] {
			errs = append(errs, "Duplicate project code: "+code)
		}
		projCodes[code] = true
	}

	// 5. Feature checks (codes, completeness, circular references)
	featCodesPerProj := map[int]map[string]bool{}
	featMap := map[int]models.ProjectFeature{}
	for _, f := range features {
		featMap[f.CategoryID] = f
		if strings.TrimSpace(f.CategoryCode) == "" || strings.TrimSpace(f.CategoryName) == "" {
			errs = append(errs, "Feature name and code must not be empty")
		}
		if featCodesPerProj[f.ProjectID] == nil {
			featCodesPerProj[f.ProjectID] = map[string]bool{}
		}
		code := strings.ToUpper(strings.TrimSpace(f.CategoryCode))
		if featCodesPerProj[f.ProjectID][code] {
			errs = append(errs, "Duplicate feature code '"+code+"' in project ID "+strconv.Itoa(f.ProjectID))
		}
		featCodesPerProj[f.ProjectID][code] = true
	}

	// Detect circular references
	for _, f := range features {
		curr := f.ParentID
		visited := map[int]bool{f.CategoryID: true}
		for curr != nil {
			if visited[*curr] {
				errs = append(errs, "Circular hierarchy detected in feature '"+f.CategoryName+"'")
				break
			}
			visited[*curr] = true
			if pNode, ok := featMap[*curr]; ok {
				curr = pNode.ParentID
			} else {
				break
			}
		}
	}

	valid := len(errs) == 0
	if valid && (prod.SetupStatus == "DRAFT" || prod.SetupStatus == "STRUCTURE_INCOMPLETE") {
		u.repository.DB().Model(&models.Product{}).Where("product_id = ?", productID).Update("setup_status", "READY_FOR_API_KEY")
	}

	return valid, errs, nil
}

func (u *usecase) CompleteSetup(actor Actor, productID int) error {
	if err := u.requireProductOwner(actor, productID); err != nil {
		return err
	}
	return u.repository.DB().Model(&models.Product{}).Where("product_id = ?", productID).Update("setup_status", "ACTIVE").Error
}
