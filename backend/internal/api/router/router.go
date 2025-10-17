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
	entityService := service.NewEntityService(entityRepo, projectRepo)
	endpointService := service.NewEndpointService(endpointRepo, entityRepo, projectRepo)
	generatorService := service.NewGeneratorService(projectRepo, entityRepo, endpointRepo)

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(db)
	projectHandler := handlers.NewProjectHandler(projectService)
	entityHandler := handlers.NewEntityHandler(entityService)
	endpointHandler := handlers.NewEndpointHandler(endpointService)
	generatorHandler := handlers.NewGeneratorHandler(generatorService)

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
			projects.GET("/:id/entities", entityHandler.GetEntitiesByProject)
			projects.GET("/:id/endpoints", endpointHandler.GetEndpointsByProject)
			projects.PUT("/:id", projectHandler.UpdateProject)
			projects.DELETE("/:id", projectHandler.DeleteProject)
		}

		// Entities
		entities := v1.Group("/entities")
		{
			entities.POST("", entityHandler.CreateEntity)
			entities.GET("/:id", entityHandler.GetEntity)
			entities.GET("/:id/endpoints", endpointHandler.GetEndpointsByEntity)
			entities.PUT("/:id", entityHandler.UpdateEntity)
			entities.DELETE("/:id", entityHandler.DeleteEntity)
		}

		// Endpoints
		endpoints := v1.Group("/endpoints")
		{
			endpoints.POST("", endpointHandler.CreateEndpoint)
			endpoints.GET("/:id", endpointHandler.GetEndpoint)
			endpoints.PUT("/:id", endpointHandler.UpdateEndpoint)
			endpoints.DELETE("/:id", endpointHandler.DeleteEndpoint)
		}

		// Code Generation
		generate := v1.Group("/generate")
		{
			generate.POST("/entity", generatorHandler.GenerateEntity)
			generate.POST("/project", generatorHandler.GenerateProject)
			generate.GET("/preview/:id", generatorHandler.PreviewEntity)
			generate.GET("/files/:id", generatorHandler.GetGeneratedFilesList)
		}

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

	return router
}
