package main

import (
	"context"
	"database/sql"
	"log"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/gofiber/fiber/v3"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/samber/lo"

	"adora-test/ent"
	"adora-test/internal/data"
	"adora-test/internal/repositories"
	"adora-test/internal/service"
)

func main() {
	db := lo.Must(sql.Open("pgx", "postgres://postgres:postgres@localhost:5432/adora?sslmode=disable"))
	client := ent.NewClient(ent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	defer client.Close()

	lo.Must0(client.Schema.Create(context.Background()))
	entManager := data.NewEntManager(client)

	entitlementRepo := repositories.NewEntitlement(entManager)
	storeEventRepo := repositories.NewStoreEventRepository(entManager)

	webhookService := service.NewWebhookStoreService(storeEventRepo, entitlementRepo, entManager)

	app := fiber.New()

	app.Post("/webhooks/store", func(c fiber.Ctx) error {
		var req struct {
			EventID     string `json:"eventId"`
			UserID      string `json:"userId"`
			Type        string `json:"type"`
			EventTimeMs int64  `json:"eventTimeMs"`
			ProductID   string `json:"productId"`
		}
		if err := c.Bind().Body(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
		}

		resp, err := webhookService.Ingest(c.Context(), &service.IngestStoreWebhookRequest{
			EventID:     req.EventID,
			UserID:      req.UserID,
			Type:        req.Type,
			EventTimeMs: req.EventTimeMs,
			ProductID:   req.ProductID,
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(resp)
	})

	log.Fatal(app.Listen(":3000"))
}
