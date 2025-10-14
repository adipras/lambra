package router

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/yourusername/lambra/internal/api/handlers"
	"github.com/yourusername/lambra/internal/api/middleware"
	"github.com/yourusername/lambra/internal/repository"
	"github.com/yourusername/lambra/internal/service"
)

func Setup(db *sqlx.DB) *gin.Engine {
	router := gin.New()

	// Middleware
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())

	// Initialize repositories
	projectRepo := repository.NewProjectRepository(db)
	entityRepo := repository.NewEntityRepository(db)
	endpointRepo := repository.NewEndpointRepository(db)

	// Initialize services
	projectService := service.NewProjectService(projectRepo)

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(db)
	projectHandler := handlers.NewProjectHandler(projectService)

	// Health check routes
	router.GET("/health", healthHandler.HealthCheck)
	router.GET("/ready", healthHandler.Readiness)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Projects
		projects := v1.Group("/projects")
		{
			projects.POST("", projectHandler.CreateProject)
			projects.GET("", projectHandler.GetAllProjects)
			projects.GET("/:id", projectHandler.GetProject)
			projects.PUT("/:id", projectHandler.UpdateProject)
			projects.DELETE("/:id", projectHandler.DeleteProject)

			// Entities (will be implemented in Phase 2)
			// projects.POST("/:id/entities", entityHandler.CreateEntity)
			// projects.GET("/:id/entities", entityHandler.GetEntitiesByProject)

			// Generation (will be implemented in Phase 2)
			// projects.POST("/:id/generate", generatorHandler.Generate)
			// projects.POST("/:id/regenerate", generatorHandler.Regenerate)
		}

		// Entities (will be implemented later)
		// entities := v1.Group("/entities")
		// {
		// 	entities.GET("/:id", entityHandler.GetEntity)
		// 	entities.PUT("/:id", entityHandler.UpdateEntity)
		// 	entities.DELETE("/:id", entityHandler.DeleteEntity)
		// }

		// Endpoints (will be implemented later)
		// endpoints := v1.Group("/endpoints")
		// {
		// 	endpoints.POST("", endpointHandler.CreateEndpoint)
		// 	endpoints.GET("/:id", endpointHandler.GetEndpoint)
		// 	endpoints.PUT("/:id", endpointHandler.UpdateEndpoint)
		// 	endpoints.DELETE("/:id", endpointHandler.DeleteEndpoint)
		// 	endpoints.POST("/:id/test", endpointHandler.TestEndpoint)
		// }

		// Deployments (will be implemented in Phase 4)
		// deployments := v1.Group("/deployments")
		// {
		// 	deployments.GET("", deploymentHandler.GetDeployments)
		// 	deployments.GET("/:id", deploymentHandler.GetDeployment)
		// 	deployments.GET("/:id/logs", deploymentHandler.GetDeploymentLogs)
		// }

		// Snapshots (will be implemented in Phase 4)
		// snapshots := v1.Group("/snapshots")
		// {
		// 	snapshots.GET("/project/:project_id", snapshotHandler.GetSnapshots)
		// 	snapshots.POST("/:id/rollback", snapshotHandler.Rollback)
		// }
	}

	// Prevent unused variable warnings (remove these when implementing)
	_ = entityRepo
	_ = endpointRepo

	return router
}
