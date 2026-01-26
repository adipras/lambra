package router

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/yourusername/lambra/internal/api/handlers"
	"github.com/yourusername/lambra/internal/api/middleware"
	"github.com/yourusername/lambra/internal/repository"
	"github.com/yourusername/lambra/internal/service"
)

func Setup(db *sqlx.DB) *gin.Engine {
	// Get workspace path from environment
	workspacePath := os.Getenv("WORKSPACE_PATH")
	if workspacePath == "" {
		workspacePath = "/workspace"
	}
	router := gin.New()

	// Middleware
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())

	// Initialize repositories
	projectRepo := repository.NewProjectRepository(db)
	entityRepo := repository.NewEntityRepository(db)
	endpointRepo := repository.NewEndpointRepository(db)
	relationRepo := repository.NewRelationRepository(db)
	snapshotRepo := repository.NewSnapshotRepository(db)
	deploymentRepo := repository.NewDeploymentRepository(db)
	deploymentLogRepo := repository.NewDeploymentLogRepository(db)

	// Initialize services
	projectService := service.NewProjectService(projectRepo)
	entityService := service.NewEntityService(entityRepo, projectRepo, endpointRepo)
	endpointService := service.NewEndpointService(endpointRepo, entityRepo, projectRepo)
	relationService := service.NewRelationService(relationRepo, entityRepo)
	generatorService := service.NewGeneratorService(projectRepo, entityRepo, endpointRepo, relationRepo)
	deploymentService := service.NewDeploymentService(projectRepo, entityRepo, endpointRepo, relationRepo, snapshotRepo, deploymentRepo, deploymentLogRepo, generatorService, workspacePath)
	exportService := service.NewExportService(projectRepo, entityRepo, endpointRepo)
	snapshotService := service.NewSnapshotService(snapshotRepo, projectRepo, entityRepo, endpointRepo, relationRepo)

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(db)
	projectHandler := handlers.NewProjectHandler(projectService)
	entityHandler := handlers.NewEntityHandler(entityService)
	endpointHandler := handlers.NewEndpointHandler(endpointService, deploymentService)
	relationHandler := handlers.NewRelationHandler(relationService)
	generatorHandler := handlers.NewGeneratorHandler(generatorService)
	deploymentHandler := handlers.NewDeploymentHandler(deploymentService)
	exportHandler := handlers.NewExportHandler(exportService)
	snapshotHandler := handlers.NewSnapshotHandler(snapshotService, deploymentService)

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

			// Nested routes for project entities and endpoints
			projects.POST("/:id/entities", entityHandler.CreateEntity)
			projects.GET("/:id/entities", entityHandler.GetEntitiesByProject)
			projects.GET("/:id/endpoints", endpointHandler.GetEndpointsByProject)

			// Deployment routes
			projects.POST("/:id/deploy", deploymentHandler.DeployProject)
			projects.POST("/:id/start", deploymentHandler.StartService)
			projects.POST("/:id/stop", deploymentHandler.StopService)
			projects.POST("/:id/redeploy", deploymentHandler.RedeployService)
			projects.DELETE("/:id/destroy", deploymentHandler.DestroyService)
			projects.GET("/:id/status", deploymentHandler.GetServiceStatus)

			// Export routes
			projects.GET("/:id/export/openapi", exportHandler.ExportOpenAPI)
			projects.GET("/:id/export/postman", exportHandler.ExportPostman)

			// Snapshot routes (under projects)
			projects.GET("/:id/snapshots", snapshotHandler.ListByProject)
			projects.POST("/:id/snapshots", snapshotHandler.Create)
		}

		// Entities
		entities := v1.Group("/entities")
		{
			entities.GET("/:id", entityHandler.GetEntity)
			entities.PUT("/:id", entityHandler.UpdateEntity)
			entities.DELETE("/:id", entityHandler.DeleteEntity)
			entities.GET("/:id/endpoints", endpointHandler.GetEndpointsByEntity)
			entities.GET("/:id/relations", relationHandler.GetEntityRelations)
		}

		// Endpoints
		endpoints := v1.Group("/endpoints")
		{
			endpoints.POST("", endpointHandler.CreateEndpoint)
			endpoints.GET("/:id", endpointHandler.GetEndpoint)
			endpoints.PUT("/:id", endpointHandler.UpdateEndpoint)
			endpoints.DELETE("/:id", endpointHandler.DeleteEndpoint)
			endpoints.POST("/:id/test", endpointHandler.TestEndpoint)
		}

		// Relations
		relations := v1.Group("/relations")
		{
			relations.POST("", relationHandler.CreateRelation)
			relations.GET("/:id", relationHandler.GetRelation)
			relations.PUT("/:id", relationHandler.UpdateRelation)
			relations.DELETE("/:id", relationHandler.DeleteRelation)
		}

		// Code Generation
		generate := v1.Group("/generate")
		{
			generate.POST("/entity", generatorHandler.GenerateEntity)
			generate.POST("/project", generatorHandler.GenerateProject)
			generate.GET("/preview/:id", generatorHandler.PreviewEntity)
			generate.GET("/files/:id", generatorHandler.GetGeneratedFilesList)
		}

		// Snapshots
		snapshots := v1.Group("/snapshots")
		{
			snapshots.GET("/:id", snapshotHandler.Get)
			snapshots.GET("/:id/metadata", snapshotHandler.GetMetadata)
			snapshots.POST("/:id/rollback", snapshotHandler.Rollback)
			snapshots.DELETE("/:id", snapshotHandler.Delete)
		}

		// Deployments
		deployments := v1.Group("/deployments")
		{
			deployments.GET("/:id", deploymentHandler.GetDeployment)
			deployments.GET("/:id/logs", deploymentHandler.GetDeploymentLogs)
			deployments.GET("/:id/logs/stream", deploymentHandler.StreamDeploymentLogs)
		}

		// Project deployments
		projects.GET("/:id/deployments", deploymentHandler.GetProjectDeployments)
		projects.GET("/:id/deployments/latest", deploymentHandler.GetLatestDeployment)
		projects.GET("/:id/container-logs", deploymentHandler.GetContainerLogs)
		projects.GET("/:id/container-logs/stream", deploymentHandler.StreamContainerLogs)
	}

	return router
}
