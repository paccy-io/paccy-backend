package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/paccy-io/paccy-backend/internal/notify"
	"github.com/paccy-io/paccy-backend/internal/notify/dto"
	"github.com/paccy-io/paccy-backend/pkg/logger"
	responses "github.com/paccy-io/paccy-backend/pkg/response"
)

type NotificationHandler struct {
	notifyService notify.NotifyService
	quota         dto.QuotaDTO
}

func NewNotificationHandler(notifyService notify.NotifyService, quota dto.QuotaDTO) *NotificationHandler {
	return &NotificationHandler{
		notifyService: notifyService,
		quota:         quota,
	}
}

func (h *NotificationHandler) SendNotification(c *fiber.Ctx) error {
	var user dto.SendNotificationDTO

	err := c.BodyParser(&user)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.APIResponse{
			Success: false,
			Message: "Invalid request",
		})
	}

	err = h.notifyService.SendNotification(c.Context(), &user)
	if err != nil {
		log := logger.Notify()
		log.WithError(err).Error("Failed to send notification")
		return c.Status(fiber.StatusInternalServerError).JSON(responses.APIResponse{
			Success: false,
			Message: "Failed to send notification",
		})
	}
	return c.Status(fiber.StatusOK).JSON(responses.APIResponse{
		Success: true,
		Message: "Notification sent successfully",
	})
}

func (h *NotificationHandler) CreateProjectQuota(c *fiber.Ctx) error {
	err := h.notifyService.CreateProjectQuota(c.Context(), &h.quota)
	if err != nil {
		log := logger.Notify()
		log.WithError(err).Error("Failed to create project quota")
		return c.Status(fiber.StatusInternalServerError).JSON(responses.APIResponse{
			Success: false,
			Message: "Failed to create project quota",
		})
	}
	return c.Status(fiber.StatusOK).JSON(responses.APIResponse{
		Success: true,
		Message: "Project quota created successfully",
	})
}
