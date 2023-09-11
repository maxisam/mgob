package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	log "github.com/sirupsen/logrus"

	"github.com/stefanprodan/mgob/pkg/backup"
	"github.com/stefanprodan/mgob/pkg/config"
	"github.com/stefanprodan/mgob/pkg/notifier"
)

func postBackup(w http.ResponseWriter, r *http.Request) {
	cfg := r.Context().Value("app.config").(config.AppConfig)
	modules := r.Context().Value("app.modules").(config.ModuleConfig)
	planID := chi.URLParam(r, "planID")
	plan, err := config.LoadPlan(cfg.ConfigPath, planID)
	if err != nil {
		render.Status(r, 500)
		render.JSON(w, r, map[string]string{"error": err.Error()})
		return
	}

	log.WithField("plan", planID).Info("On demand backup started")

	res, err := backup.Run(plan, &cfg, &modules)
	if err != nil {
		log.WithField("plan", planID).Errorf("On demand backup failed %v", err)
		if err := notifier.SendNotification(fmt.Sprintf("BACKUP FAILED: %v on demand backup failed", planID),
			err.Error(), true, plan); err != nil {
			log.WithField("plan", plan.Name).Errorf("Notifier failed for on demand backup %v", err)
		}
		render.Status(r, 500)
		render.JSON(w, r, map[string]string{"error": err.Error()})
	} else {
		log.WithField("plan", plan.Name).Infof("On demand backup finished in %v archive %v size %v",
			res.Duration, res.Name, humanize.Bytes(uint64(res.Size)))
		if err := notifier.SendNotification(fmt.Sprintf("%v on demand backup finished", plan.Name),
			fmt.Sprintf("%v backup finished in %v archive size %v",
				res.Name, res.Duration, humanize.Bytes(uint64(res.Size))),
			false, plan); err != nil {
			log.WithField("plan", plan.Name).Errorf("Notifier failed for on demand backup %v", err)
		}
		render.JSON(w, r, toBackupResult(res))
	}
}

type backupResult struct {
	Plan      string    `json:"plan"`
	File      string    `json:"file"`
	Duration  string    `json:"duration"`
	Size      string    `json:"size"`
	Timestamp time.Time `json:"timestamp"`
}

func toBackupResult(res backup.Result) backupResult {
	return backupResult{
		Plan:      res.Plan,
		Duration:  fmt.Sprintf("%v", res.Duration),
		File:      res.Name,
		Size:      humanize.Bytes(uint64(res.Size)),
		Timestamp: res.Timestamp,
	}
}
