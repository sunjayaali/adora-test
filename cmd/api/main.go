package main

import (
	"context"
	"database/sql"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/gofiber/fiber/v3"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/samber/lo"

	"adora-test/ent"
	"adora-test/internal/data"
	"adora-test/internal/repositories"
	"adora-test/internal/service"
	"adora-test/internal/service/logging"
	mockcarrier "adora-test/internal/service/mock_carrier"
)

func main() {
	var devMode bool
	flag.BoolVar(&devMode, "dev", false, "Run in development mode")
	flag.Parse()

	if devMode {
		_ = godotenv.Load()
	}

	db := lo.Must(sql.Open("pgx", os.Getenv("DB_DSN")))
	client := ent.NewClient(ent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	defer client.Close()

	lo.Must0(client.Schema.Create(context.Background()))
	entManager := data.NewEntManager(client)

	entitlementRepo := repositories.NewEntitlement(entManager)
	storeEventRepo := repositories.NewStoreEventRepository(entManager)

	webhookService := service.NewWebhookStoreService(storeEventRepo, entitlementRepo, entManager)
	entitlementService := service.NewEntitlementService(entitlementRepo)

	ts := lo.Must(mockcarrier.NewServer())
	defer ts.Close()

	carrierClient := service.NewCarrierClient(ts.Client(), ts.URL)
	var poller service.Poller
	{
		poller = service.NewPoll(carrierClient, entitlementRepo, entManager)
		poller = logging.NewPoll(poller)
	}

	pollWorker := service.NewPollWorker(entManager, entitlementRepo, poller)

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

	app.Get("/users/:id/entitlement", func(c fiber.Ctx) error {
		userID := c.Params("id")

		resp, err := entitlementService.GetEntitlement(c.Context(), &service.GetEntitlementRequest{UserID: userID})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		jsonResp := struct {
			Active        bool   `json:"active"`
			Source        string `json:"source"`
			ExpiresAt     string `json:"expiresAt"`
			LastChangedAt string `json:"lastChangedAt"`
			Reason        string `json:"reason"`
		}{
			Active:        resp.Entitlement.IsActive,
			Source:        string(resp.Entitlement.Source),
			ExpiresAt:     resp.Entitlement.ExpiresAt.UTC().Format(time.RFC3339),
			LastChangedAt: resp.Entitlement.LastChangedAt.UTC().Format(time.RFC3339),
			Reason:        resp.Entitlement.Reason,
		}
		return c.JSON(jsonResp)
	})

	app.Post("/webhooks/marketplace/revoke", func(c fiber.Ctx) error {
		var req struct {
			UserIDs []string `json:"userIds"`
		}
		if err := c.Bind().Body(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
		}

		resp, err := service.NewRevoke(entitlementRepo, entManager).RevokeEntitlement(c.Context(), &service.RevokeEntitlementRequest{
			UserIDs: req.UserIDs,
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(resp)
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		lo.Must0(app.Listen(":3000"))
	}()

	go func() {
		StartScheduler(ctx, pollWorker)
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(shutdown)

	<-shutdown
	cancel()

	lo.Must0(app.ShutdownWithTimeout(30 * time.Second))
}

func StartScheduler(ctx context.Context, worker *service.PollWorker) {
	interval := 5 * time.Minute
	// Use a shorter interval for testing purposes
	// interval := 10 * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		_ = worker.Run(ctx)
		<-ticker.C
	}
}
