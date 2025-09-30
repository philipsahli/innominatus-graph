package api

import (
	"net/http"

	"idp-orchestrator/pkg/export"
	"idp-orchestrator/pkg/storage"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RESTHandler struct {
	repository storage.RepositoryInterface
	exporter   *export.Exporter
}

func NewRESTHandler(repository storage.RepositoryInterface) *RESTHandler {
	return &RESTHandler{
		repository: repository,
		exporter:   export.NewExporter(),
	}
}

func (h *RESTHandler) Close() error {
	return h.exporter.Close()
}

func (h *RESTHandler) SetupRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	{
		api.GET("/graph", h.GetGraph)
		api.POST("/graph/export", h.ExportGraph)
		api.GET("/apps/:app/runs", h.GetGraphRuns)
		api.POST("/apps/:app/runs", h.CreateGraphRun)
		api.PUT("/runs/:runId", h.UpdateGraphRun)
	}
}

type GetGraphResponse struct {
	Graph   interface{} `json:"graph"`
	AppName string      `json:"app_name"`
	Version int         `json:"version,omitempty"`
}

func (h *RESTHandler) GetGraph(c *gin.Context) {
	appName := c.Query("app")
	if appName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "app parameter is required"})
		return
	}

	graph, err := h.repository.LoadGraph(appName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Graph not found: " + err.Error()})
		return
	}

	response := GetGraphResponse{
		Graph:   graph,
		AppName: appName,
		Version: graph.Version,
	}

	c.JSON(http.StatusOK, response)
}

type ExportRequest struct {
	Format  string   `json:"format" form:"format"`
	NodeIDs []string `json:"node_ids,omitempty" form:"node_ids"`
}

func (h *RESTHandler) ExportGraph(c *gin.Context) {
	appName := c.Query("app")
	if appName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "app parameter is required"})
		return
	}

	var req ExportRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	if req.Format == "" {
		req.Format = "dot"
	}

	graph, err := h.repository.LoadGraph(appName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Graph not found: " + err.Error()})
		return
	}

	exportGraph := graph
	if len(req.NodeIDs) > 0 {
		exportGraph, err = h.exporter.CreateSubgraph(graph, req.NodeIDs)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create subgraph: " + err.Error()})
			return
		}
	}

	var format export.Format
	var contentType string
	var fileExtension string

	switch req.Format {
	case "dot":
		format = export.FormatDOT
		contentType = "text/plain"
		fileExtension = "dot"
	case "svg":
		format = export.FormatSVG
		contentType = "image/svg+xml"
		fileExtension = "svg"
	case "png":
		format = export.FormatPNG
		contentType = "image/png"
		fileExtension = "png"
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported format. Use dot, svg, or png"})
		return
	}

	data, err := h.exporter.ExportGraph(exportGraph, format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export graph: " + err.Error()})
		return
	}

	filename := appName + "-graph." + fileExtension
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, contentType, data)
}

func (h *RESTHandler) GetGraphRuns(c *gin.Context) {
	appName := c.Param("app")

	runs, err := h.repository.GetGraphRuns(appName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get graph runs: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"runs": runs})
}

type CreateGraphRunRequest struct {
	Version int `json:"version" binding:"required"`
}

func (h *RESTHandler) CreateGraphRun(c *gin.Context) {
	appName := c.Param("app")

	var req CreateGraphRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	run, err := h.repository.CreateGraphRun(appName, req.Version)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create graph run: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, run)
}

type UpdateGraphRunRequest struct {
	Status       string  `json:"status" binding:"required"`
	ErrorMessage *string `json:"error_message,omitempty"`
}

func (h *RESTHandler) UpdateGraphRun(c *gin.Context) {
	runIDStr := c.Param("runId")

	var req UpdateGraphRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	runID, err := parseUUID(runIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid run ID"})
		return
	}

	err = h.repository.UpdateGraphRun(runID, req.Status, req.ErrorMessage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update graph run: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Graph run updated successfully"})
}

func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}