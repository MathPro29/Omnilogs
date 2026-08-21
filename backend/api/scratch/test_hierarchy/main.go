package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"omnilogs-api/configs"
	"omnilogs-api/models"

	"gorm.io/gorm"
)

var db *gorm.DB

func main() {
	// Set dotenv path to root of backend api
	os.Setenv("ENV_PATH", "../../.env")

	env := configs.LoadEnv()
	if err := env.Validate(); err != nil {
		log.Fatalf("invalid environment configuration: %v", err)
	}

	var err error
	db, err = configs.ConnectDB(env)
	if err != nil {
		log.Fatalf("connect database failed: %v", err)
	}

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/validate", validateHandler)

	port := ":9999"
	fmt.Printf("==================================================\n")
	fmt.Printf("  UI Test Server is running at http://localhost%s\n", port)
	fmt.Printf("==================================================\n")
	log.Fatal(http.ListenAndServe(port, nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Hierarchy Validator Test</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100 p-8">
    <div class="max-w-4xl mx-auto bg-white p-6 rounded shadow">
        <h1 class="text-2xl font-bold mb-4">Log Hierarchy Validator (Test UI)</h1>
        
        <div class="grid grid-cols-2 gap-6">
            <!-- Form Section -->
            <div>
                <h2 class="text-lg font-semibold mb-2">Input Parameters</h2>
                <div class="space-y-4">
                    <div>
                        <label class="block text-sm font-medium text-gray-700">Product ID</label>
                        <input type="number" id="product_id" value="9999" class="mt-1 block w-full border border-gray-300 rounded-md p-2">
                    </div>
                    <div>
                        <label class="block text-sm font-medium text-gray-700">Environment ID</label>
                        <input type="number" id="environment_id" value="8888" class="mt-1 block w-full border border-gray-300 rounded-md p-2">
                    </div>
                    <div>
                        <label class="block text-sm font-medium text-gray-700">Project ID (Leave blank for nil)</label>
                        <input type="number" id="project_id" value="7777" class="mt-1 block w-full border border-gray-300 rounded-md p-2">
                    </div>
                    <div>
                        <label class="block text-sm font-medium text-gray-700">Category ID (Leave blank for nil)</label>
                        <input type="number" id="category_id" value="6666" class="mt-1 block w-full border border-gray-300 rounded-md p-2">
                    </div>
                    <button onclick="runValidation()" class="w-full bg-blue-600 text-white py-2 rounded hover:bg-blue-700 font-bold">Run Validation</button>
                </div>
            </div>

            <!-- Cases Section -->
            <div>
                <h2 class="text-lg font-semibold mb-2">Pre-defined Cases</h2>
                <div class="space-y-2">
                    <button onclick="setCase(9999, 8888, 7777, 6666)" class="w-full text-left p-3 border rounded hover:bg-gray-50">
                        <div class="font-semibold">Case 1: Valid Hierarchy</div>
                        <div class="text-xs text-gray-500">All IDs match correctly</div>
                    </button>
                    <button onclick="setCase(9999, 1234, 7777, 6666)" class="w-full text-left p-3 border rounded hover:bg-gray-50">
                        <div class="font-semibold">Case 2: Invalid Environment</div>
                        <div class="text-xs text-gray-500">Wrong Environment ID</div>
                    </button>
                    <button onclick="setCase(9999, 8888, null, 6666)" class="w-full text-left p-3 border rounded hover:bg-gray-50">
                        <div class="font-semibold">Case 3: Orphan Category</div>
                        <div class="text-xs text-gray-500">Has Category but no Project</div>
                    </button>
                    <button onclick="setCase(9999, 8888, 5555, null)" class="w-full text-left p-3 border rounded hover:bg-gray-50">
                        <div class="font-semibold">Case 4: Project Mismatch</div>
                        <div class="text-xs text-gray-500">Project does not belong to Product</div>
                    </button>
                    <button onclick="setCase(9999, 8888, 7777, 1111)" class="w-full text-left p-3 border rounded hover:bg-gray-50">
                        <div class="font-semibold">Case 5: Category Mismatch</div>
                        <div class="text-xs text-gray-500">Category does not belong to Project</div>
                    </button>
                    <button onclick="setCase(9999, 8888, 7778, null)" class="w-full text-left p-3 border rounded hover:bg-gray-50">
                        <div class="font-semibold">Case 6: Inactive Project</div>
                        <div class="text-xs text-gray-500">Project exists but is_active = false</div>
                    </button>
                </div>
            </div>
        </div>

        <div class="mt-8">
            <h2 class="text-lg font-semibold mb-2">Result</h2>
            <div id="result" class="p-4 border rounded bg-gray-50 min-h-[100px] whitespace-pre-wrap font-mono text-sm">Waiting for input...</div>
        </div>
    </div>

    <script>
        function setCase(prod, env, proj, cat) {
            document.getElementById('product_id').value = prod !== null ? prod : '';
            document.getElementById('environment_id').value = env !== null ? env : '';
            document.getElementById('project_id').value = proj !== null ? proj : '';
            document.getElementById('category_id').value = cat !== null ? cat : '';
        }

        async function runValidation() {
            const resultBox = document.getElementById('result');
            resultBox.innerHTML = '<span class="text-gray-500">Running validation...</span>';

            const payload = {
                product_id: parseInt(document.getElementById('product_id').value) || null,
                environment_id: parseInt(document.getElementById('environment_id').value) || null,
                project_id: document.getElementById('project_id').value ? parseInt(document.getElementById('project_id').value) : null,
                category_id: document.getElementById('category_id').value ? parseInt(document.getElementById('category_id').value) : null
            };

            try {
                const res = await fetch('/validate', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(payload)
                });
                const data = await res.json();
                
                if (data.success) {
                    resultBox.innerHTML = '<span class="text-green-600 font-bold">SUCCESS</span>\n\nValidation passed perfectly!';
                } else {
                    resultBox.innerHTML = '<span class="text-red-600 font-bold">FAILED</span>\n\nError: ' + data.error;
                }
            } catch (err) {
                resultBox.innerHTML = '<span class="text-red-600 font-bold">ERROR</span>\n\n' + err.message;
            }
        }
    </script>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

type ValidateRequest struct {
	ProductID     int  `json:"product_id"`
	EnvironmentID int  `json:"environment_id"`
	ProjectID     *int `json:"project_id"`
	CategoryID    *int `json:"category_id"`
}

func validateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx := db.Begin()
	defer tx.Rollback()

	// 1. Setup Mock Data inside transaction (same as CLI)
	setupMockData(tx)

	// 2. Run Validation
	err := validateLogHierarchy(tx, req.ProductID, req.EnvironmentID, req.ProjectID, req.CategoryID)

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]any{"success": false, "error": err.Error()})
	} else {
		json.NewEncoder(w).Encode(map[string]any{"success": true})
	}
}

func setupMockData(tx *gorm.DB) {
	tx.Create(&models.Product{ProductID: 9999, ProductName: "Mock Product", ProductCode: "mock-product", IsActive: true})
	tx.Create(&models.ProductEnvironment{EnvironmentID: 8888, ProductID: 9999, EnvironmentCode: "mock-env", EnvironmentName: "Mock Environment"})
	tx.Create(&models.Project{ProjectID: 7777, ProductID: 9999, ProjectCode: "mock-proj", ProjectName: "Mock Project", IsActive: true})

	pInactive := models.Project{ProjectID: 7778, ProductID: 9999, ProjectCode: "mock-proj-inactive", ProjectName: "Mock Project Inactive", IsActive: false}
	tx.Create(&pInactive)
	tx.Model(&pInactive).Update("is_active", false)

	tx.Create(&models.ProjectFeature{CategoryID: 6666, ProductID: 9999, ProjectID: 7777, CategoryCode: "mock-cat", CategoryName: "Mock Category", IsActive: true})
}

func validateLogHierarchy(
	tx *gorm.DB,
	productID int,
	environmentID int,
	projectID *int,
	categoryID *int,
) error {
	var environmentCount int64
	if err := tx.Model(&models.ProductEnvironment{}).Where("environment_id = ? AND product_id = ? AND deleted_at IS NULL", environmentID, productID).Count(&environmentCount).Error; err != nil {
		return fmt.Errorf("validate environment: %w", err)
	}
	if environmentCount == 0 {
		return errors.New("environment does not belong to product")
	}

	if projectID == nil && categoryID != nil {
		return errors.New("project_id is required when category_id is provided")
	}

	if projectID != nil {
		var projectCount int64
		if err := tx.Model(&models.Project{}).Where("project_id = ? AND product_id = ? AND is_active = TRUE", *projectID, productID).Count(&projectCount).Error; err != nil {
			return fmt.Errorf("validate project: %w", err)
		}
		if projectCount == 0 {
			return errors.New("project does not belong to product")
		}
	}

	if categoryID != nil {
		var categoryCount int64
		if err := tx.Model(&models.ProjectFeature{}).Where("category_id = ? AND project_id = ? AND product_id = ? AND is_active = TRUE", *categoryID, *projectID, productID).Count(&categoryCount).Error; err != nil {
			return fmt.Errorf("validate category: %w", err)
		}
		if categoryCount == 0 {
			return errors.New("category does not belong to the specified product and project")
		}
	}

	return nil
}
