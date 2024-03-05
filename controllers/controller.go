package controllers

import (
	"os"

	"false.kr/WebChecker-Node/checker"
	"false.kr/WebChecker-Node/dto"
	"false.kr/WebChecker-Node/files"
	"github.com/gofiber/fiber/v2"
)

func API(c *fiber.Ctx) error {
	requestData := new(dto.RequestCheckDTO)
	if err := c.BodyParser(requestData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ResponseDTO{
			Status:  fiber.StatusBadRequest,
			Message: "Invalid JSON",
		})
	}

	responseData := checker.CheckPage(requestData, c.BaseURL())

	return c.Status(fiber.StatusOK).JSON(dto.ResponseDTO{
		Status:       fiber.StatusOK,
		Message:      "Request OK, Check ResponseData",
		ResponseData: responseData,
	})
}

func GetImage(c *fiber.Ctx) error {
	id := c.Params("targetId")

	if _, err := os.Stat(files.Config.Screenshot + id + "_screenshot.png"); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ResponseDTO{
			Status:  fiber.StatusBadRequest,
			Message: "Image Doesn't Exist",
		})
	}
	return c.SendFile(files.Config.Screenshot + id + "_screenshot.png")
}
