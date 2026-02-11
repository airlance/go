package http

import (
	"math"
	"net/http"
	"strconv"

	"git.emercury.dev/emercury/senderscore/api/internal/domain"
	"git.emercury.dev/emercury/senderscore/api/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type GroupHandler struct {
	groupUC usecase.GroupUseCase
	ipUC    usecase.IPUseCase
}

func NewGroupHandler(groupUC usecase.GroupUseCase, ipUC usecase.IPUseCase) *GroupHandler {
	return &GroupHandler{
		groupUC: groupUC,
		ipUC:    ipUC,
	}
}

func (h *GroupHandler) ListGroups(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	withIPs := c.DefaultQuery("with_ips", "false") == "true"

	pagination := usecase.PaginationDTO{
		Page:     page,
		PageSize: pageSize,
	}

	groups, total, err := h.groupUC.ListGroups(c.Request.Context(), pagination, withIPs)
	if err != nil {
		logrus.WithError(err).Error("Failed to list groups")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve groups",
		})
		return
	}

	responses := toGroupResponses(groups)
	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	c.JSON(http.StatusOK, PaginatedResponse{
		Data:       responses,
		Page:       page,
		PageSize:   pageSize,
		TotalItems: total,
		TotalPages: totalPages,
	})
}

func (h *GroupHandler) GetGroup(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid group ID format",
		})
		return
	}

	withIPs := c.DefaultQuery("with_ips", "true") == "true"

	group, err := h.groupUC.GetGroupByID(c.Request.Context(), uint(id), withIPs)
	if err != nil {
		if err == domain.ErrGroupNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "not_found",
				Message: "Group not found",
			})
			return
		}
		logrus.WithError(err).Error("Failed to get group")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve group",
		})
		return
	}

	response := toGroupResponse(group)
	c.JSON(http.StatusOK, response)
}

func (h *GroupHandler) GetGroupByGroupID(c *gin.Context) {
	groupIDParam := c.Param("group_id")
	groupID, err := strconv.Atoi(groupIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_group_id",
			Message: "Invalid group_id format",
		})
		return
	}

	withIPs := c.DefaultQuery("with_ips", "true") == "true"

	group, err := h.groupUC.GetGroupByGroupID(c.Request.Context(), groupID, withIPs)
	if err != nil {
		if err == domain.ErrGroupNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "not_found",
				Message: "Group not found",
			})
			return
		}
		logrus.WithError(err).Error("Failed to get group")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve group",
		})
		return
	}

	response := toGroupResponse(group)
	c.JSON(http.StatusOK, response)
}

func (h *GroupHandler) CreateGroup(c *gin.Context) {
	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	dto := toCreateGroupDTO(req)
	group, err := h.groupUC.CreateGroup(c.Request.Context(), dto)
	if err != nil {
		if err == domain.ErrGroupAlreadyExists {
			c.JSON(http.StatusConflict, ErrorResponse{
				Error:   "group_exists",
				Message: "Group with this group_id already exists",
			})
			return
		}
		logrus.WithError(err).Error("Failed to create group")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to create group",
		})
		return
	}

	response := toGroupResponse(group)
	c.JSON(http.StatusCreated, response)
}

func (h *GroupHandler) UpdateGroupCounters(c *gin.Context) {
	groupIDParam := c.Param("group_id")
	groupID, err := strconv.Atoi(groupIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_group_id",
			Message: "Invalid group_id format",
		})
		return
	}

	if err := h.groupUC.UpdateCounters(c.Request.Context(), groupID); err != nil {
		logrus.WithError(err).Error("Failed to update group counters")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to update counters",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Counters updated successfully",
	})
}

func (h *GroupHandler) DeleteGroup(c *gin.Context) {
	groupIDParam := c.Param("group_id")
	groupID, err := strconv.Atoi(groupIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_group_id",
			Message: "Invalid group_id format",
		})
		return
	}

	if err := h.groupUC.DeleteGroup(c.Request.Context(), groupID); err != nil {
		if err == domain.ErrGroupNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "not_found",
				Message: "Group not found",
			})
			return
		}
		logrus.WithError(err).Error("Failed to delete group")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to delete group",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Group deleted successfully",
	})
}

func (h *GroupHandler) AddIP(c *gin.Context) {
	var req AddIPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	dto := toAddIPDTO(req)
	ip, err := h.ipUC.AddIP(c.Request.Context(), dto)
	if err != nil {
		logrus.WithError(err).Error("Failed to add IP")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to add IP address",
		})
		return
	}

	response := toIPResponse(ip)
	c.JSON(http.StatusCreated, response)
}

func (h *GroupHandler) AddIPs(c *gin.Context) {
	var req AddIPsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	dtos := toAddIPDTOs(req)
	result, err := h.ipUC.AddIPs(c.Request.Context(), dtos)
	if err != nil {
		logrus.WithError(err).Error("Failed to add IPs in batch")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to add IP addresses",
		})
		return
	}

	response := toAddIPsResponse(result)
	c.JSON(http.StatusCreated, response)
}

func (h *GroupHandler) SubmitScore(c *gin.Context) {
	var req SubmitScoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	dto := toSubmitScoreDTO(req)
	result, err := h.ipUC.SubmitScore(c.Request.Context(), dto)
	if err != nil {
		logrus.WithError(err).Error("Failed to submit score")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to process score submission",
		})
		return
	}

	response := toSubmitScoreResponse(result)
	c.JSON(http.StatusCreated, response)
}

func (h *GroupHandler) GetOldestIP(c *gin.Context) {
	ip, err := h.ipUC.GetOldestIP(c.Request.Context())
	if err != nil {
		if err == domain.ErrIPNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "not_found",
				Message: "No IP addresses found",
			})
			return
		}
		logrus.WithError(err).Error("Failed to get oldest IP")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve IP",
		})
		return
	}

	response := toIPResponse(ip)
	c.JSON(http.StatusOK, response)
}
