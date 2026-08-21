package routes

import (
	"net/http"
	"os"
	"path/filepath"

	"omnilogs-api/configs"

	"github.com/gin-gonic/gin"
)

func SwaggerRoutes(router *gin.Engine, env *configs.Env) {
	router.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/swagger/index.html")
	})
	router.GET("/swagger/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/swagger/index.html")
	})

	router.GET("/swagger/index.html", func(c *gin.Context) {
		if !env.SwaggerEnabled {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "SWAGGER_DISABLED",
					"message": "swagger UI is disabled in this environment",
					"status":  http.StatusNotFound,
				},
			})
			return
		}

		path, err := resolveProjectFile("docs", "swagger", "index.html")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "SWAGGER_FILE_MISSING",
					"message": "swagger UI file is missing",
					"status":  http.StatusInternalServerError,
					"details": err.Error(),
				},
			})
			return
		}

		data, err := os.ReadFile(path)
		if err != nil {
			c.String(http.StatusInternalServerError, "failed to read swagger index")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})

	router.GET("/swagger/openapi.yaml", func(c *gin.Context) {
		if !env.SwaggerEnabled {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "SWAGGER_DISABLED",
					"message": "swagger document is disabled in this environment",
					"status":  http.StatusNotFound,
				},
			})
			return
		}

		path, err := resolveProjectFile("docs", "swagger", "openapi.yaml")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "SWAGGER_FILE_MISSING",
					"message": "swagger document file is missing",
					"status":  http.StatusInternalServerError,
					"details": err.Error(),
				},
			})
			return
		}

		data, err := os.ReadFile(path)
		if err != nil {
			c.String(http.StatusInternalServerError, "failed to read openapi spec")
			return
		}
		c.Data(http.StatusOK, "text/yaml; charset=utf-8", data)
	})
}

func resolveProjectFile(parts ...string) (string, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	current := workingDir
	for {
		candidate := filepath.Join(append([]string{current}, parts...)...)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	return "", os.ErrNotExist
}
