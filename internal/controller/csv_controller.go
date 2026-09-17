package controller

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"tensechassignment/internal/middleware"
	"tensechassignment/internal/service"
	"tensechassignment/internal/utils"
)

type CSVController struct {
	service *service.CSVService
}

func NewCSVController(service *service.CSVService) *CSVController {
	return &CSVController{
		service: service,
	}
}

func (cc *CSVController) UploadCSV(c *gin.Context) {
	tenantID := middleware.TenantID(c)

	fileHeader, err := c.FormFile("file")
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "csv file is required")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		log.Println("failed to open uploaded file:", err)
		utils.Error(c, http.StatusInternalServerError, "failed to open uploaded file")
		return
	}
	defer file.Close()

	result, err := cc.service.ProcessCSV(
		c.Request.Context(),
		tenantID,
		file,
	)
	if err != nil {
		log.Println("csv processing failed:", err)
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.Success(c, http.StatusOK, result)
}